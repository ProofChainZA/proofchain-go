// Package proofchain provides a Go client for the ProofChain API.
package proofchain

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ProofChainZA/proofchain-go/proofchain/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	defaultGRPCEndpoint = "grpc.proofchain.co.za:443"
	defaultGRPCTimeout  = 30 * time.Second
)

// GRPCEvent represents an event to be sent via gRPC streaming.
type GRPCEvent struct {
	UserID       string                 `json:"user_id"`
	EventType    string                 `json:"event_type"`
	DocumentHash string                 `json:"document_hash,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Timestamp    *time.Time             `json:"timestamp,omitempty"`
}

// GRPCResponse represents a response from the gRPC stream.
type GRPCResponse struct {
	EventID       string `json:"event_id"`
	CertificateID string `json:"certificate_id"`
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
}

// StreamStats contains statistics about a streaming session.
type StreamStats struct {
	TotalSent     int64
	TotalSuccess  int64
	TotalFailed   int64
	TotalDropped  int64 // Events dropped due to buffer full (only for TrySend)
	Duration      time.Duration
	EventsPerSec  float64
	ActiveStreams int
}

// GRPCClientOption configures the gRPC client.
type GRPCClientOption func(*GRPCClient)

// WithGRPCEndpoint sets a custom gRPC endpoint.
func WithGRPCEndpoint(endpoint string) GRPCClientOption {
	return func(c *GRPCClient) {
		c.endpoint = endpoint
	}
}

// WithGRPCTimeout sets the connection timeout.
func WithGRPCTimeout(timeout time.Duration) GRPCClientOption {
	return func(c *GRPCClient) {
		c.timeout = timeout
	}
}

// WithTLS enables or disables TLS (enabled by default for port 443).
func WithTLS(enabled bool) GRPCClientOption {
	return func(c *GRPCClient) {
		c.useTLS = enabled
	}
}

// WithNumStreams sets the number of parallel streams for multi-stream mode.
func WithNumStreams(n int) GRPCClientOption {
	return func(c *GRPCClient) {
		if n > 0 {
			c.numStreams = n
		}
	}
}

// GRPCClient provides high-performance gRPC streaming for event ingestion.
// Supports single-stream and multi-stream modes for maximum throughput.
//
// Multi-stream mode creates multiple parallel connections to distribute
// load across server pods, achieving 5-10x higher throughput than single-stream.
type GRPCClient struct {
	apiKey     string
	endpoint   string
	timeout    time.Duration
	useTLS     bool
	numStreams int

	mu    sync.RWMutex
	conns []*grpc.ClientConn
}

// NewGRPCClient creates a new gRPC streaming client.
//
// Example:
//
//	client := proofchain.NewGRPCClient("your-api-key")
//	defer client.Close()
//
//	stats, err := client.StreamEvents(ctx, events)
func NewGRPCClient(apiKey string, opts ...GRPCClientOption) *GRPCClient {
	c := &GRPCClient{
		apiKey:     apiKey,
		endpoint:   defaultGRPCEndpoint,
		timeout:    defaultGRPCTimeout,
		useTLS:     true,
		numStreams: 1,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Auto-detect TLS based on port
	if !strings.Contains(c.endpoint, ":443") && c.useTLS {
		c.useTLS = false
	}

	return c
}

// Connect establishes gRPC connections. Call this before streaming.
// For multi-stream mode, this creates multiple connections.
func (c *GRPCClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close existing connections
	for _, conn := range c.conns {
		if conn != nil {
			conn.Close()
		}
	}

	c.conns = make([]*grpc.ClientConn, c.numStreams)

	for i := 0; i < c.numStreams; i++ {
		conn, err := c.dialEndpoint(ctx, c.endpoint)
		if err != nil {
			// Close already established connections
			for j := 0; j < i; j++ {
				if c.conns[j] != nil {
					c.conns[j].Close()
				}
			}
			return fmt.Errorf("failed to connect stream %d: %w", i, err)
		}
		c.conns[i] = conn
	}

	return nil
}

// Close closes all gRPC connections.
func (c *GRPCClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var lastErr error
	for _, conn := range c.conns {
		if conn != nil {
			if err := conn.Close(); err != nil {
				lastErr = err
			}
		}
	}
	c.conns = nil
	return lastErr
}

func (c *GRPCClient) dialEndpoint(ctx context.Context, endpoint string) (*grpc.ClientConn, error) {
	var creds grpc.DialOption
	if c.useTLS {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	} else {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	dialCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return grpc.DialContext(dialCtx, endpoint,
		creds,
		grpc.WithBlock(),
	)
}

// StreamEvents streams events using bidirectional gRPC streaming.
// In multi-stream mode, events are distributed across parallel streams.
//
// Example:
//
//	events := make(chan *proofchain.GRPCEvent, 1000)
//	go func() {
//	    for i := 0; i < 10000; i++ {
//	        events <- &proofchain.GRPCEvent{
//	            UserID:    fmt.Sprintf("user-%d", i),
//	            EventType: "action",
//	            Data:      map[string]interface{}{"index": i},
//	        }
//	    }
//	    close(events)
//	}()
//
//	stats, err := client.StreamEvents(ctx, events)
func (c *GRPCClient) StreamEvents(ctx context.Context, events <-chan *GRPCEvent) (*StreamStats, error) {
	c.mu.RLock()
	if len(c.conns) == 0 {
		c.mu.RUnlock()
		return nil, fmt.Errorf("not connected, call Connect() first")
	}
	numConns := len(c.conns)
	c.mu.RUnlock()

	// Add API key to context
	ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", c.apiKey)

	start := time.Now()
	var totalSent, totalSuccess, totalFailed int64

	if numConns == 1 {
		// Single stream mode
		sent, success, failed := c.runSingleStream(ctx, c.conns[0], events)
		totalSent = sent
		totalSuccess = success
		totalFailed = failed
	} else {
		// Multi-stream mode - distribute events across streams
		totalSent, totalSuccess, totalFailed = c.runMultiStream(ctx, events)
	}

	elapsed := time.Since(start)
	rate := float64(totalSent) / elapsed.Seconds()

	return &StreamStats{
		TotalSent:     totalSent,
		TotalSuccess:  totalSuccess,
		TotalFailed:   totalFailed,
		Duration:      elapsed,
		EventsPerSec:  rate,
		ActiveStreams: numConns,
	}, nil
}

// StreamEventsSlice is a convenience method that streams a slice of events.
func (c *GRPCClient) StreamEventsSlice(ctx context.Context, events []*GRPCEvent) (*StreamStats, error) {
	ch := make(chan *GRPCEvent, len(events))
	for _, e := range events {
		ch <- e
	}
	close(ch)
	return c.StreamEvents(ctx, ch)
}

func (c *GRPCClient) runSingleStream(ctx context.Context, conn *grpc.ClientConn, events <-chan *GRPCEvent) (sent, success, failed int64) {
	// Create EventService client from the generated proto
	client := pb.NewEventServiceClient(conn)

	// Open bidirectional stream
	stream, err := client.StreamEvents(ctx)
	if err != nil {
		// If stream fails to open, count all events as failed
		for range events {
			sent++
			failed++
		}
		return
	}

	// Start goroutine to receive responses
	responseChan := make(chan *pb.EventResponse, 1000)
	go func() {
		defer close(responseChan)
		for {
			resp, err := stream.Recv()
			if err != nil {
				return
			}
			responseChan <- resp
		}
	}()

	// Track send errors separately from server-side failures
	var sendErrors int64

	// Send events
	for event := range events {
		// Convert GRPCEvent to proto EventRequest
		req := &pb.EventRequest{
			TenantId:     "", // Will be set from API key context
			UserId:       event.UserID,
			EventType:    event.EventType,
			DocumentHash: event.DocumentHash,
		}

		// Add timestamp if provided
		if event.Timestamp != nil {
			req.Timestamp = &pb.Timestamp{
				Seconds: event.Timestamp.Unix(),
				Nanos:   int32(event.Timestamp.Nanosecond()),
			}
		}

		// Convert Data map to Metadata
		if event.Data != nil {
			req.Metadata = &pb.Metadata{
				Fields: make(map[string]string),
			}
			for k, v := range event.Data {
				// Convert value to string (JSON for complex types)
				switch val := v.(type) {
				case string:
					req.Metadata.Fields[k] = val
				default:
					jsonBytes, _ := json.Marshal(val)
					req.Metadata.Fields[k] = string(jsonBytes)
				}
			}
		}

		sent++ // Count all attempts
		if err := stream.Send(req); err != nil {
			sendErrors++
		}
	}

	// Close send side
	stream.CloseSend()

	// Drain responses to get server-side success/failure counts
	var serverSuccess, serverFailed int64
	for resp := range responseChan {
		if resp.Status == "error" || resp.Status == "failed" {
			serverFailed++
		} else {
			serverSuccess++
		}
	}

	// Calculate final counts:
	// - If we got responses, use them as the authoritative count
	// - If no responses (async processing), assume sent - sendErrors succeeded
	if serverSuccess > 0 || serverFailed > 0 {
		success = serverSuccess
		failed = serverFailed + sendErrors
	} else {
		// No responses received - assume all successfully sent events succeeded
		success = sent - sendErrors
		failed = sendErrors
	}

	return
}

func (c *GRPCClient) runMultiStream(ctx context.Context, events <-chan *GRPCEvent) (totalSent, totalSuccess, totalFailed int64) {
	c.mu.RLock()
	numConns := len(c.conns)
	conns := make([]*grpc.ClientConn, numConns)
	copy(conns, c.conns)
	c.mu.RUnlock()

	// Create per-stream channels
	streamChans := make([]chan *GRPCEvent, numConns)
	for i := range streamChans {
		streamChans[i] = make(chan *GRPCEvent, 10000)
	}

	var wg sync.WaitGroup
	var sent, success, failed int64

	// Start stream workers
	for i, conn := range conns {
		wg.Add(1)
		go func(idx int, conn *grpc.ClientConn, ch <-chan *GRPCEvent) {
			defer wg.Done()
			s, succ, f := c.runSingleStream(ctx, conn, ch)
			atomic.AddInt64(&sent, s)
			atomic.AddInt64(&success, succ)
			atomic.AddInt64(&failed, f)
		}(i, conn, streamChans[i])
	}

	// Distribute events round-robin
	idx := 0
	for event := range events {
		streamChans[idx%numConns] <- event
		idx++
	}

	// Close all stream channels
	for _, ch := range streamChans {
		close(ch)
	}

	wg.Wait()
	return sent, success, failed
}

// MultiStreamClient provides a higher-level API for multi-stream ingestion.
// It automatically manages connections and provides simple Send/Flush methods.
type MultiStreamClient struct {
	*GRPCClient
	eventChan chan *GRPCEvent
	doneChan  chan struct{}
	stats     *StreamStats
	err       error
	wg        sync.WaitGroup
	started   bool
	mu        sync.Mutex
	dropped   int64 // Count of events dropped by TrySend
}

// NewMultiStreamClient creates a client optimized for high-throughput streaming.
//
// Example:
//
//	client, err := proofchain.NewMultiStreamClient("your-api-key",
//	    proofchain.WithNumStreams(8),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Send events
//	for i := 0; i < 100000; i++ {
//	    client.Send(&proofchain.GRPCEvent{
//	        UserID:    fmt.Sprintf("user-%d", i%1000),
//	        EventType: "action",
//	        Data:      map[string]interface{}{"index": i},
//	    })
//	}
//
//	// Wait for completion and get stats
//	stats, err := client.Flush()
//	fmt.Printf("Sent %d events at %.2f events/sec\n", stats.TotalSent, stats.EventsPerSec)
func NewMultiStreamClient(apiKey string, opts ...GRPCClientOption) (*MultiStreamClient, error) {
	// Default to 4 streams for multi-stream client
	defaultOpts := []GRPCClientOption{WithNumStreams(4)}
	opts = append(defaultOpts, opts...)

	grpcClient := NewGRPCClient(apiKey, opts...)

	ctx, cancel := context.WithTimeout(context.Background(), grpcClient.timeout)
	defer cancel()

	if err := grpcClient.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &MultiStreamClient{
		GRPCClient: grpcClient,
		eventChan:  make(chan *GRPCEvent, 100000),
		doneChan:   make(chan struct{}),
	}, nil
}

// Start begins the streaming session. Must be called before Send.
func (c *MultiStreamClient) Start(ctx context.Context) {
	c.mu.Lock()
	if c.started {
		c.mu.Unlock()
		return
	}
	c.started = true
	c.mu.Unlock()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		stats, err := c.StreamEvents(ctx, c.eventChan)
		c.stats = stats
		c.err = err
		close(c.doneChan)
	}()
}

// Send queues an event for streaming, blocking if the buffer is full.
// This ensures no events are silently dropped.
// Automatically starts the streaming session on first call if not already started.
func (c *MultiStreamClient) Send(event *GRPCEvent) bool {
	c.mu.Lock()
	if !c.started {
		c.started = true
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			stats, err := c.StreamEvents(context.Background(), c.eventChan)
			c.stats = stats
			c.err = err
			close(c.doneChan)
		}()
	}
	c.mu.Unlock()

	c.eventChan <- event
	return true
}

// TrySend attempts to queue an event without blocking.
// Returns false if the client is not started or buffer is full.
// Use Send() instead unless you need non-blocking behavior and handle drops.
func (c *MultiStreamClient) TrySend(event *GRPCEvent) bool {
	c.mu.Lock()
	started := c.started
	c.mu.Unlock()

	if !started {
		return false
	}

	select {
	case c.eventChan <- event:
		return true
	default:
		atomic.AddInt64(&c.dropped, 1)
		return false
	}
}

// SendBlocking is an alias for Send (which now blocks by default).
// Deprecated: Use Send() instead.
func (c *MultiStreamClient) SendBlocking(event *GRPCEvent) {
	c.eventChan <- event
}

// Flush closes the event channel and waits for all events to be sent.
// Returns the final statistics including any dropped events from TrySend.
func (c *MultiStreamClient) Flush() (*StreamStats, error) {
	c.mu.Lock()
	if !c.started {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not started")
	}
	c.mu.Unlock()

	close(c.eventChan)
	c.wg.Wait()

	// Add dropped count to stats
	if c.stats != nil {
		c.stats.TotalDropped = atomic.LoadInt64(&c.dropped)
	}

	return c.stats, c.err
}

// Close closes the client and all connections.
func (c *MultiStreamClient) Close() error {
	c.mu.Lock()
	if c.started {
		select {
		case <-c.doneChan:
			// Already done
		default:
			close(c.eventChan)
		}
	}
	c.mu.Unlock()

	c.wg.Wait()
	return c.GRPCClient.Close()
}

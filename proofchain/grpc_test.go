package proofchain

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/ProofChainZA/proofchain-go/proofchain/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// testEventServer is an in-memory implementation of the gRPC EventService
// used to exercise the streaming client without a live backend.
type testEventServer struct {
	pb.UnimplementedEventServiceServer

	sendResponses  bool   // if false, the server drains without replying (async path)
	responseStatus string // status returned per event; "" => "queued"

	mu      sync.Mutex
	gotReqs []*pb.EventRequest
}

func (s *testEventServer) StreamEvents(stream grpc.BidiStreamingServer[pb.EventRequest, pb.EventResponse]) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		s.mu.Lock()
		s.gotReqs = append(s.gotReqs, req)
		s.mu.Unlock()

		if s.sendResponses {
			status := s.responseStatus
			if status == "" {
				status = "queued"
			}
			if err := stream.Send(&pb.EventResponse{EventId: req.UserId, Status: status}); err != nil {
				return err
			}
		}
	}
}

func (s *testEventServer) requestCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.gotReqs)
}

// startBufconnServer spins up an in-process gRPC server backed by bufconn and
// returns a live *grpc.ClientConn wired to it. Everything is torn down via t.Cleanup.
func startBufconnServer(t *testing.T, srv pb.EventServiceServer) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	gs := grpc.NewServer()
	pb.RegisterEventServiceServer(gs, srv)
	go func() { _ = gs.Serve(lis) }()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufconn: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		gs.Stop()
		_ = lis.Close()
	})
	return conn
}

func channelOf(events ...*GRPCEvent) chan *GRPCEvent {
	ch := make(chan *GRPCEvent, len(events))
	for _, e := range events {
		ch <- e
	}
	close(ch)
	return ch
}

// mustGRPCClient fails the test on constructor error.
func mustGRPCClient(t *testing.T, apiKey string, opts ...GRPCClientOption) *GRPCClient {
	t.Helper()
	c, err := NewGRPCClient(apiKey, opts...)
	if err != nil {
		t.Fatalf("NewGRPCClient: %v", err)
	}
	return c
}

func TestNewGRPCClientDefaults(t *testing.T) {
	c := mustGRPCClient(t, "my-key")
	if c.apiKey != "my-key" {
		t.Errorf("apiKey = %q, want my-key", c.apiKey)
	}
	if c.endpoint != defaultGRPCEndpoint {
		t.Errorf("endpoint = %q, want %q", c.endpoint, defaultGRPCEndpoint)
	}
	if c.numStreams != 1 {
		t.Errorf("numStreams = %d, want 1", c.numStreams)
	}
	if c.plaintext {
		t.Error("plaintext = true, want TLS by default")
	}
}

func TestGRPCClientOptions(t *testing.T) {
	tests := []struct {
		name           string
		opts           []GRPCClientOption
		wantEndpoint   string
		wantNumStreams int
		wantPlaintext  bool
		wantErr        bool
	}{
		{
			name:           "non-443 endpoint still uses TLS",
			opts:           []GRPCClientOption{WithGRPCEndpoint("localhost:50051")},
			wantEndpoint:   "localhost:50051",
			wantNumStreams: 1,
		},
		{
			name:           "explicit 443 endpoint keeps TLS",
			opts:           []GRPCClientOption{WithGRPCEndpoint("grpc.example.com:443")},
			wantEndpoint:   "grpc.example.com:443",
			wantNumStreams: 1,
		},
		{
			name:           "num streams and endpoint",
			opts:           []GRPCClientOption{WithGRPCEndpoint("host:8080"), WithNumStreams(8)},
			wantEndpoint:   "host:8080",
			wantNumStreams: 8,
		},
		{
			name:           "num streams zero is ignored",
			opts:           []GRPCClientOption{WithGRPCEndpoint("host:8080"), WithNumStreams(0)},
			wantEndpoint:   "host:8080",
			wantNumStreams: 1,
		},
		{
			name:           "plaintext allowed for loopback",
			opts:           []GRPCClientOption{WithGRPCEndpoint("localhost:50051"), WithInsecurePlaintext()},
			wantEndpoint:   "localhost:50051",
			wantNumStreams: 1,
			wantPlaintext:  true,
		},
		{
			name:    "plaintext rejected for remote host",
			opts:    []GRPCClientOption{WithGRPCEndpoint("host:443"), WithInsecurePlaintext()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewGRPCClient("k", tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if c.endpoint != tt.wantEndpoint {
				t.Errorf("endpoint = %q, want %q", c.endpoint, tt.wantEndpoint)
			}
			if c.numStreams != tt.wantNumStreams {
				t.Errorf("numStreams = %d, want %d", c.numStreams, tt.wantNumStreams)
			}
			if c.plaintext != tt.wantPlaintext {
				t.Errorf("plaintext = %v, want %v", c.plaintext, tt.wantPlaintext)
			}
		})
	}
}

func TestWithGRPCTimeout(t *testing.T) {
	c := mustGRPCClient(t, "k", WithGRPCTimeout(90*time.Second))
	if c.timeout != 90*time.Second {
		t.Errorf("timeout = %v, want 90s", c.timeout)
	}
}

func TestStreamEventsNotConnected(t *testing.T) {
	c := mustGRPCClient(t, "k")
	_, err := c.StreamEvents(context.Background(), channelOf())
	if err == nil {
		t.Fatal("expected error when not connected, got nil")
	}
}

func TestRunSingleStreamAllSucceed(t *testing.T) {
	srv := &testEventServer{sendResponses: true}
	conn := startBufconnServer(t, srv)
	c := mustGRPCClient(t, "k")

	events := channelOf(
		&GRPCEvent{UserID: "u1", EventType: "action"},
		&GRPCEvent{UserID: "u2", EventType: "action"},
		&GRPCEvent{UserID: "u3", EventType: "action"},
	)
	sent, success, failed := c.runSingleStream(context.Background(), conn, events)

	if sent != 3 {
		t.Errorf("sent = %d, want 3", sent)
	}
	if success != 3 {
		t.Errorf("success = %d, want 3", success)
	}
	if failed != 0 {
		t.Errorf("failed = %d, want 0", failed)
	}
	if srv.requestCount() != 3 {
		t.Errorf("server received %d events, want 3", srv.requestCount())
	}
}

func TestRunSingleStreamServerReportsFailures(t *testing.T) {
	srv := &testEventServer{sendResponses: true, responseStatus: "error"}
	conn := startBufconnServer(t, srv)
	c := mustGRPCClient(t, "k")

	events := channelOf(
		&GRPCEvent{UserID: "u1", EventType: "action"},
		&GRPCEvent{UserID: "u2", EventType: "action"},
	)
	sent, success, failed := c.runSingleStream(context.Background(), conn, events)

	if sent != 2 {
		t.Errorf("sent = %d, want 2", sent)
	}
	if success != 0 {
		t.Errorf("success = %d, want 0", success)
	}
	if failed != 2 {
		t.Errorf("failed = %d, want 2", failed)
	}
}

func TestRunSingleStreamNoResponsesAssumesSuccess(t *testing.T) {
	// When the server acknowledges asynchronously (no per-event responses),
	// the client assumes all successfully-sent events succeeded.
	srv := &testEventServer{sendResponses: false}
	conn := startBufconnServer(t, srv)
	c := mustGRPCClient(t, "k")

	events := channelOf(
		&GRPCEvent{UserID: "u1", EventType: "action"},
		&GRPCEvent{UserID: "u2", EventType: "action"},
	)
	sent, success, failed := c.runSingleStream(context.Background(), conn, events)

	if sent != 2 {
		t.Errorf("sent = %d, want 2", sent)
	}
	if success != 2 {
		t.Errorf("success = %d, want 2", success)
	}
	if failed != 0 {
		t.Errorf("failed = %d, want 0", failed)
	}
}

func TestRunSingleStreamConvertsMetadataAndTimestamp(t *testing.T) {
	srv := &testEventServer{sendResponses: true}
	conn := startBufconnServer(t, srv)
	c := mustGRPCClient(t, "k")

	ts := time.Date(2026, 7, 21, 12, 0, 0, 500, time.UTC)
	events := channelOf(&GRPCEvent{
		UserID:       "u1",
		EventType:    "purchase",
		DocumentHash: "hash123",
		Timestamp:    &ts,
		Data: map[string]any{
			"plain":   "value",                // strings pass through verbatim
			"complex": map[string]int{"n": 1}, // non-strings are JSON-encoded
		},
	})

	c.runSingleStream(context.Background(), conn, events)

	srv.mu.Lock()
	defer srv.mu.Unlock()
	if len(srv.gotReqs) != 1 {
		t.Fatalf("server received %d requests, want 1", len(srv.gotReqs))
	}
	req := srv.gotReqs[0]
	if req.UserId != "u1" || req.EventType != "purchase" || req.DocumentHash != "hash123" {
		t.Errorf("core fields not mapped: %+v", req)
	}
	if req.Timestamp == nil || req.Timestamp.Seconds != ts.Unix() {
		t.Errorf("timestamp not mapped: %+v", req.Timestamp)
	}
	if req.Metadata == nil {
		t.Fatal("metadata is nil")
	}
	if req.Metadata.Fields["plain"] != "value" {
		t.Errorf("plain field = %q, want value", req.Metadata.Fields["plain"])
	}
	if req.Metadata.Fields["complex"] != `{"n":1}` {
		t.Errorf("complex field = %q, want JSON-encoded", req.Metadata.Fields["complex"])
	}
}

func TestRunMultiStreamDistributesEvents(t *testing.T) {
	srv := &testEventServer{sendResponses: true}
	// Two connections to the same in-memory server; round-robin distribution
	// should still deliver every event exactly once.
	conn1 := startBufconnServer(t, srv)
	conn2 := startBufconnServer(t, srv)

	c := mustGRPCClient(t, "k", WithNumStreams(2))
	c.conns = []*grpc.ClientConn{conn1, conn2}

	const n = 10
	events := make([]*GRPCEvent, n)
	for i := range events {
		events[i] = &GRPCEvent{UserID: "u", EventType: "action"}
	}
	ch := channelOf(events...)

	sent, success, failed := c.runMultiStream(context.Background(), ch)
	if sent != n {
		t.Errorf("sent = %d, want %d", sent, n)
	}
	if success != n {
		t.Errorf("success = %d, want %d", success, n)
	}
	if failed != 0 {
		t.Errorf("failed = %d, want 0", failed)
	}
	if srv.requestCount() != n {
		t.Errorf("server received %d events, want %d", srv.requestCount(), n)
	}
}

func TestStreamEventsSliceSingleStream(t *testing.T) {
	srv := &testEventServer{sendResponses: true}
	conn := startBufconnServer(t, srv)

	c := mustGRPCClient(t, "k")
	c.conns = []*grpc.ClientConn{conn}

	stats, err := c.StreamEventsSlice(context.Background(), []*GRPCEvent{
		{UserID: "u1", EventType: "action"},
		{UserID: "u2", EventType: "action"},
	})
	if err != nil {
		t.Fatalf("StreamEventsSlice failed: %v", err)
	}
	if stats.TotalSent != 2 {
		t.Errorf("TotalSent = %d, want 2", stats.TotalSent)
	}
	if stats.TotalSuccess != 2 {
		t.Errorf("TotalSuccess = %d, want 2", stats.TotalSuccess)
	}
	if stats.ActiveStreams != 1 {
		t.Errorf("ActiveStreams = %d, want 1", stats.ActiveStreams)
	}
	if stats.EventsPerSec <= 0 {
		t.Errorf("EventsPerSec = %f, want > 0", stats.EventsPerSec)
	}
}

func TestStreamEventsMultiStream(t *testing.T) {
	srv := &testEventServer{sendResponses: true}
	conn1 := startBufconnServer(t, srv)
	conn2 := startBufconnServer(t, srv)

	c := mustGRPCClient(t, "k", WithNumStreams(2))
	c.conns = []*grpc.ClientConn{conn1, conn2}

	stats, err := c.StreamEventsSlice(context.Background(), []*GRPCEvent{
		{UserID: "u1", EventType: "action"},
		{UserID: "u2", EventType: "action"},
		{UserID: "u3", EventType: "action"},
	})
	if err != nil {
		t.Fatalf("StreamEvents failed: %v", err)
	}
	if stats.TotalSent != 3 {
		t.Errorf("TotalSent = %d, want 3", stats.TotalSent)
	}
	if stats.ActiveStreams != 2 {
		t.Errorf("ActiveStreams = %d, want 2", stats.ActiveStreams)
	}
}

func TestMultiStreamClientFlushBeforeStart(t *testing.T) {
	// Constructed manually to avoid dialing a live endpoint.
	c := &MultiStreamClient{
		GRPCClient: mustGRPCClient(t, "k"),
		eventChan:  make(chan *GRPCEvent, 10),
		doneChan:   make(chan struct{}),
	}
	if _, err := c.Flush(); err == nil {
		t.Error("expected error flushing before start, got nil")
	}
}

func TestMultiStreamClientTrySendBeforeStart(t *testing.T) {
	c := &MultiStreamClient{
		GRPCClient: mustGRPCClient(t, "k"),
		eventChan:  make(chan *GRPCEvent, 10),
		doneChan:   make(chan struct{}),
	}
	if c.TrySend(&GRPCEvent{UserID: "u"}) {
		t.Error("TrySend before start should return false")
	}
}

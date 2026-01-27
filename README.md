# ProofChain Go SDK

Official Go SDK for [ProofChain](https://proofchain.co.za) - blockchain-anchored document attestation.

## Installation

```bash
go get github.com/ProofChainZA/proofchain-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ProofChainZA/proofchain-go/proofchain"
)

func main() {
    // Create client
    client := proofchain.NewClient("your-api-key")

    // Attest a document
    result, err := client.Documents.Attest(context.Background(), &proofchain.AttestRequest{
        FilePath:  "contract.pdf",
        UserID:    "user@example.com",
        EventType: "contract_signed",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("IPFS Hash: %s\n", result.IPFSHash)
    fmt.Printf("Verify URL: %s\n", result.VerifyURL)

    // Verify
    verification, err := client.Verify(context.Background(), result.IPFSHash)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Valid: %v\n", verification.Valid)
}
```

## High-Performance Ingestion

For maximum throughput, use the dedicated `IngestionClient` which connects directly to the Rust ingestion API:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ProofChainZA/proofchain-go/proofchain"
)

func main() {
    // Create high-performance ingestion client
    ingestion := proofchain.NewIngestionClient("your-api-key")
    ctx := context.Background()

    // Single event (immediate attestation)
    result, err := ingestion.Ingest(ctx, &proofchain.IngestEventRequest{
        UserID:    "user-123",
        EventType: "purchase",
        Data:      map[string]interface{}{"amount": 99.99, "product": "widget"},
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Event ID: %s\n", result.EventID)
    fmt.Printf("Certificate: %s\n", result.CertificateID)

    // Batch events (up to 1000 per request)
    batchResult, err := ingestion.IngestBatch(ctx, &proofchain.BatchIngestRequest{
        Events: []proofchain.IngestEventRequest{
            {UserID: "user-1", EventType: "click", Data: map[string]interface{}{"page": "/home"}},
            {UserID: "user-2", EventType: "click", Data: map[string]interface{}{"page": "/products"}},
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Queued: %d\n", batchResult.Queued)
    fmt.Printf("Failed: %d\n", batchResult.Failed)
}
```

## gRPC Multi-Stream Mode (Maximum Throughput)

For maximum throughput (1000+ events/sec), use gRPC multi-stream mode which creates
parallel connections to distribute load across server pods:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ProofChainZA/proofchain-go/proofchain"
)

func main() {
    // Create multi-stream client with 8 parallel streams
    client, err := proofchain.NewMultiStreamClient("your-api-key",
        proofchain.WithNumStreams(8),
        proofchain.WithGRPCEndpoint("grpc.proofchain.co.za:443"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Start streaming
    ctx := context.Background()
    client.Start(ctx)

    // Send 100,000 events
    for i := 0; i < 100000; i++ {
        client.SendBlocking(&proofchain.GRPCEvent{
            UserID:    fmt.Sprintf("user-%d", i%1000),
            EventType: "user.action",
            Data: map[string]interface{}{
                "action": "click",
                "index":  i,
            },
        })
    }

    // Flush and get stats
    stats, err := client.Flush()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Sent: %d events\n", stats.TotalSent)
    fmt.Printf("Success: %d, Failed: %d\n", stats.TotalSuccess, stats.TotalFailed)
    fmt.Printf("Duration: %v\n", stats.Duration)
    fmt.Printf("Rate: %.2f events/sec\n", stats.EventsPerSec)
}
```

### Lower-Level gRPC Streaming

For more control, use the `GRPCClient` directly:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ProofChainZA/proofchain-go/proofchain"
)

func main() {
    // Create gRPC client with 4 parallel streams
    client := proofchain.NewGRPCClient("your-api-key",
        proofchain.WithNumStreams(4),
    )

    // Connect
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Create event channel
    events := make(chan *proofchain.GRPCEvent, 1000)

    // Send events in goroutine
    go func() {
        for i := 0; i < 10000; i++ {
            events <- &proofchain.GRPCEvent{
                UserID:    fmt.Sprintf("user-%d", i%100),
                EventType: "action",
                Data:      map[string]interface{}{"index": i},
            }
        }
        close(events)
    }()

    // Stream and get stats
    stats, err := client.StreamEvents(ctx, events)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Rate: %.2f events/sec\n", stats.EventsPerSec)
}
```

## State Channels (High-Volume Streaming)

For high-throughput scenarios (100K+ events/sec):

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ProofChainZA/proofchain-go/proofchain"
)

func main() {
    client := proofchain.NewClient("your-api-key")
    ctx := context.Background()

    // Create a state channel
    channel, err := client.Channels.Create(ctx, &proofchain.CreateChannelRequest{
        Name: "iot-sensors",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Stream events
    for i := 0; i < 1000; i++ {
        _, err := client.Channels.Stream(ctx, channel.ChannelID, &proofchain.StreamEventRequest{
            EventType: "sensor_reading",
            UserID:    "sensor-001",
            Data: map[string]interface{}{
                "temperature": 22.5,
                "humidity":    65.0,
            },
        })
        if err != nil {
            log.Printf("Stream error: %v", err)
        }
    }

    // Settle on-chain
    settlement, err := client.Channels.Settle(ctx, channel.ChannelID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("TX Hash: %s\n", settlement.TxHash)
    fmt.Printf("Events Settled: %d\n", settlement.EventCount)
}
```

## Features

### Documents

```go
// Attest a file
result, err := client.Documents.Attest(ctx, &proofchain.AttestRequest{
    FilePath:  "document.pdf",
    UserID:    "user@example.com",
    EventType: "document_uploaded",
    Metadata:  map[string]interface{}{"department": "legal"},
})

// Attest raw bytes
result, err := client.Documents.AttestBytes(ctx, &proofchain.AttestBytesRequest{
    Content:   []byte("raw content"),
    Filename:  "data.json",
    UserID:    "user@example.com",
})

// Get document by hash
doc, err := client.Documents.Get(ctx, "Qm...")
```

### Events

```go
// Create an event
event, err := client.Events.Create(ctx, &proofchain.CreateEventRequest{
    EventType: "user_action",
    UserID:    "user123",
    Data:      map[string]interface{}{"action": "login"},
})

// List events
events, err := client.Events.List(ctx, &proofchain.ListEventsRequest{
    UserID: "user123",
    Limit:  100,
})

// Search events
results, err := client.Events.Search(ctx, &proofchain.SearchRequest{
    Query:     "login",
    StartDate: "2024-01-01",
    EndDate:   "2024-12-31",
})
```

### Verification

```go
// Verify by IPFS hash
result, err := client.Verify(ctx, "Qm...")

// Check result
if result.Valid {
    fmt.Printf("Verified at: %s\n", result.Timestamp)
    fmt.Printf("Blockchain TX: %s\n", result.BlockchainTx)
}
```

### Certificates

```go
// Issue a certificate
cert, err := client.Certificates.Issue(ctx, &proofchain.IssueCertificateRequest{
    RecipientName:  "John Doe",
    RecipientEmail: "john@example.com",
    Title:          "Course Completion",
    Description:    "Completed Go Fundamentals",
    Metadata:       map[string]interface{}{"course_id": "GO101"},
})

fmt.Printf("Certificate ID: %s\n", cert.CertificateID)
fmt.Printf("QR Code: %s\n", cert.QRCodeURL)

// Revoke
_, err = client.Certificates.Revoke(ctx, cert.CertificateID, "Issued in error")
```

### Webhooks

```go
// Register a webhook
webhook, err := client.Webhooks.Create(ctx, &proofchain.CreateWebhookRequest{
    URL:    "https://your-app.com/webhook",
    Events: []string{"document.attested", "channel.settled"},
    Secret: "your-secret",
})

// List webhooks
webhooks, err := client.Webhooks.List(ctx)

// Delete
err = client.Webhooks.Delete(ctx, webhook.ID)
```

## Error Handling

```go
result, err := client.Documents.Attest(ctx, req)
if err != nil {
    switch e := err.(type) {
    case *proofchain.AuthenticationError:
        log.Fatal("Invalid API key")
    case *proofchain.RateLimitError:
        log.Printf("Rate limited, retry after %d seconds", e.RetryAfter)
    case *proofchain.ValidationError:
        log.Printf("Validation error: %s", e.Message)
    case *proofchain.NotFoundError:
        log.Fatal("Resource not found")
    default:
        log.Fatalf("API error: %v", err)
    }
}
```

## Configuration

```go
// Custom configuration
client := proofchain.NewClient(
    "your-api-key",
    proofchain.WithBaseURL("https://ingest.proofchain.co.za"),  // Ingestion API (Rust)
    proofchain.WithMgmtURL("https://api.proofchain.co.za"),     // Management API (Python)
    proofchain.WithTimeout(30 * time.Second),
    proofchain.WithRetries(3),
)

// From environment variable
// Set PROOFCHAIN_API_KEY
client := proofchain.NewClientFromEnv()
```

## Context Support

All methods accept a `context.Context` for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := client.Documents.Attest(ctx, req)
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- Documentation: https://proofchain.co.za/docs
- Email: support@proofchain.co.za
- GitHub Issues: https://github.com/proofchain/proofchain-go/issues
# proofchain-go
# proofchain-go

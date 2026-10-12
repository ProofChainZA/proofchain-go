# AGENTS.md — ProofChain Go Service

## Scope

Go service providing blockchain-anchored document attestation, event streaming, certificates, and identity management for the ProofChain platform.

This is the **Go service** (backend API + gRPC ingestion), distinct from `proofchain-sdks/go/` which is the **client SDK**.

## Conventions

### Code Style
- Go 1.26+; SDK module: `github.com/ProofChainZA/proofchain-go/proofchain` (nested `proofchain/go.mod`; the root `go.mod` only covers `examples/`)
- **Resource pattern**: each domain has a `*Resource` struct with HTTP client reference (e.g., `DocumentsResource`, `EventsResource`).
- **Constructor pattern**: `NewClient(apiKey, opts...) (*Client, error)`, `NewClientFromEnv(opts...)`, `NewEndUserClient(token, tenant, opts...)`, `NewIngestionClient(...)`, `NewGRPCClient(...)` — all return `(T, error)`; the error is a rejected base URL / endpoint.
- **Functional options**: `HTTPClientOption` for configuring base URL, timeouts, retries, `WithInsecureHTTP()`.
- **Transport security**: base URLs must be `https://` (plain `http://` only for loopback hosts or with the explicit insecure option); redirects are never followed; gRPC always uses TLS ≥1.2 unless `WithInsecurePlaintext()` targets a loopback endpoint.
- **Path identifiers**: every caller-supplied path segment goes through `seg()` (`url.PathEscape`); query parameters use `url.Values`.
- **Secrets**: clients and `CreatedAPIKey` redact the key in `String()`/`%v`/`%#v`; `CreateAPIKey` requires non-admin scopes and defaults to a 90-day expiry; there is no key export.
- **Webhooks**: verify deliveries with `VerifyWebhookSignature(secret, rawBody, header)` (`X-Webhook-Signature: sha256=<hex>`).
- **Context-first**: all public methods accept `context.Context` as first argument.
- **Error handling**: return `(T, error)`; never panic in library code.

### Types
- **Custom `Timestamp` type**: handles multiple JSON date formats (RFC3339, RFC3339Nano, etc.).
- **String enums**: use typed constants (e.g., `AttestationMode`, `EventStatus`, `ChannelState`).
- **Pointer fields for optional values**: e.g., `*string`, `*Timestamp` for omitempty.

### gRPC
- Protobuf definitions in `proofchain/pb/`
- Generated files: `*.pb.go`, `*_grpc.pb.go` — do not hand-edit.
- Regenerate with: `protoc --go_out=. --go-grpc_out=. <proto file>`

### Testing
- Unit tests: `*_test.go` alongside source files.
- Use `httptest` for HTTP client mocking.
- Examples in `examples/` directory (quickstart, csv_ingest, channels).

## Anti-Patterns

- **NO `panic` in production code** — always return errors.
- **NO hardcoded API keys** — use `NewClientFromEnv()` or pass via options.
- **NO ignoring errors** from HTTP requests — always check `err != nil`.
- **NO manual JSON parsing** — use struct tags and `json.Unmarshal`.
- **NO breaking changes to public API** without version bump.
- **NO mixing SDK and service concerns**: this is the service, not the client SDK.

## Dependencies

- **Core**: `google.golang.org/grpc` v1.83.1 (`golang.org/x/net` v0.60.0)
- **Protobuf**: `google.golang.org/protobuf` v1.36.11
- **Standard library only** for HTTP client (no external HTTP libs).

## Testing

```bash
go test ./...                 # live API tests run only with PROOFCHAIN_LIVE_TESTS=1 and PROOFCHAIN_API_KEY
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go vet ./...
go fmt ./...
```

- Run tests before any commit.
- Use `go mod tidy` to clean dependencies.
- Verify examples compile: `go build ./examples/...`

## Examples

```
proofchain/client.go          # Main client constructor
proofchain/types.go           # Shared types and enums
proofchain/documents.go       # Document attestation resource
proofchain/events.go          # Event management resource
proofchain/channels.go        # State channel resource
proofchain/certificates.go    # Certificate resource
proofchain/grpc.go            # gRPC client methods
proofchain/pb/                # Protobuf generated code
examples/quickstart/main.go   # Basic usage example
```

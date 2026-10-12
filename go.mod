module github.com/ProofChainZA/proofchain-go

go 1.26.0

require google.golang.org/grpc v1.83.1 // indirect

require (
	github.com/ProofChainZA/proofchain-go/proofchain v0.0.0
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/ProofChainZA/proofchain-go/proofchain => ./proofchain

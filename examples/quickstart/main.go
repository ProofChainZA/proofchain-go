// ProofChain Go SDK Quick Start Example
//
// This example demonstrates basic usage of the ProofChain Go SDK.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/proofchain/proofchain-go/proofchain"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("PROOFCHAIN_API_KEY")
	if apiKey == "" {
		log.Fatal("PROOFCHAIN_API_KEY environment variable not set")
	}

	// Create client
	client := proofchain.NewClient(apiKey)
	ctx := context.Background()

	// Example 1: Create an event
	fmt.Println("Creating an event...")
	event, err := client.Events.Create(ctx, &proofchain.CreateEventRequest{
		EventType: "user_login",
		UserID:    "user@example.com",
		Data: map[string]interface{}{
			"ip_address": "192.168.1.1",
			"user_agent": "Mozilla/5.0...",
			"location":   "Cape Town, ZA",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create event: %v", err)
	}
	fmt.Printf("  Event ID: %s\n", event.ID)
	fmt.Printf("  IPFS Hash: %s\n", event.IPFSHash)
	fmt.Printf("  Certificate: %s\n", event.CertificateID)
	fmt.Println()

	// Example 2: Verify the event
	fmt.Println("Verifying the event...")
	verification, err := client.Verify(ctx, event.IPFSHash)
	if err != nil {
		log.Fatalf("Failed to verify: %v", err)
	}
	fmt.Printf("  Valid: %v\n", verification.Valid)
	fmt.Printf("  Message: %s\n", verification.Message)
	fmt.Println()

	// Example 3: List recent events
	fmt.Println("Listing recent events...")
	events, err := client.Events.List(ctx, &proofchain.ListEventsRequest{
		Limit: 5,
	})
	if err != nil {
		log.Fatalf("Failed to list events: %v", err)
	}
	for _, e := range events {
		fmt.Printf("  - %s: %s (%s)\n", e.EventType, e.CertificateID, e.Status)
	}
	fmt.Println()

	// Example 4: Search events
	fmt.Println("Searching events...")
	results, err := client.Events.Search(ctx, &proofchain.SearchRequest{
		Query: "login",
		Limit: 10,
	})
	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}
	fmt.Printf("  Found %d matching events\n", results.Total)
	fmt.Println()

	fmt.Println("Done!")
}

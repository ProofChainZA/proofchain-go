// ProofChain Go SDK State Channels Example
//
// This example demonstrates high-throughput event streaming using state channels.
// State channels allow you to stream 100K+ events/sec with periodic on-chain settlement.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/proofchain/proofchain-go/proofchain"
)

func main() {
	apiKey := os.Getenv("PROOFCHAIN_API_KEY")
	if apiKey == "" {
		log.Fatal("PROOFCHAIN_API_KEY environment variable not set")
	}

	client := proofchain.NewClient(apiKey)
	ctx := context.Background()

	// Create a state channel
	fmt.Println("Creating state channel...")
	channel, err := client.Channels.Create(ctx, &proofchain.CreateChannelRequest{
		Name:        "iot-temperature-sensors",
		Description: "Temperature readings from factory floor sensors",
	})
	if err != nil {
		log.Fatalf("Failed to create channel: %v", err)
	}
	fmt.Printf("  Channel ID: %s\n", channel.ChannelID)
	fmt.Printf("  State: %s\n", channel.State)
	fmt.Println()

	// Simulate streaming sensor data
	fmt.Println("Streaming sensor events...")
	sensorIDs := []string{"sensor-001", "sensor-002", "sensor-003"}
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		sensorID := sensorIDs[rand.Intn(len(sensorIDs))]
		temperature := 20.0 + rand.Float64()*10.0 // 20-30°C
		humidity := 40.0 + rand.Float64()*30.0    // 40-70%

		_, err := client.Channels.Stream(ctx, channel.ChannelID, &proofchain.StreamEventRequest{
			EventType: "temperature_reading",
			UserID:    sensorID,
			Data: map[string]interface{}{
				"temperature": fmt.Sprintf("%.2f", temperature),
				"humidity":    fmt.Sprintf("%.2f", humidity),
				"unit":        "celsius",
				"reading_id":  i + 1,
			},
		})
		if err != nil {
			log.Printf("Stream error: %v", err)
		}

		if (i+1)%25 == 0 {
			fmt.Printf("  Streamed %d events...\n", i+1)
		}
	}
	fmt.Println()

	// Check channel status
	fmt.Println("Checking channel status...")
	status, err := client.Channels.Status(ctx, channel.ChannelID)
	if err != nil {
		log.Fatalf("Failed to get status: %v", err)
	}
	fmt.Printf("  Total events: %d\n", status.EventCount)
	fmt.Printf("  Synced to IPFS: %d\n", status.SyncedCount)
	fmt.Printf("  Pending sync: %d\n", status.PendingCount)
	fmt.Println()

	// Wait for sync to complete
	fmt.Println("Waiting for IPFS sync...")
	time.Sleep(5 * time.Second)

	// Settle on-chain
	fmt.Println("Settling channel on-chain...")
	settlement, err := client.Channels.Settle(ctx, channel.ChannelID)
	if err != nil {
		fmt.Printf("  Settlement pending or error: %v\n", err)
	} else {
		fmt.Printf("  Transaction: %s\n", settlement.TxHash)
		fmt.Printf("  Block: %d\n", settlement.BlockNumber)
		fmt.Printf("  Events settled: %d\n", settlement.EventCount)
		fmt.Printf("  Gas used: %d\n", settlement.GasUsed)
	}
	fmt.Println()

	// Close the channel
	fmt.Println("Closing channel...")
	closed, err := client.Channels.Close(ctx, channel.ChannelID)
	if err != nil {
		log.Fatalf("Failed to close channel: %v", err)
	}
	fmt.Printf("  Final state: %s\n", closed.State)
	fmt.Println()

	fmt.Println("Done!")
}

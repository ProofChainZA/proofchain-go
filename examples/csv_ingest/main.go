package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ProofChainZA/proofchain-go/proofchain"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("PROOFCHAIN_API_KEY")
	if apiKey == "" {
		log.Fatal("PROOFCHAIN_API_KEY environment variable is required")
	}

	// CSV files to ingest
	csvFiles := []string{
		"../../docs/source/20ae15b3-f456-433c-bdc5-a0224703c3f7.csv",
		"../../docs/source/9bcdb6ac-4c21-4f9d-b89c-8cb50bd8590d.csv",
	}

	// Create ingestion client
	client := proofchain.NewIngestionClient(apiKey)
	ctx := context.Background()

	totalEvents := 0
	totalQueued := 0
	totalFailed := 0

	for _, csvPath := range csvFiles {
		absPath, err := filepath.Abs(csvPath)
		if err != nil {
			log.Printf("Failed to resolve path %s: %v", csvPath, err)
			continue
		}

		events, err := parseCSV(absPath)
		if err != nil {
			log.Printf("Failed to parse %s: %v", csvPath, err)
			continue
		}

		log.Printf("Parsed %d events from %s", len(events), filepath.Base(csvPath))
		totalEvents += len(events)

		// Batch ingest in chunks of 100
		batchSize := 100
		for i := 0; i < len(events); i += batchSize {
			end := i + batchSize
			if end > len(events) {
				end = len(events)
			}

			batch := events[i:end]
			result, err := client.IngestBatch(ctx, &proofchain.BatchIngestRequest{
				Events: batch,
			})

			if err != nil {
				log.Printf("Batch %d-%d failed: %v", i, end, err)
				totalFailed += len(batch)
				continue
			}

			totalQueued += result.Queued
			totalFailed += result.Failed
			log.Printf("Batch %d-%d: queued=%d, failed=%d", i, end, result.Queued, result.Failed)
		}
	}

	fmt.Println("\n========== INGESTION COMPLETE ==========")
	fmt.Printf("Total events:  %d\n", totalEvents)
	fmt.Printf("Total queued:  %d\n", totalQueued)
	fmt.Printf("Total failed:  %d\n", totalFailed)
}

func parseCSV(path string) ([]proofchain.IngestEventRequest, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Find column indices
	userIDIdx := -1
	eventIdx := -1
	atIdx := -1

	for i, col := range header {
		switch col {
		case "userid":
			userIDIdx = i
		case "event":
			eventIdx = i
		case "at":
			atIdx = i
		}
	}

	if userIDIdx == -1 || eventIdx == -1 {
		return nil, fmt.Errorf("missing required columns (userid, event)")
	}

	var events []proofchain.IngestEventRequest

	for {
		record, err := reader.Read()
		if err != nil {
			break // EOF or error
		}

		userID := record[userIDIdx]
		eventType := record[eventIdx]

		data := map[string]interface{}{
			"source": "csv_import",
		}

		// Parse timestamp if available
		if atIdx != -1 && atIdx < len(record) {
			timestamp := record[atIdx]
			data["original_timestamp"] = timestamp

			// Parse and convert to ISO format
			if t, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
				data["timestamp"] = t.Format(time.RFC3339)
			}
		}

		events = append(events, proofchain.IngestEventRequest{
			UserID:      userID,
			EventType:   eventType,
			Data:        data,
			EventSource: "csv_import",
		})
	}

	return events, nil
}

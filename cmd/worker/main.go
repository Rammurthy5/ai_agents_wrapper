package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rammurthy5/ai_agents_wrapper/internal/facade"
)

func main() {
	// Load configuration
	cfg, err := facade.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize facade
	f := facade.NewFacade(cfg)

	// Worker logic
	ticker := time.NewTicker(10 * time.Second)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Worker started. Press Ctrl+C to stop.")
	for {
		select {
		case <-ticker.C:
			prompt := "Tell me something interesting"
			result := f.GetMergedResults(prompt)
			for _, r := range result.Results {
				if r.Error != nil {
					fmt.Printf("Worker: %s failed: %v\n", r.Source, r.Error)
				} else {
					fmt.Printf("Worker: %s response: %s\n", r.Source, r.Message)
				}
			}
		case <-sigChan:
			ticker.Stop()
			fmt.Println("Worker shutting down")
			return
		}
	}
}

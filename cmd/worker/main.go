package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rammurthy5/ai_agents_wrapper/internal/facade"
	"github.com/Rammurthy5/ai_agents_wrapper/internal/queue"
)

func main() {
	// Load configuration
	cfg, err := facade.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// initialize facade
	f := facade.NewFacade(cfg)

	// initialize RabbitMQ
	rabbit, err := queue.NewRabbitMQ(cfg.RABBITMQ_URL)
	if err != nil {
		log.Fatal("Failed to initialize RabbitMQ:", err)
	}
	defer rabbit.Close()

	// start consuming messages
	msgs, err := rabbit.Consume()
	if err != nil {
		log.Fatal("Failed to consume from queue:", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Worker started. Consuming from queue. Press Ctrl+C to stop.")
	for {
		select {
		case msg := <-msgs:
			var task queue.Message
			if err := json.Unmarshal(msg.Body, &task); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			// process the prompt
			result := f.GetMergedResults(task.Prompt)
			for _, r := range result.Results {
				if r.Error != nil {
					fmt.Printf("Worker: %s failed for prompt '%s': %v\n", r.Source, task.Prompt, r.Error)
				} else {
					fmt.Printf("Worker: %s response for prompt '%s': %s\n", r.Source, task.Prompt, r.Message)
				}
			}
		case <-sigChan:
			fmt.Println("Worker shutting down")
			return
		}
	}
}

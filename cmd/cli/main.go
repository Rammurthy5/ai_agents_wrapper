package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Rammurthy5/ai_agents_wrapper/internal/facade"
	"github.com/Rammurthy5/ai_agents_wrapper/internal/queue"
)

func main() {
	prompt := flag.String("prompt", "", "The prompt to process (required)")
	queueMode := flag.Bool("queue", false, "Send prompt to RabbitMQ instead of processing directly")
	verbose := flag.Bool("v", false, "Enable verbose output")
	flag.Parse()

	if *prompt == "" {
		fmt.Println("Error: -prompt is required")
		flag.Usage()
		os.Exit(1)
	}

	if !*verbose {
		log.SetOutput(io.Discard) // Suppress logs unless -v is set
	}

	cfg, err := facade.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	f := facade.NewFacade(cfg)

	if *queueMode {
		rabbit, err := queue.NewRabbitMQ(cfg.RABBITMQ_URL)
		if err != nil {
			log.Fatalf("Failed to initialize RabbitMQ: %v", err)
		}
		defer rabbit.Close()

		msg := queue.Message{Prompt: *prompt}
		if err := rabbit.Publish(msg); err != nil {
			log.Fatalf("Failed to enqueue prompt: %v", err)
		}
		fmt.Printf("Prompt '%s' successfully queued\n", *prompt)
	} else {
		result := f.GetMergedResults(*prompt)
		for _, r := range result.Results {
			if r.Error != nil {
				fmt.Printf("%s failed: %v\n", r.Source, r.Error)
			} else {
				fmt.Printf("%s: %s\n", r.Source, r.Message)
			}
		}
	}
}

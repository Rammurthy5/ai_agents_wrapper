package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Rammurthy5/ai_agents_wrapper/internal/facade"
	"github.com/Rammurthy5/ai_agents_wrapper/internal/queue"
	"github.com/Rammurthy5/ai_agents_wrapper/internal/storage"

	"github.com/google/uuid"
)

func main() {
	prompt := flag.String("prompt", "", "The prompt to process")
	queueMode := flag.Bool("queue", false, "Send prompt to RabbitMQ instead of processing directly")
	taskID := flag.String("task", "", "Fetch result for a given task ID")
	verbose := flag.Bool("v", false, "Enable verbose output")
	flag.Parse()

	if !*verbose {
		log.SetOutput(io.Discard)
	}

	cfg, err := facade.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *taskID != "" {
		// Fetch result mode
		redisClient, err := storage.NewRedisClient(cfg.Redis_URL)
		if err != nil {
			log.Fatalf("Failed to initialize Redis: %v", err)
		}
		defer redisClient.Close()

		result, err := redisClient.GetResult(*taskID)
		if err != nil {
			log.Fatalf("Failed to fetch result: %v", err)
		}
		if result == nil {
			fmt.Printf("No result found for task %s\n", *taskID)
			os.Exit(1)
		}

		for _, r := range result.Results {
			if r.Error != nil {
				fmt.Printf("%s failed: %v\n", r.Source, r.Error)
			} else {
				fmt.Printf("%s: %s\n", r.Source, r.Message)
			}
		}
		return
	}

	if *prompt == "" {
		fmt.Println("Error: -prompt is required unless -task is provided")
		flag.Usage()
		os.Exit(1)
	}

	f := facade.NewFacade(cfg)

	if *queueMode {
		rabbit, err := queue.NewRabbitMQ(cfg.RABBITMQ_URL)
		if err != nil {
			log.Fatalf("Failed to initialize RabbitMQ: %v", err)
		}
		defer rabbit.Close()

		taskID := uuid.New().String()
		msg := queue.Message{TaskID: taskID, Prompt: *prompt}
		if err := rabbit.Publish(msg); err != nil {
			log.Fatalf("Failed to enqueue prompt: %v", err)
		}
		fmt.Printf("Prompt '%s' queued with task ID: %s\n", *prompt, taskID)
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

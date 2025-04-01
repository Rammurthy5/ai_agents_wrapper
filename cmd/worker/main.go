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
	"github.com/Rammurthy5/ai_agents_wrapper/internal/storage"
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

	redisClient, err := storage.NewRedisClient(cfg.Redis_URL)
	if err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}
	defer redisClient.Close()

	// start consuming messages
	msgs, err := rabbit.Consume()
	if err != nil {
		log.Fatal("Failed to consume from queue:", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Worker started. Consuming from queue. Press Ctrl+C to stop..")
	for {
		select {
		case msg := <-msgs:
			var task queue.Message
			fmt.Printf("Received message %v\n", msg.Body)
			if err := json.Unmarshal(msg.Body, &task); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				fmt.Printf("Failed to unmarshal message: %v", err)
				continue
			}
			fmt.Printf("json unmarshalled task %v\n", task.Prompt)
			// process the prompt
			result := f.GetMergedResults(task.Prompt)
			fmt.Printf("Result: %v\n", result)
			if err := redisClient.StoreResult(task.TaskID, result); err != nil {
				log.Printf("Failed to store result for task %s: %v", task.TaskID, err)
				fmt.Printf("Failed to store result for task %s: %v", task.TaskID, err)
				continue
			}
			fmt.Printf("Stored result for task %s\n", task.TaskID)
		case <-sigChan:
			fmt.Println("Worker shutting down")
			return
		}
	}
}

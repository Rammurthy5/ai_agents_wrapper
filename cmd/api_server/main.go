package main

import (
	"fmt"
	"log"

	"github.com/Rammurthy5/ai_agents_wrapper/internal/facade"
	"github.com/Rammurthy5/ai_agents_wrapper/internal/queue"
	"github.com/Rammurthy5/ai_agents_wrapper/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	// Load configuration
	cfg, err := facade.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize RabbitMQ
	rabbit, err := queue.NewRabbitMQ(cfg.RABBITMQ_URL)
	if err != nil {
		log.Fatal("Failed to initialize RabbitMQ:", err)
	}
	defer rabbit.Close()

	redisClient, err := storage.NewRedisClient(cfg.Redis_URL)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redisClient.Close()

	// Set up gin router
	r := gin.Default()
	r.GET("/getMergedResults", func(c *gin.Context) {
		prompt := c.Query("prompt")
		taskID := uuid.New().String()
		if prompt == "" {
			c.JSON(400, gin.H{"error": "Missing 'prompt' query parameter"})
			return
		}

		// Enqueue the prompt
		msg := queue.Message{Prompt: prompt, TaskID: taskID}
		if err := rabbit.Publish(msg); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to enqueue request: %v", err)})
			return
		}

		c.JSON(202, gin.H{"message": "Request queued successfully", "task_id": taskID})
	})

	r.GET("/results/:taskID", func(c *gin.Context) {
		taskID := c.Param("taskID")
		result, err := redisClient.GetResult(taskID)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to fetch result: %v", err)})
			return
		}
		if result == nil {
			c.JSON(404, gin.H{"error": "Result not found or still processing"})
			return
		}
		c.JSON(200, result)
	})

	// Start server
	log.Println("API server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

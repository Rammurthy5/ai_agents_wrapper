package main

import (
	"fmt"
	"log"

	"github.com/Rammurthy5/ai_agents_wrapper/internal/facade"
	"github.com/Rammurthy5/ai_agents_wrapper/internal/queue"

	"github.com/gin-gonic/gin"
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

	// Set up gin router
	r := gin.Default()
	r.GET("/getMergedResults", func(c *gin.Context) {
		prompt := c.Query("prompt")
		if prompt == "" {
			c.JSON(400, gin.H{"error": "Missing 'prompt' query parameter"})
			return
		}

		// Enqueue the prompt
		msg := queue.Message{Prompt: prompt}
		if err := rabbit.Publish(msg); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to enqueue request: %v", err)})
			return
		}

		c.JSON(202, gin.H{"message": "Request queued successfully"})
	})

	// Start server
	log.Println("API server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

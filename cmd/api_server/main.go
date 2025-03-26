package main

import (
	"log"

	"github.com/Rammurthy5/ai_agents_wrapper/internal/facade"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := facade.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize facade
	f := facade.NewFacade(cfg)

	// Set up gin router
	r := gin.Default()
	r.GET("/getMergedResults", f.Handler)

	// Start server
	log.Println("API server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

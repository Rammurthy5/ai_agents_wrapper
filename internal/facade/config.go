package facade

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds configuration for AI clients and facade
type Config struct {
	OpenAIKey      string
	OpenAIURL      string
	HuggingFaceKey string
	HuggingFaceURL string
	GeminiKey      string
	GeminiURL      string
	RABBITMQ_URL   string
	Timeout        time.Duration // HTTP client timeout
	MaxRetries     uint          // Max retry attempts for API calls
	RetryDelay     time.Duration // Delay between retries
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if present
	_ = godotenv.Load() // Ignore error if no .env file

	config := &Config{
		OpenAIKey:      os.Getenv("OPENAI_API_KEY"),
		OpenAIURL:      os.Getenv("OPENAI_URL"),
		HuggingFaceKey: os.Getenv("HUGGINGFACE_API_KEY"),
		HuggingFaceURL: os.Getenv("HUGGINGFACE_URL"),
		GeminiKey:      os.Getenv("GEMINI_API_KEY"),
		GeminiURL:      os.Getenv("GEMINI_URL"),
		RABBITMQ_URL:   os.Getenv("RABBITMQ_URL"),
		Timeout:        10 * time.Second, // Default timeout
		MaxRetries:     3,                // Default retries
		RetryDelay:     1 * time.Second,  // Default delay
	}

	// Validate required fields
	if config.OpenAIKey == "" || config.HuggingFaceKey == "" || config.GeminiKey == "" {
		return nil, fmt.Errorf("missing required API keys")
	}
	if config.OpenAIURL == "" {
		config.OpenAIURL = "https://api.openai.com/v1/chat/completions"
	}
	if config.HuggingFaceURL == "" {
		config.HuggingFaceURL = "https://api-inference.huggingface.co/models/mixtral/mixtral-8x7b"
	}
	if config.GeminiURL == "" {
		config.GeminiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent"
	}
	if config.RABBITMQ_URL == "" {
		return nil, fmt.Errorf("missing RabbitMQ URL")
	}

	return config, nil
}

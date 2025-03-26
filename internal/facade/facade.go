package facade

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
)

// Facade provides a unified interface for calling multiple AI APIs
type Facade struct {
	clients []AIClient
}

// NewFacade initializes the Facade with AI clients from config
func NewFacade(cfg *Config) *Facade {
	return &Facade{
		clients: []AIClient{
			NewOpenAIClient(cfg),
			NewHuggingFaceClient(cfg),
			NewGeminiClient(cfg),
		},
	}
}

// GetMergedResults calls all AI APIs concurrently and merges results
func (f *Facade) GetMergedResults(prompt string) MergedApiResponse {
	var wg sync.WaitGroup
	resultsChan := make(chan ApiResponse, len(f.clients))

	// Fork: Launch goroutines for each API call
	for _, client := range f.clients {
		wg.Add(1)
		go func(c AIClient) {
			defer wg.Done()
			resultsChan <- c.Call(prompt)
		}(client)
	}

	// Join: Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var results []ApiResponse
	for resp := range resultsChan {
		results = append(results, resp)
	}

	return MergedApiResponse{Results: results}
}

// Handler is the gin-compatible handler
func (f *Facade) Handler(c *gin.Context) {
	prompt := c.Query("prompt")
	if prompt == "" {
		c.JSON(400, gin.H{"error": "Missing 'prompt' query parameter"})
		return
	}

	result := f.GetMergedResults(prompt)
	for _, r := range result.Results {
		if r.Error != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("%s failed: %v", r.Source, r.Error)})
			return
		}
	}

	c.JSON(200, result)
}

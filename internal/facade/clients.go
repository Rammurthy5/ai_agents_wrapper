package facade

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/sony/gobreaker"
)

// AIClient defines the interface for AI API clients
type AIClient interface {
	Call(prompt string) ApiResponse
	Source() string
}

// Client-specific circuit breaker settings
type breakerClient struct {
	httpClient *http.Client
	apiKey     string
	url        string
	maxRetries uint
	retryDelay time.Duration
	cb         *gobreaker.CircuitBreaker // New: Circuit breaker instance
}

// OpenAIClient implements AIClient for OpenAI
type OpenAIClient struct {
	breakerClient
}

func NewOpenAIClient(cfg *Config) *OpenAIClient {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "OpenAI",
		MaxRequests: 2,                // Half-open state allows 2 requests to test recovery
		Interval:    60 * time.Second, // Reset failure count every 60s in closed state
		Timeout:     30 * time.Second, // Open state lasts 30s before half-open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5 // Trip after 5 consecutive failures
		},
	})
	return &OpenAIClient{
		breakerClient: breakerClient{
			httpClient: &http.Client{Timeout: cfg.Timeout},
			apiKey:     cfg.OpenAIKey,
			url:        cfg.OpenAIURL,
			maxRetries: cfg.MaxRetries,
			retryDelay: cfg.RetryDelay,
			cb:         cb,
		},
	}
}

func (c *OpenAIClient) Call(prompt string) ApiResponse {
	payload := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	resp := callAPI(c.url, c.apiKey, "Bearer", c.Source(), c.maxRetries, c.retryDelay, payload, c.httpClient, c.cb)
	if resp.Error == nil {
		var result map[string]interface{}
		json.Unmarshal([]byte(resp.Message), &result)
		resp.Message = result["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	}
	return resp
}

func (c *OpenAIClient) Source() string {
	return "OpenAI"
}

// HuggingFaceClient implements AIClient for Hugging Face
type HuggingFaceClient struct {
	breakerClient
}

func NewHuggingFaceClient(cfg *Config) *HuggingFaceClient {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "HuggingFace",
		MaxRequests: 2,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	})
	return &HuggingFaceClient{
		breakerClient: breakerClient{
			httpClient: &http.Client{Timeout: cfg.Timeout},
			apiKey:     cfg.HuggingFaceKey,
			url:        cfg.HuggingFaceURL,
			maxRetries: cfg.MaxRetries,
			retryDelay: cfg.RetryDelay,
			cb:         cb,
		},
	}
}

func (c *HuggingFaceClient) Call(prompt string) ApiResponse {
	payload := map[string]string{
		"inputs": prompt,
	}

	resp := callAPI(c.url, c.apiKey, "Bearer", c.Source(), c.maxRetries, c.retryDelay, payload, c.httpClient, c.cb)
	if resp.Error == nil {
		var result []map[string]interface{}
		json.Unmarshal([]byte(resp.Message), &result)
		resp.Message = result[0]["generated_text"].(string)
	}
	return resp
}

func (c *HuggingFaceClient) Source() string {
	return "HuggingFace"
}

// GeminiClient implements AIClient for Google Gemini
type GeminiClient struct {
	breakerClient
}

func NewGeminiClient(cfg *Config) *GeminiClient {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "Gemini",
		MaxRequests: 2,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	})
	return &GeminiClient{
		breakerClient: breakerClient{
			httpClient: &http.Client{Timeout: cfg.Timeout},
			apiKey:     cfg.GeminiKey,
			url:        fmt.Sprintf("%s?key=%s", cfg.GeminiURL, cfg.GeminiKey),
			maxRetries: cfg.MaxRetries,
			retryDelay: cfg.RetryDelay,
			cb:         cb,
		},
	}
}

func (c *GeminiClient) Call(prompt string) ApiResponse {
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}
	resp := callAPI(c.url, "", "", c.Source(), c.maxRetries, c.retryDelay, payload, c.httpClient, c.cb)
	if resp.Error == nil {
		var result map[string]interface{}
		json.Unmarshal([]byte(resp.Message), &result)
		resp.Message = result["candidates"].([]interface{})[0].(map[string]interface{})["content"].(map[string]interface{})["parts"].([]interface{})[0].(map[string]interface{})["text"].(string)
	}
	return resp
}

func (c *GeminiClient) Source() string {
	return "Gemini"
}

// callAPI makes HTTP requests with retries
func callAPI(url, apiKey, authType, source string, maxRetries uint, retryDelay time.Duration,
	payload interface{}, httpClient *http.Client, cb *gobreaker.CircuitBreaker) ApiResponse {
	var apiResp ApiResponse
	err := retry.Do(
		func() error {
			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("marshal error: %v", err)
			}

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
			if err != nil {
				return fmt.Errorf("request error: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			if apiKey != "" && authType != "" {
				req.Header.Set("Authorization", fmt.Sprintf("%s %s", authType, apiKey))
			}

			// Wrap HTTP request with circuit breaker
			httpResp, err := cb.Execute(func() (interface{}, error) {
				resp, err := httpClient.Do(req)
				if err != nil {
					return nil, fmt.Errorf("http error: %v", err)
				}
				return resp, nil
			})

			if err != nil {
				return fmt.Errorf("circuit breaker error: %v", err)
			}

			respObj := httpResp.(*http.Response)
			defer respObj.Body.Close()

			if respObj.StatusCode != http.StatusOK {
				return fmt.Errorf("status %d: %s", respObj.StatusCode, respObj.Status)
			}

			var rawContent bytes.Buffer
			_, err = rawContent.ReadFrom(respObj.Body)
			if err != nil {
				return fmt.Errorf("read error: %v", err)
			}

			apiResp = ApiResponse{Source: source, Message: rawContent.String()}
			return nil
		},
		retry.Attempts(maxRetries),
		retry.Delay(retryDelay),
		retry.RetryIf(func(err error) bool {
			return err != nil && !isPermanentError(err)
		}),
	)

	if err != nil {
		return ApiResponse{Source: source, Error: err}
	}
	return apiResp
}

func isPermanentError(err error) bool {
	return false
}

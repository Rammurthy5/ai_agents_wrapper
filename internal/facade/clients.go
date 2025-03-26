package facade

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
)

// AIClient defines the interface for AI API clients
type AIClient interface {
	Call(prompt string) ApiResponse
	Source() string
}

// OpenAIClient implements AIClient for OpenAI
type OpenAIClient struct {
	httpClient *http.Client
	apiKey     string
	url        string
	maxRetries uint
	retryDelay time.Duration
}

func NewOpenAIClient(cfg *Config) *OpenAIClient {
	return &OpenAIClient{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		apiKey:     cfg.OpenAIKey,
		url:        cfg.OpenAIURL,
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
	}
}

func (c *OpenAIClient) Call(prompt string) ApiResponse {
	payload := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	resp := c.callAPI(c.url, c.apiKey, payload, "Bearer")
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
	httpClient *http.Client
	apiKey     string
	url        string
	maxRetries uint
	retryDelay time.Duration
}

func NewHuggingFaceClient(cfg *Config) *HuggingFaceClient {
	return &HuggingFaceClient{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		apiKey:     cfg.HuggingFaceKey,
		url:        cfg.HuggingFaceURL,
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
	}
}

func (c *HuggingFaceClient) Call(prompt string) ApiResponse {
	payload := map[string]string{
		"inputs": prompt,
	}
	resp := c.callAPI(c.url, c.apiKey, payload, "Bearer")
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
	httpClient *http.Client
	apiKey     string
	url        string
	maxRetries uint
	retryDelay time.Duration
}

func NewGeminiClient(cfg *Config) *GeminiClient {
	return &GeminiClient{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		apiKey:     cfg.GeminiKey,
		url:        fmt.Sprintf("%s?key=%s", cfg.GeminiURL, cfg.GeminiKey),
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
	}
}

func (c *GeminiClient) Call(prompt string) ApiResponse {
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}
	resp := c.callAPI(c.url, "", payload, "")
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
func (c *OpenAIClient) callAPI(url, apiKey string, payload interface{}, authType string) ApiResponse {
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

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("http error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("status %d: %s", resp.StatusCode, resp.Status)
			}

			var rawContent bytes.Buffer
			_, err = rawContent.ReadFrom(resp.Body)
			if err != nil {
				return fmt.Errorf("read error: %v", err)
			}

			apiResp = ApiResponse{Source: c.Source(), Message: rawContent.String()}
			return nil
		},
		retry.Attempts(c.maxRetries),
		retry.Delay(c.retryDelay),
		retry.RetryIf(func(err error) bool {
			return err != nil && !isPermanentError(err)
		}),
	)

	if err != nil {
		return ApiResponse{Source: c.Source(), Error: err}
	}
	return apiResp
}

// callAPI for HuggingFaceClient (to satisfy interface)
func (c *HuggingFaceClient) callAPI(url, apiKey string, payload interface{}, authType string) ApiResponse {
	return (&OpenAIClient{httpClient: c.httpClient, apiKey: c.apiKey, url: c.url, maxRetries: c.maxRetries, retryDelay: c.retryDelay}).callAPI(url, apiKey, payload, authType)
}

// callAPI for GeminiClient (to satisfy interface)
func (c *GeminiClient) callAPI(url, apiKey string, payload interface{}, authType string) ApiResponse {
	return (&OpenAIClient{httpClient: c.httpClient, apiKey: c.apiKey, url: c.url, maxRetries: c.maxRetries, retryDelay: c.retryDelay}).callAPI(url, apiKey, payload, authType)
}

func isPermanentError(err error) bool {
	return false
}

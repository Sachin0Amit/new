package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaClient integrates with Ollama for local LLM inference
type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
	timeout    time.Duration
}

type ollamaRequest struct {
	Model       string        `json:"model"`
	Prompt      string        `json:"prompt"`
	Messages    []ollamaMsg   `json:"messages,omitempty"`
	Stream      bool          `json:"stream"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	NumPredict  *int          `json:"num_predict,omitempty"`
}

type ollamaMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Model              string    `json:"model"`
	CreatedAt          string    `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	SampleCount        int       `json:"sample_count,omitempty"`
	SampleDuration     int64     `json:"sample_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// NewOllamaClient creates a new Ollama-based LLM client
func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for streaming
		},
		timeout: 30 * time.Second,
	}
}

// Health checks if Ollama is available
func (c *OllamaClient) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama not responding: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	return nil
}

// GetModel returns the currently configured model
func (c *OllamaClient) GetModel() string {
	return c.model
}

// Complete sends a chat completion request and returns the response
func (c *OllamaClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Convert messages to Ollama format
	ollamaMsgs := make([]ollamaMsg, 0, len(req.Messages))
	for _, msg := range req.Messages {
		ollamaMsgs = append(ollamaMsgs, ollamaMsg{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	ollamareq := ollamaRequest{
		Model:      c.model,
		Messages:   ollamaMsgs,
		Stream:     false,
		Temperature: parseFloat(req.Temperature),
		TopP:       parseFloat(req.TopP),
		NumPredict: parseInt(req.MaxTokens),
	}

	body, err := json.Marshal(ollamareq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &CompletionResponse{
		Content:    ollamaResp.Response,
		Model:      ollamaResp.Model,
		StopReason: "stop",
		RawResponse: ollamaResp.Response,
		Usage: UsageInfo{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
	}, nil
}

// Stream sends a chat completion request and streams the response
func (c *OllamaClient) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, <-chan error, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)

	// Convert messages to Ollama format
	ollamaMsgs := make([]ollamaMsg, 0, len(req.Messages))
	for _, msg := range req.Messages {
		ollamaMsgs = append(ollamaMsgs, ollamaMsg{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	ollamareq := ollamaRequest{
		Model:      c.model,
		Messages:   ollamaMsgs,
		Stream:     true,
		Temperature: parseFloat(req.Temperature),
		TopP:       parseFloat(req.TopP),
		NumPredict: parseInt(req.MaxTokens),
	}

	body, err := json.Marshal(ollamareq)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("ollama request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		cancel()
		return nil, nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	chunks := make(chan StreamChunk, 10)
	errors := make(chan error, 1)

	go func() {
		defer cancel()
		defer resp.Body.Close()
		defer close(chunks)
		defer close(errors)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var ollamaChunk ollamaResponse
			if err := json.Unmarshal(line, &ollamaChunk); err != nil {
				// Log but don't fail on individual chunk parse errors
				continue
			}

			chunk := StreamChunk{
				Delta:       ollamaChunk.Response,
				Timestamp:   time.Now().UnixNano(),
				RawResponse: ollamaChunk.Response,
			}

			if ollamaChunk.Done {
				chunk.StopReason = "stop"
				chunk.Usage = UsageInfo{
					PromptTokens:     ollamaChunk.PromptEvalCount,
					CompletionTokens: ollamaChunk.EvalCount,
					TotalTokens:      ollamaChunk.PromptEvalCount + ollamaChunk.EvalCount,
				}
			}

			select {
			case chunks <- chunk:
			case <-ctx.Done():
				errors <- ctx.Err()
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errors <- fmt.Errorf("stream reading error: %w", err)
		}
	}()

	return chunks, errors, nil
}

// Helper functions
func parseFloat(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

func parseInt(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

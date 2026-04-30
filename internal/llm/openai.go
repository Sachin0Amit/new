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

// OpenAIClient implements the Client interface for OpenAI-compatible APIs
// with native function/tool calling support.
type OpenAIClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// OpenAI request/response types

type openAIRequest struct {
	Model       string            `json:"model"`
	Messages    []openAIMessage   `json:"messages"`
	Tools       []openAITool      `json:"tools,omitempty"`
	ToolChoice  interface{}       `json:"tool_choice,omitempty"` // "auto", "none", or specific
	Temperature *float64          `json:"temperature,omitempty"`
	MaxTokens   *int              `json:"max_tokens,omitempty"`
	TopP        *float64          `json:"top_p,omitempty"`
	Stream      bool              `json:"stream"`
}

type openAIMessage struct {
	Role       string              `json:"role"`
	Content    string              `json:"content,omitempty"`
	ToolCalls  []openAIToolCall    `json:"tool_calls,omitempty"`
	ToolCallID string              `json:"tool_call_id,omitempty"`
}

type openAITool struct {
	Type     string           `json:"type"` // "function"
	Function openAIFunction   `json:"function"`
}

type openAIFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int           `json:"index"`
		Message      openAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"` // "stop", "tool_calls", "length"
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string           `json:"role,omitempty"`
			Content   string           `json:"content,omitempty"`
			ToolCalls []openAIToolCall  `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// NewOpenAIClient creates a new OpenAI-compatible client with function calling.
func NewOpenAIClient(baseURL, apiKey, model string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// Health checks if the OpenAI API is reachable.
func (c *OpenAIClient) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("OpenAI not responding: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenAI returned status %d", resp.StatusCode)
	}
	return nil
}

// GetModel returns the configured model name.
func (c *OpenAIClient) GetModel() string { return c.model }

// Complete sends a chat completion with optional function/tool calling.
func (c *OpenAIClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	oaiReq := c.buildRequest(req)
	oaiReq.Stream = false

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var oaiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := oaiResp.Choices[0]
	result := &CompletionResponse{
		Content:    choice.Message.Content,
		Model:      oaiResp.Model,
		StopReason: choice.FinishReason,
		Usage: UsageInfo{
			PromptTokens:     oaiResp.Usage.PromptTokens,
			CompletionTokens: oaiResp.Usage.CompletionTokens,
			TotalTokens:      oaiResp.Usage.TotalTokens,
		},
	}

	// Convert tool calls
	for _, tc := range choice.Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: json.RawMessage(tc.Function.Arguments),
		})
	}

	return result, nil
}

// Stream sends a streaming chat completion with function calling support.
func (c *OpenAIClient) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, <-chan error, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)

	oaiReq := c.buildRequest(req)
	oaiReq.Stream = true

	body, err := json.Marshal(oaiReq)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("OpenAI request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		cancel()
		return nil, nil, fmt.Errorf("OpenAI returned status %d: %s", resp.StatusCode, string(bodyBytes))
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
			line := scanner.Text()
			if line == "" || line == "data: [DONE]" {
				continue
			}
			// Strip "data: " prefix
			if len(line) > 6 && line[:6] == "data: " {
				line = line[6:]
			}

			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) == 0 {
				continue
			}

			delta := chunk.Choices[0].Delta
			sc := StreamChunk{
				Delta:     delta.Content,
				Timestamp: time.Now().UnixNano(),
			}

			// Handle tool calls in stream
			for _, tc := range delta.ToolCalls {
				sc.ToolCalls = append(sc.ToolCalls, ToolCall{
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: json.RawMessage(tc.Function.Arguments),
				})
			}

			if chunk.Choices[0].FinishReason != nil {
				sc.StopReason = *chunk.Choices[0].FinishReason
			}

			select {
			case chunks <- sc:
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

// buildRequest converts our generic CompletionRequest to OpenAI format.
func (c *OpenAIClient) buildRequest(req *CompletionRequest) openAIRequest {
	msgs := make([]openAIMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, openAIMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}

	oaiReq := openAIRequest{
		Model:       c.model,
		Messages:    msgs,
		Temperature: parseFloat(req.Temperature),
		TopP:        parseFloat(req.TopP),
		MaxTokens:   parseInt(req.MaxTokens),
	}

	// Convert tool definitions to OpenAI format (function calling)
	for _, tool := range req.Tools {
		oaiReq.Tools = append(oaiReq.Tools, openAITool{
			Type: "function",
			Function: openAIFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		})
	}

	if len(oaiReq.Tools) > 0 {
		oaiReq.ToolChoice = "auto"
	}

	return oaiReq
}

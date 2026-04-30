package llm

import (
	"context"
)

// MessageRole represents the role of a message in conversation
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
	RoleTool      MessageRole = "tool"
)

// Message represents a single message in the conversation
type Message struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
	ToolID  string      `json:"tool_id,omitempty"`
}

// CompletionRequest represents a request to the LLM
type CompletionRequest struct {
	Model       string     `json:"model"`
	Messages    []Message  `json:"messages"`
	Temperature float64    `json:"temperature,omitempty"`
	MaxTokens   int        `json:"max_tokens,omitempty"`
	TopP        float64    `json:"top_p,omitempty"`
	Stream      bool       `json:"stream,omitempty"`
	Tools       []ToolDef  `json:"tools,omitempty"`
}

// CompletionResponse represents a response from the LLM
type CompletionResponse struct {
	Content     string     `json:"content"`
	Model       string     `json:"model"`
	ToolCalls   []ToolCall `json:"tool_calls,omitempty"`
	StopReason  string     `json:"stop_reason"`
	Usage       UsageInfo  `json:"usage"`
	RawResponse string     `json:"raw_response"`
}

// StreamChunk represents a single chunk of streamed response
type StreamChunk struct {
	Delta       string        `json:"delta"`
	ToolCalls   []ToolCall    `json:"tool_calls,omitempty"`
	StopReason  string        `json:"stop_reason,omitempty"`
	Usage       UsageInfo     `json:"usage,omitempty"`
	Timestamp   int64         `json:"timestamp"`
	RawResponse string        `json:"raw_response,omitempty"`
}

// UsageInfo tracks token usage
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ToolDef defines a tool that the LLM can call
type ToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	InputSchema interface{}            `json:"input_schema,omitempty"`
}

// ToolCall represents a tool invocation by the LLM
type ToolCall struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Arguments interface{} `json:"arguments"`
}

// Client defines the interface for LLM operations
type Client interface {
	// Complete sends a completion request and returns the response
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Stream sends a completion request and streams the response
	Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, <-chan error, error)

	// Health checks if the LLM service is available
	Health(ctx context.Context) error

	// GetModel returns the currently configured model
	GetModel() string
}

// StreamReceiver is called for each streamed chunk
type StreamReceiver func(ctx context.Context, chunk StreamChunk) error

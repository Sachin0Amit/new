package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Sachin0Amit/new/internal/llm"
)

// ContextManager manages conversation history with token limits and compression
type ContextManager struct {
	messages      []llm.Message
	maxTokens     int
	compressionRatio float64
	compressor    Compressor
	mu            sync.RWMutex
	tokenCounter  TokenCounter
}

// TokenCounter counts tokens in messages
type TokenCounter interface {
	CountTokens(text string) int
	EstimateTokens(messages []llm.Message) int
}

// Compressor compresses old messages
type Compressor interface {
	Compress(ctx context.Context, messages []llm.Message) (string, error)
}

// NewContextManager creates a new context manager
func NewContextManager(maxTokens int, compressor Compressor, tokenCounter TokenCounter) *ContextManager {
	return &ContextManager{
		messages:      make([]llm.Message, 0),
		maxTokens:     maxTokens,
		compressionRatio: 0.5,
		compressor:    compressor,
		tokenCounter:  tokenCounter,
	}
}

// AddMessage adds a message to the context
func (cm *ContextManager) AddMessage(ctx context.Context, role llm.MessageRole, content string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	msg := llm.Message{
		Role:    role,
		Content: content,
	}

	cm.messages = append(cm.messages, msg)

	// Check if we need to compress
	tokens := cm.tokenCounter.EstimateTokens(cm.messages)
	if tokens > cm.maxTokens {
		return cm.compressContextLocked(ctx)
	}

	return nil
}

// GetMessages returns a copy of the current messages
func (cm *ContextManager) GetMessages() []llm.Message {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	messages := make([]llm.Message, len(cm.messages))
	copy(messages, cm.messages)
	return messages
}

// GetMessageCount returns the number of messages
func (cm *ContextManager) GetMessageCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.messages)
}

// GetEstimatedTokens returns the estimated token count
func (cm *ContextManager) GetEstimatedTokens() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.tokenCounter.EstimateTokens(cm.messages)
}

// Clear clears all messages
func (cm *ContextManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.messages = make([]llm.Message, 0)
}

// compressContextLocked compresses old messages when token limit is exceeded
func (cm *ContextManager) compressContextLocked(ctx context.Context) error {
	if cm.compressor == nil || len(cm.messages) < 4 {
		// If no compressor or too few messages, just trim old messages
		targetTokens := int(float64(cm.maxTokens) * cm.compressionRatio)
		cm.trimToTokens(targetTokens)
		return nil
	}

	// Compress the oldest 50% of messages
	splitIdx := len(cm.messages) / 2
	oldMessages := cm.messages[:splitIdx]
	recentMessages := cm.messages[splitIdx:]

	compressed, err := cm.compressor.Compress(ctx, oldMessages)
	if err != nil {
		// Fall back to trimming if compression fails
		targetTokens := int(float64(cm.maxTokens) * cm.compressionRatio)
		cm.trimToTokens(targetTokens)
		return fmt.Errorf("compression failed, fell back to trimming: %w", err)
	}

	// Replace old messages with compressed summary
	cm.messages = append([]llm.Message{
		{
			Role:    llm.RoleSystem,
			Content: fmt.Sprintf("Previous conversation summary:\n%s", compressed),
		},
	}, recentMessages...)

	return nil
}

// trimToTokens removes messages until we're under the target token count
func (cm *ContextManager) trimToTokens(targetTokens int) {
	// Keep the first system messages and the most recent messages
	systemMessages := 0
	for i, msg := range cm.messages {
		if msg.Role == llm.RoleSystem {
			systemMessages = i + 1
		} else {
			break
		}
	}

	if systemMessages >= len(cm.messages) {
		return // All system messages, don't trim
	}

	// Start from the end and remove messages until we're under the limit
	for len(cm.messages) > systemMessages && cm.tokenCounter.EstimateTokens(cm.messages) > targetTokens {
		cm.messages = append(cm.messages[:systemMessages], cm.messages[systemMessages+1:]...)
	}
}

// SimpleTokenCounter is a basic token counter using character estimation
type SimpleTokenCounter struct {
	avgCharsPerToken float64
}

// NewSimpleTokenCounter creates a new simple token counter
func NewSimpleTokenCounter() *SimpleTokenCounter {
	return &SimpleTokenCounter{
		avgCharsPerToken: 4.0, // Rough estimate
	}
}

// CountTokens estimates tokens in text (rough estimate)
func (stc *SimpleTokenCounter) CountTokens(text string) int {
	return int(float64(len(text)) / stc.avgCharsPerToken)
}

// EstimateTokens estimates tokens in messages
func (stc *SimpleTokenCounter) EstimateTokens(messages []llm.Message) int {
	total := 0
	for _, msg := range messages {
		total += stc.CountTokens(msg.Content) + 4 // Add overhead for role/formatting
	}
	return total
}

// SimpleCompressor compresses messages using extractive summarization
type SimpleCompressor struct {
	maxSummaryTokens int
}

// NewSimpleCompressor creates a new simple compressor
func NewSimpleCompressor(maxTokens int) *SimpleCompressor {
	return &SimpleCompressor{
		maxSummaryTokens: maxTokens,
	}
}

// Compress creates a summary of messages
func (sc *SimpleCompressor) Compress(ctx context.Context, messages []llm.Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	// Simple extraction: take key sentences from messages
	var summary string
	for _, msg := range messages {
		if msg.Role == llm.RoleUser {
			summary += fmt.Sprintf("- User: %s\n", msg.Content[:min(len(msg.Content), 200)])
		} else if msg.Role == llm.RoleAssistant {
			summary += fmt.Sprintf("- Assistant: %s\n", msg.Content[:min(len(msg.Content), 200)])
		}
	}

	return summary, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ConversationTurn represents a single turn in the conversation
type ConversationTurn struct {
	ID            string
	Timestamp     time.Time
	UserMessage   string
	AssistantMsg  string
	ToolCalls     []llm.ToolCall
	Duration      time.Duration
	TokensUsed    int
}

// ConversationHistory tracks the conversation history
type ConversationHistory struct {
	turns []ConversationTurn
	mu    sync.RWMutex
}

// NewConversationHistory creates a new conversation history
func NewConversationHistory() *ConversationHistory {
	return &ConversationHistory{
		turns: make([]ConversationTurn, 0),
	}
}

// AddTurn adds a turn to the history
func (ch *ConversationHistory) AddTurn(turn ConversationTurn) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.turns = append(ch.turns, turn)
}

// GetTurns returns all turns
func (ch *ConversationHistory) GetTurns() []ConversationTurn {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	turns := make([]ConversationTurn, len(ch.turns))
	copy(turns, ch.turns)
	return turns
}

// GetLastN returns the last N turns
func (ch *ConversationHistory) GetLastN(n int) []ConversationTurn {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	if n > len(ch.turns) {
		n = len(ch.turns)
	}
	turns := make([]ConversationTurn, n)
	copy(turns, ch.turns[len(ch.turns)-n:])
	return turns
}

// Clear clears all turns
func (ch *ConversationHistory) Clear() {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.turns = make([]ConversationTurn, 0)
}

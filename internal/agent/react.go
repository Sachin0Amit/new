package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Sachin0Amit/new/internal/llm"
	"github.com/google/uuid"
)

// ReActAgent implements the Reason + Act loop
type ReActAgent struct {
	id                string
	llmClient         llm.Client
	toolExecutor      *ToolExecutor
	contextManager    *ContextManager
	memoryStore       MemoryStore
	maxSteps          int
	mu                sync.RWMutex
	lastStep          *Step
	confidenceThreshold float64
}

// Step represents a single reasoning step
type Step struct {
	Number      int
	Timestamp   time.Time
	Thought     string
	Action      *llm.ToolCall
	Observation string
	Error       error
	Confidence  float64
	Duration    time.Duration
}

// MemoryStore stores agent memories (short, episodic, semantic)
type MemoryStore interface {
	// Store episodic memory
	StoreEpisode(ctx context.Context, key string, data interface{}, ttl time.Duration) error

	// Retrieve episodic memory
	RetrieveEpisode(ctx context.Context, key string) (interface{}, error)

	// Store semantic memory (embeddings)
	StoreSemanticMemory(ctx context.Context, text string, embedding []float32) error

	// Search semantic memory by similarity
	SearchSemanticMemory(ctx context.Context, query string, limit int) ([]interface{}, error)

	// Get short-term memory (in-context)
	GetShortTermMemory(ctx context.Context) ([]llm.Message, error)
}

// NewReActAgent creates a new ReAct agent
func NewReActAgent(
	llmClient llm.Client,
	toolExecutor *ToolExecutor,
	contextManager *ContextManager,
	memoryStore MemoryStore,
) *ReActAgent {
	return &ReActAgent{
		id:                  uuid.New().String(),
		llmClient:           llmClient,
		toolExecutor:        toolExecutor,
		contextManager:      contextManager,
		memoryStore:         memoryStore,
		maxSteps:            8,
		confidenceThreshold: 0.7,
	}
}

// ReasonResult represents the result of a reasoning step
type ReasonResult struct {
	ID              string
	Steps           []Step
	FinalResponse   string
	Success         bool
	TotalDuration   time.Duration
	TokensUsed      int
	LoopDetected    bool
	ToolErrors      int
}

// Reason executes the ReAct loop for a given prompt
func (ra *ReActAgent) Reason(ctx context.Context, prompt string, onStep func(*Step)) (*ReasonResult, error) {
	start := time.Now()
	result := &ReasonResult{
		ID:    uuid.New().String(),
		Steps: make([]Step, 0),
	}

	// Add prompt to context
	if err := ra.contextManager.AddMessage(ctx, llm.RoleUser, prompt); err != nil {
		return result, fmt.Errorf("failed to add message to context: %w", err)
	}

	// Main ReAct loop
	for stepNum := 1; stepNum <= ra.maxSteps; stepNum++ {
		step := Step{
			Number:    stepNum,
			Timestamp: time.Now(),
		}

		stepStart := time.Now()

		// Pre-declare variables used after goto targets
		var action *llm.ToolCall
		var shouldStop bool
		var err error

		// 1. THOUGHT: Get reasoning from LLM
		var thought string
		thought, err = ra.generateThought(ctx)
		if err != nil {
			step.Error = fmt.Errorf("thought generation failed: %w", err)
			step.Observation = ""
			goto recordStep
		}

		step.Thought = thought

		// Check for loop detection (if we've seen this thought before)
		if ra.detectLoop(result.Steps, thought) {
			result.LoopDetected = true
			step.Observation = "LOOP DETECTED: Cannot continue reasoning with the same thought pattern"
			goto recordStep
		}

		// 2. ACTION: Decide on tool use or final answer
		action, shouldStop, err = ra.decideAction(ctx, thought)
		if err != nil {
			step.Error = fmt.Errorf("action decision failed: %w", err)
			goto recordStep
		}

		if shouldStop {
			step.Observation = "FINAL ANSWER"
			result.FinalResponse = thought
			result.Success = true
			goto recordStep
		}

		if action != nil {
			step.Action = action

			// 3. OBSERVATION: Execute tool
			obs, toolErr := ra.executeTool(ctx, *action)
			if toolErr != nil {
				step.Error = toolErr
				step.Observation = fmt.Sprintf("Tool error: %v", toolErr)
				result.ToolErrors++

				// If confidence is low, try self-correction
				if step.Confidence < ra.confidenceThreshold {
					correctedObs, corrErr := ra.selfCorrect(ctx, thought, toolErr)
					if corrErr == nil {
						step.Observation = correctedObs
						step.Error = nil
					}
				}
			} else {
				step.Observation = obs
				step.Error = nil
			}

			// Add tool result to context
			if addErr := ra.contextManager.AddMessage(ctx, llm.RoleTool, step.Observation); addErr != nil {
				step.Error = fmt.Errorf("failed to add tool result: %w", addErr)
			}
		}

	recordStep:
		step.Duration = time.Since(stepStart)
		result.Steps = append(result.Steps, step)

		if onStep != nil {
			onStep(&step)
		}

		// Stop if we have a final answer
		if shouldStop || result.Success || (step.Error != nil && result.LoopDetected) {
			break
		}
	}

	// If we didn't get a final answer, use the last thought
	if !result.Success && len(result.Steps) > 0 {
		result.FinalResponse = result.Steps[len(result.Steps)-1].Thought
	}

	result.TotalDuration = time.Since(start)
	result.Success = result.FinalResponse != ""

	// Add final response to context
	ra.contextManager.AddMessage(ctx, llm.RoleAssistant, result.FinalResponse)

	return result, nil
}

// generateThought generates a thought/reasoning step
func (ra *ReActAgent) generateThought(ctx context.Context) (string, error) {
	messages := ra.contextManager.GetMessages()

	// Add system prompt for reasoning
	systemPrompt := llm.Message{
		Role: llm.RoleSystem,
		Content: `You are a highly intelligent AI assistant. Think step by step.
Format your response as:
THOUGHT: [Your reasoning about what to do next]
ACTION: [Either "FINISH" for a final answer, or use a tool like: <tool name="web_search" args="{\"query\": \"...\"}"/>]
OBSERVATION: [What you observe or learn]`,
	}

	allMessages := append([]llm.Message{systemPrompt}, messages...)

	req := &llm.CompletionRequest{
		Model:       ra.llmClient.GetModel(),
		Messages:    allMessages,
		Temperature: 0.7,
		MaxTokens:   1000,
		Stream:      false,
	}

	resp, err := ra.llmClient.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("llm completion failed: %w", err)
	}

	return resp.Content, nil
}

// decideAction decides whether to use a tool or finish
func (ra *ReActAgent) decideAction(ctx context.Context, thought string) (*llm.ToolCall, bool, error) {
	// Parse the thought to extract action
	if containsString(thought, "FINISH") || containsString(thought, "final answer") {
		return nil, true, nil
	}

	// Look for tool calls in the thought
	parser := NewToolResponseParser()
	toolCalls := parser.ParseToolCalls(thought)

	if len(toolCalls) > 0 {
		return &toolCalls[0], false, nil
	}

	return nil, false, nil
}

// executeTool executes a tool call
func (ra *ReActAgent) executeTool(ctx context.Context, toolCall llm.ToolCall) (string, error) {
	result, err := ra.toolExecutor.Execute(ctx, toolCall)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to serialize result: %w", err)
	}

	return string(resultBytes), nil
}

// selfCorrect attempts to correct an error by trying a different approach
func (ra *ReActAgent) selfCorrect(ctx context.Context, lastThought string, err error) (string, error) {
	// Add error context to messages
	errMsg := fmt.Sprintf("Previous approach failed with error: %v. Try a different approach.", err)
	ra.contextManager.AddMessage(ctx, llm.RoleSystem, errMsg)

	// Generate a new thought with correction
	thought, genErr := ra.generateThought(ctx)
	if genErr != nil {
		return "", genErr
	}

	return thought, nil
}

// detectLoop checks if we're in a reasoning loop
func (ra *ReActAgent) detectLoop(steps []Step, currentThought string) bool {
	// Simple loop detection: if we've seen similar thoughts
	for i := len(steps) - 1; i >= 0 && i > len(steps)-4; i-- {
		if similarity(steps[i].Thought, currentThought) > 0.8 {
			return true
		}
	}
	return false
}

// GetLastStep returns the last reasoning step
func (ra *ReActAgent) GetLastStep() *Step {
	ra.mu.RLock()
	defer ra.mu.RUnlock()
	return ra.lastStep
}

// Helper functions

func containsString(text, substring string) bool {
	for i := 0; i <= len(text)-len(substring); i++ {
		if text[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}

// similarity calculates a simple string similarity (Jaccard)
func similarity(s1, s2 string) float64 {
	// Simple word-based Jaccard similarity
	words1 := splitWords(s1)
	words2 := splitWords(s2)

	intersection := 0
	wordMap := make(map[string]bool)
	for _, w := range words1 {
		wordMap[w] = true
	}
	for _, w := range words2 {
		if wordMap[w] {
			intersection++
		}
	}

	union := len(words1) + len(words2) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func splitWords(s string) []string {
	words := make([]string, 0)
	word := ""
	for _, c := range s {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
			word += string(c)
		} else {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

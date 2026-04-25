package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/Sachin0Amit/new/internal/llm"
)

// Tool defines a tool that the agent can use
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns the tool description
	Description() string

	// Parameters returns the tool's parameter schema
	Parameters() map[string]interface{}

	// Execute executes the tool with the given arguments
	Execute(ctx context.Context, args json.RawMessage) (interface{}, error)
}

// ToolExecutor manages tool registration and execution
type ToolExecutor struct {
	tools map[string]Tool
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool
func (te *ToolExecutor) Register(tool Tool) error {
	if te.tools[tool.Name()] != nil {
		return fmt.Errorf("tool %s already registered", tool.Name())
	}
	te.tools[tool.Name()] = tool
	return nil
}

// GetTools returns the list of available tools as LLM tool definitions
func (te *ToolExecutor) GetTools() []llm.ToolDef {
	tools := make([]llm.ToolDef, 0, len(te.tools))
	for _, tool := range te.tools {
		tools = append(tools, llm.ToolDef{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
		})
	}
	return tools
}

// Execute executes a tool call
func (te *ToolExecutor) Execute(ctx context.Context, toolCall llm.ToolCall) (interface{}, error) {
	tool, ok := te.tools[toolCall.Name]
	if !ok {
		return nil, fmt.Errorf("tool %s not found", toolCall.Name)
	}

	args, err := json.Marshal(toolCall.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	return tool.Execute(ctx, args)
}

// --- Built-in Tools ---

// WebSearchTool searches the web
type WebSearchTool struct {
	searchFn func(ctx context.Context, query string) ([]SearchResult, error)
}

// SearchResult represents a search result
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// NewWebSearchTool creates a new web search tool
func NewWebSearchTool(searchFn func(ctx context.Context, query string) ([]SearchResult, error)) *WebSearchTool {
	return &WebSearchTool{searchFn: searchFn}
}

func (t *WebSearchTool) Name() string {
	return "web_search"
}

func (t *WebSearchTool) Description() string {
	return "Search the web for information. Returns top results with titles, URLs, and snippets."
}

func (t *WebSearchTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results (default 5)",
				"default":     5,
			},
		},
		"required": []string{"query"},
	}
}

func (t *WebSearchTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if req.Limit == 0 {
		req.Limit = 5
	}

	results, err := t.searchFn(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(results) > req.Limit {
		results = results[:req.Limit]
	}

	return map[string]interface{}{
		"results": results,
		"query":   req.Query,
		"count":   len(results),
	}, nil
}

// ReadFileTool reads a file from the filesystem
type ReadFileTool struct {
	readFn func(ctx context.Context, path string) (string, error)
}

// NewReadFileTool creates a new read file tool
func NewReadFileTool(readFn func(ctx context.Context, path string) (string, error)) *ReadFileTool {
	return &ReadFileTool{readFn: readFn}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file. Path should be relative to the current working directory."
}

func (t *ReadFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to read",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	content, err := t.readFn(ctx, req.Path)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return map[string]interface{}{
		"path":    req.Path,
		"content": content,
		"size":    len(content),
	}, nil
}

// MathSolverTool solves mathematical expressions
type MathSolverTool struct {
	solveFn func(ctx context.Context, expr string, op string) (string, error)
}

// NewMathSolverTool creates a new math solver tool
func NewMathSolverTool(solveFn func(ctx context.Context, expr string, op string) (string, error)) *MathSolverTool {
	return &MathSolverTool{solveFn: solveFn}
}

func (t *MathSolverTool) Name() string {
	return "solve_math"
}

func (t *MathSolverTool) Description() string {
	return "Solve mathematical expressions, derivatives, integrals, and matrix operations."
}

func (t *MathSolverTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"description": "The mathematical expression to solve",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The operation: 'simplify', 'solve', 'diff', 'integrate', 'solve_linear'",
				"enum":        []string{"simplify", "solve", "diff", "integrate", "solve_linear"},
				"default":     "simplify",
			},
		},
		"required": []string{"expression"},
	}
}

func (t *MathSolverTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req struct {
		Expression string `json:"expression"`
		Operation  string `json:"operation"`
	}
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if req.Operation == "" {
		req.Operation = "simplify"
	}

	result, err := t.solveFn(ctx, req.Expression, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("solve failed: %w", err)
	}

	return map[string]interface{}{
		"expression": req.Expression,
		"operation":  req.Operation,
		"result":     result,
	}, nil
}

// CodeExecutionTool executes code (sandboxed)
type CodeExecutionTool struct {
	executeFn func(ctx context.Context, code, language string) (string, error)
}

// NewCodeExecutionTool creates a new code execution tool
func NewCodeExecutionTool(executeFn func(ctx context.Context, code, language string) (string, error)) *CodeExecutionTool {
	return &CodeExecutionTool{executeFn: executeFn}
}

func (t *CodeExecutionTool) Name() string {
	return "run_code"
}

func (t *CodeExecutionTool) Description() string {
	return "Execute code in a sandboxed environment. Supports Python, JavaScript, and Bash."
}

func (t *CodeExecutionTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":        "string",
				"description": "The code to execute",
			},
			"language": map[string]interface{}{
				"type":        "string",
				"description": "Programming language: 'python', 'javascript', 'bash'",
				"enum":        []string{"python", "javascript", "bash"},
				"default":     "python",
			},
		},
		"required": []string{"code"},
	}
}

func (t *CodeExecutionTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req struct {
		Code     string `json:"code"`
		Language string `json:"language"`
	}
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if req.Language == "" {
		req.Language = "python"
	}

	output, err := t.executeFn(ctx, req.Code, req.Language)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	return map[string]interface{}{
		"language": req.Language,
		"output":   output,
		"success":  err == nil,
	}, nil
}

// ToolResponseParser parses tool calls from LLM responses
type ToolResponseParser struct {
	toolCallRegex *regexp.Regexp
}

// NewToolResponseParser creates a new parser
func NewToolResponseParser() *ToolResponseParser {
	// Pattern: <tool name="..." args="..." />
	pattern := regexp.MustCompile(`<tool\s+name="([^"]+)"\s+args="([^"]+)"\s*/>`)
	return &ToolResponseParser{toolCallRegex: pattern}
}

// ParseToolCalls extracts tool calls from a response text
func (p *ToolResponseParser) ParseToolCalls(text string) []llm.ToolCall {
	matches := p.toolCallRegex.FindAllStringSubmatch(text, -1)
	calls := make([]llm.ToolCall, 0)
	for _, match := range matches {
		if len(match) >= 3 {
			call := llm.ToolCall{
				ID:        fmt.Sprintf("call_%d", time.Now().UnixNano()),
				Name:      match[1],
				Arguments: json.RawMessage(match[2]),
			}
			calls = append(calls, call)
		}
	}
	return calls
}

// FormatToolCall formats a tool call for display
func FormatToolCall(call llm.ToolCall) string {
	argsStr := fmt.Sprintf("%v", call.Arguments)
	return fmt.Sprintf("<tool name=\"%s\" args=\"%s\" />", call.Name, argsStr)
}

// FormatToolResult formats a tool result for the context
func FormatToolResult(call llm.ToolCall, result interface{}, err error) string {
	if err != nil {
		return fmt.Sprintf("<tool_result name=\"%s\" error=\"%v\" />", call.Name, err)
	}
	resultStr := fmt.Sprintf("%v", result)
	return fmt.Sprintf("<tool_result name=\"%s\">%s</tool_result>", call.Name, resultStr)
}

package titan

/*
#cgo LDFLAGS: -L. -ltitan -lstdc++ -lm
#include <stdlib.h>
#include <stdint.h>
#include "../../cpp/engine/titan_engine.h"
*/
import "C"
import (
	"context"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/Sachin0Amit/new/internal/models"
)

// NeuralEngine provides access to the Sovereign Neural Architecture via CGo.
type NeuralEngine struct {
	contexts  map[string]C.TitanNeuralContext
	tier      string // Current/default tier
	modelInfo string
	mu        sync.RWMutex
}

// NeuralResult holds the output of a neural generation.
type NeuralResult struct {
	Text            string
	TokenIDs        []int
	TokenCount      int
	TokensPerSecond float64
	LatencyMS       float64
	MemoryUsedBytes int64
	ModelInfo       string
}

// NewNeuralEngine creates a new neural inference engine.
// tier: "local" (CPU, ~2GB), "mid" (24GB GPU), "max" (multi-node)
func NewNeuralEngine(tier string) *NeuralEngine {
	ne := &NeuralEngine{
		contexts: make(map[string]C.TitanNeuralContext),
		tier:     tier,
	}

	cTier := C.CString(tier)
	defer C.free(unsafe.Pointer(cTier))

	ctx := C.titan_neural_init(cTier)
	ne.contexts[tier] = ctx

	// Cache model info for default tier
	info := C.titan_neural_model_info(ctx)
	ne.modelInfo = C.GoString(info)

	return ne
}

// getContext retrieves a context for the given tier, initializing it if necessary.
func (ne *NeuralEngine) getContext(tier string) (C.TitanNeuralContext, string) {
	ne.mu.RLock()
	ctx, ok := ne.contexts[tier]
	info := ne.modelInfo
	ne.mu.RUnlock()

	if ok {
		return ctx, info
	}

	// Initialize new tier context
	ne.mu.Lock()
	defer ne.mu.Unlock()

	// Double-check after lock
	if ctx, ok = ne.contexts[tier]; ok {
		return ctx, ne.modelInfo
	}

	fmt.Printf("[NeuralBridge] Initializing new specialized tier: %s\n", tier)
	cTier := C.CString(tier)
	defer C.free(unsafe.Pointer(cTier))

	newCtx := C.titan_neural_init(cTier)
	if newCtx == nil {
		fmt.Printf("[NeuralBridge] Warning: failed to initialize tier %s, falling back to default\n", tier)
		return ne.contexts[ne.tier], ne.modelInfo
	}

	ne.contexts[tier] = newCtx
	
	// Update model info if this is the first switch or we want to track it
	infoC := C.titan_neural_model_info(newCtx)
	info = C.GoString(infoC)

	return newCtx, info
}

// Generate runs auto-regressive text generation using a specific context.
func (ne *NeuralEngine) Generate(ctx context.Context, tier string, prompt string, maxTokens int, temperature float32) (*NeuralResult, error) {
	tCtx, _ := ne.getContext(tier)
	
	// We still use a lock for the generation call itself to be safe, 
	// though ideally we'd have a pool of contexts per tier for true parallelism.
	ne.mu.Lock()
	defer ne.mu.Unlock()

	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	cResult := C.titan_neural_generate(
		tCtx, cPrompt,
		C.int(maxTokens),
		C.float(temperature),
	)
	defer C.titan_neural_free_result(cResult)

	result := &NeuralResult{
		Text:            C.GoString(cResult.text),
		TokenCount:      int(cResult.token_count),
		TokensPerSecond: float64(cResult.tokens_per_sec),
		LatencyMS:       float64(cResult.latency_ms),
		MemoryUsedBytes: int64(cResult.memory_used),
	}
	if cResult.model_info != nil {
		result.ModelInfo = C.GoString(cResult.model_info)
	}

	// Copy token IDs
	if cResult.token_ids != nil && cResult.token_count > 0 {
		result.TokenIDs = make([]int, cResult.token_count)
		cTokens := (*[1 << 20]C.int)(unsafe.Pointer(cResult.token_ids))
		for i := 0; i < int(cResult.token_count); i++ {
			result.TokenIDs[i] = int(cTokens[i])
		}
	}

	return result, nil
}

// Tokenize converts text to token IDs.
func (ne *NeuralEngine) Tokenize(text string) []int {
	ne.mu.RLock()
	defer ne.mu.RUnlock()
	
	// Use default tier for tokenization (usually same across tiers)
	ctx := ne.contexts[ne.tier]

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	maxIDs := 4096
	ids := make([]C.int, maxIDs)
	count := C.titan_tokenize(ctx, cText, &ids[0], C.int(maxIDs))

	result := make([]int, int(count))
	for i := 0; i < int(count); i++ {
		result[i] = int(ids[i])
	}
	return result
}

// Detokenize converts token IDs back to text.
func (ne *NeuralEngine) Detokenize(ids []int) string {
	ne.mu.RLock()
	defer ne.mu.RUnlock()
	
	ctx := ne.contexts[ne.tier]

	cIDs := make([]C.int, len(ids))
	for i, id := range ids {
		cIDs[i] = C.int(id)
	}

	maxLen := 8192
	buf := make([]byte, maxLen)
	C.titan_detokenize(ctx, &cIDs[0], C.int(len(ids)),
		(*C.char)(unsafe.Pointer(&buf[0])), C.int(maxLen))

	return string(buf[:cStringLen(buf)])
}

// ModelInfo returns a human-readable model description.
func (ne *NeuralEngine) ModelInfo() string {
	ne.mu.RLock()
	defer ne.mu.RUnlock()
	return ne.modelInfo
}

// CacheMemory returns estimated KV cache memory in bytes.
func (ne *NeuralEngine) CacheMemory() int64 {
	ne.mu.RLock()
	defer ne.mu.RUnlock()
	return int64(C.titan_neural_cache_memory(ne.contexts[ne.tier]))
}

// Close releases the neural engine resources.
func (ne *NeuralEngine) Close() {
	ne.mu.Lock()
	defer ne.mu.Unlock()
	for tier, ctx := range ne.contexts {
		if ctx != nil {
			C.titan_neural_free(ctx)
			ne.contexts[tier] = nil
		}
	}
}

// Infer implements the InferenceEngine interface for language tasks.
func (ne *NeuralEngine) Infer(ctx context.Context, payload map[string]interface{}) (*models.TaskResult, error) {
	prompt, _ := payload["prompt"].(string)
	if prompt == "" {
		prompt = fmt.Sprintf("%v", payload)
	}

	// Dynamic tier switching (cached)
	requestedTier, _ := payload["tier"].(string)
	if requestedTier == "" {
		requestedTier = ne.tier
	}
	
	fmt.Printf("[NeuralBridge] Processing request on tier: '%s'\n", requestedTier)
	
	maxTokens := 256
	if mt, ok := payload["max_tokens"].(int); ok {
		maxTokens = mt
	}
	temperature := float32(0.7)
	if t, ok := payload["temperature"].(float64); ok {
		temperature = float32(t)
	}

	result, err := ne.Generate(ctx, requestedTier, prompt, maxTokens, temperature)
	if err != nil {
		return nil, err
	}

	return &models.TaskResult{
		Data: map[string]interface{}{
			"output":       result.Text,
			"model":        result.ModelInfo,
			"token_count":  result.TokenCount,
			"architecture": result.ModelInfo,
		},
		Metrics: models.InferenceMetrics{
			TokensPerSecond: result.TokensPerSecond,
			MemoryUsedBytes: result.MemoryUsedBytes,
			LatencyMS:       int64(result.LatencyMS),
			HardwareAgent:   ne.tier,
		},
		Completed: time.Now(),
	}, nil
}

// cStringLen finds the null terminator in a byte slice.
func cStringLen(b []byte) int {
	for i, v := range b {
		if v == 0 {
			return i
		}
	}
	return len(b)
}

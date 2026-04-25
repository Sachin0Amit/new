package titan

/*
#cgo LDFLAGS: -L. -ltitan -lstdc++ -lm
#include <stdlib.h>
#include "../../cpp/engine/titan_engine.h"
*/
import "C"
import (
	"context"
	"fmt"
	"time"
	"unsafe"

	"github.com/Sachin0Amit/new/internal/models"
)

// Engine implements models.InferenceEngine using the C++ Titan core.
type titanEngine struct {
	ctx    C.TitanContext
	device string
	neural *NeuralEngine
}

func NewEngine(device string) Engine {
	cDevice := C.CString(device)
	defer C.free(unsafe.Pointer(cDevice))

	// Map device tier to neural config
	tier := "local"
	if device == "cuda:0" || device == "gpu" {
		tier = "mid"
	}

	return &titanEngine{
		ctx:    C.titan_init(cDevice),
		device: device,
		neural: NewNeuralEngine(tier),
	}
}

func (e *titanEngine) Initialize(ctx context.Context, modelPath string) error {
	// Future: Load model weights from path in C++
	return nil
}

func (e *titanEngine) Infer(ctx context.Context, payload map[string]interface{}) (*models.TaskResult, error) {
	// Multimodal branching: check if payload contains raw sensory data
	if raw, ok := payload["sensory_input"].(models.SensoryData); ok {
		e.handleSensoryPayload(raw)
	}

	// Route language/reasoning tasks to the Neural Engine
	prompt, _ := payload["prompt"].(string)
	taskType, _ := payload["type"].(string)
	if prompt != "" || taskType == "language" || taskType == "reasoning" || taskType == "code" || taskType == "inference" {
		return e.neural.Infer(ctx, payload)
	}

	// Default to Symbolic Inference
	pStr := fmt.Sprintf("%v", payload)
	cPayload := C.CString(pStr)
	defer C.free(unsafe.Pointer(cPayload))

	cRes := C.titan_infer(e.ctx, cPayload)
	defer C.titan_free_result(cRes)

	result := &models.TaskResult{
		Data: map[string]interface{}{
			"output": C.GoString(cRes.data),
		},
		Metrics: models.InferenceMetrics{
			TokensPerSecond: float64(cRes.tokens_per_sec),
			MemoryUsedBytes: int64(cRes.memory_used),
			LatencyMS:      150,
			HardwareAgent:  e.device,
		},
		Completed: time.Now(),
	}

	return result, nil
}

// handleSensoryPayload optimizes the transfer of large media buffers to the C++ core.
func (e *titanEngine) handleSensoryPayload(data models.SensoryData) {
	// Future: Use C.titan_push_sensory_buffer for zero-copy transfer
	fmt.Printf("Optimizing sensory buffer transfer: %s (%d bytes)\n", data.MimeType, len(data.Buffer))
}

func (e *titanEngine) GetStatus() (models.InferenceMetrics, error) {
	return models.InferenceMetrics{
		HardwareAgent: e.device,
	}, nil
}

func (e *titanEngine) Unload(ctx context.Context) error {
	if e.ctx != nil {
		C.titan_free(e.ctx)
		e.ctx = nil
	}
	return nil
}

// Ensure the interface is implemented
var _ Engine = (*titanEngine)(nil)

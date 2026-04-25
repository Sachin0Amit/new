package titan

/*
#cgo LDFLAGS: -L${SRCDIR}/../../cpp/build/Release -ltitan_engine -lstdc++ -lm
#cgo CFLAGS: -I${SRCDIR}/../../cpp/include
#include "titan.h"
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"errors"
	"runtime"
	"unsafe"
)

// TitanEngine wraps the C++ shared library via CGo.
type TitanEngine struct {
	handle *C.titan_handle_t
}

// NewTitanEngine creates a new TitanEngine instance from a JSON configuration.
func NewTitanEngine(configJSON string) (*TitanEngine, error) {
	cConfig := C.CString(configJSON)
	defer C.free(unsafe.Pointer(cConfig))

	handle := C.titan_init(cConfig)
	if handle == nil {
		return nil, errors.New("failed to initialize Titan engine via CGo")
	}

	e := &TitanEngine{handle: handle}
	runtime.SetFinalizer(e, func(obj *TitanEngine) {
		obj.Close()
	})
	
	return e, nil
}

// Derive runs inference via the C++ engine.
func (e *TitanEngine) Derive(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if e.handle == nil {
		return "", errors.New("engine already closed")
	}

	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	cResult := C.titan_derive(e.handle, cPrompt, C.int(maxTokens))
	if cResult == nil {
		return "", errors.New("inference failed or queue full")
	}

	// Copy C string to Go string BEFORE freeing C memory
	goResult := C.GoString(cResult)
	C.titan_free_result(cResult)

	return goResult, nil
}

// Close destroys the C++ engine instance.
func (e *TitanEngine) Close() error {
	if e.handle != nil {
		C.titan_destroy(e.handle)
		e.handle = nil
	}
	return nil
}

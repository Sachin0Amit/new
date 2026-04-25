package titan

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnginePool(t *testing.T) {
	mockFactory := func() (Engine, error) {
		return &FallbackEngine{URL: "http://mock"}, nil
	}

	pool, err := NewPooledEngine(2, mockFactory)
	assert.NoError(t, err)
	defer pool.Close()

	t.Run("Concurrent Acquisition", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Acquire all
		engine1, err := pool.pool.Acquire(ctx)
		assert.NoError(t, err)
		
		engine2, err := pool.pool.Acquire(ctx)
		assert.NoError(t, err)

		// Third should timeout
		_, err = pool.pool.Acquire(ctx)
		assert.Error(t, err)

		// Release one and re-acquire
		pool.pool.Release(engine1)
		engine3, err := pool.pool.Acquire(ctx)
		assert.NoError(t, err)
		assert.Equal(t, engine1, engine3)

		pool.pool.Release(engine2)
		pool.pool.Release(engine3)
	})
}

func TestFallback(t *testing.T) {
	engine, err := LoadEngine(context.Background(), "{}", "http://fallback.api")
	assert.NoError(t, err)
	assert.NotNil(t, engine)

	// Since we likely don't have the shared lib in the build path, 
	// it should have fallen back.
	_, isFallback := engine.(*FallbackEngine)
	if !isFallback {
		_, isCGo := engine.(*CGoEngine)
		assert.True(t, isCGo, "Engine must be either CGo or Fallback")
	}
}

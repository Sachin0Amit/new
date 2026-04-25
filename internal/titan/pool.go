package titan

import (
	"context"
	"fmt"
)

type EnginePool struct {
	engines chan Engine
}

func NewEnginePool(size int, factory func() (Engine, error)) (*EnginePool, error) {
	pool := &EnginePool{
		engines: make(chan Engine, size),
	}

	for i := 0; i < size; i++ {
		engine, err := factory()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize engine %d: %w", i, err)
		}
		pool.engines <- engine
	}

	return pool, nil
}

func (p *EnginePool) Acquire(ctx context.Context) (Engine, error) {
	select {
	case engine := <-p.engines:
		return engine, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *EnginePool) Release(engine Engine) {
	p.engines <- engine
}

func (p *EnginePool) Close() error {
	close(p.engines)
	for engine := range p.engines {
		engine.Close()
	}
	return nil
}

// PooledEngine wraps an Engine from a pool to automatically release it.
type PooledEngine struct {
	pool *EnginePool
}

func NewPooledEngine(size int, factory func() (Engine, error)) (*PooledEngine, error) {
	p, err := NewEnginePool(size, factory)
	if err != nil {
		return nil, err
	}
	return &PooledEngine{pool: p}, nil
}

func (pe *PooledEngine) Derive(ctx context.Context, prompt string, maxTokens int) (string, error) {
	engine, err := pe.pool.Acquire(ctx)
	if err != nil {
		return "", err
	}
	defer pe.pool.Release(engine)

	return engine.Derive(ctx, prompt, maxTokens)
}

func (pe *PooledEngine) Close() error {
	return pe.pool.Close()
}

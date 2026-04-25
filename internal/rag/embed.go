package rag

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Embedder struct {
	Endpoint string
	Model    string
	cache    sync.Map
	client   *http.Client
}

func NewEmbedder(endpoint, model string) *Embedder {
	return &Embedder{
		Endpoint: endpoint,
		Model:    model,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (e *Embedder) Embed(text string) ([]float32, error) {
	// Check cache
	hash := sha256.Sum256([]byte(text))
	key := hex.EncodeToString(hash[:])
	if val, ok := e.cache.Load(key); ok {
		return val.([]float32), nil
	}

	// Retry loop
	var lastErr error
	for i := 0; i < 3; i++ {
		vec, err := e.callOllama(text)
		if err == nil {
			e.cache.Store(key, vec)
			return vec, nil
		}
		lastErr = err
		time.Sleep(time.Duration(1<<i) * 100 * time.Millisecond) // Exponential backoff
	}
	
	return nil, fmt.Errorf("ollama embedding failed after 3 retries: %w", lastErr)
}

type ollamaReq struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
}

type ollamaRes struct {
	Embedding []float32 `json:"embedding"`
}

func (e *Embedder) callOllama(text string) ([]float32, error) {
	reqBody, _ := json.Marshal(ollamaReq{
		Model:  e.Model,
		Prompt: text,
	})

	resp, err := e.client.Post(e.Endpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var res ollamaRes
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Embedding, nil
}

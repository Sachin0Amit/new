package titan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type FallbackEngine struct {
	URL string
}

type openAIReq struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type openAIRes struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func (e *FallbackEngine) Derive(ctx context.Context, prompt string, maxTokens int) (string, error) {
	reqBody, _ := json.Marshal(openAIReq{
		Model:     "sovereign-fallback",
		Prompt:    prompt,
		MaxTokens: maxTokens,
	})

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", e.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fallback request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fallback returned status %d", resp.StatusCode)
	}

	var res openAIRes
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("fallback returned no choices")
	}

	return res.Choices[0].Text, nil
}

func (e *FallbackEngine) Close() error {
	return nil
}

// LoadEngine attempts to load the CGo engine, falling back to HTTP if it fails.
func LoadEngine(ctx context.Context, configJSON string, fallbackURL string) (Engine, error) {
	// 1. Try CGo (local C++ shared lib)
	engine, err := NewTitanEngine(configJSON)
	if err == nil {
		fmt.Println("🚀 Titan C++ Engine loaded successfully.")
		return engine, nil
	}

	// 2. Fallback to Remote API
	fmt.Printf("⚠️  Titan C++ Engine load failed (%v). Falling back to: %s\n", err, fallbackURL)
	return &FallbackEngine{URL: fallbackURL}, nil
}

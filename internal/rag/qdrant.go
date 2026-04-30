package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// QdrantClient implements vector search over a Qdrant collection.
// This provides a production-grade alternative to the in-memory HNSW index.
type QdrantClient struct {
	baseURL       string
	collectionName string
	httpClient    *http.Client
	vectorSize    int
}

// QdrantPoint represents a point in the Qdrant collection.
type QdrantPoint struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// QdrantSearchResult represents a single search hit.
type QdrantSearchResult struct {
	ID      string                 `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

// NewQdrantClient creates a client for the given Qdrant instance and collection.
func NewQdrantClient(baseURL, collection string, vectorSize int) *QdrantClient {
	if baseURL == "" {
		baseURL = "http://localhost:6333"
	}
	return &QdrantClient{
		baseURL:        baseURL,
		collectionName: collection,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		vectorSize:     vectorSize,
	}
}

// EnsureCollection creates the collection if it doesn't exist.
func (q *QdrantClient) EnsureCollection(ctx context.Context) error {
	body := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     q.vectorSize,
			"distance": "Cosine",
		},
	}
	data, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/collections/%s", q.baseURL, q.collectionName)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("qdrant not reachable: %w", err)
	}
	defer resp.Body.Close()

	// 409 = already exists → that's fine
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: %d %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// Upsert inserts or updates a batch of points.
func (q *QdrantClient) Upsert(ctx context.Context, points []QdrantPoint) error {
	body := map[string]interface{}{
		"points": points,
	}
	data, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/collections/%s/points", q.baseURL, q.collectionName)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upsert failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upsert failed: %d %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// Search performs a nearest-neighbor vector search.
func (q *QdrantClient) Search(ctx context.Context, queryVector []float32, limit int) ([]QdrantSearchResult, error) {
	body := map[string]interface{}{
		"vector":      queryVector,
		"limit":       limit,
		"with_payload": true,
	}
	data, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/collections/%s/points/search", q.baseURL, q.collectionName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %d %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Result []QdrantSearchResult `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	return result.Result, nil
}

// Delete removes points by IDs.
func (q *QdrantClient) Delete(ctx context.Context, ids []string) error {
	body := map[string]interface{}{
		"points": ids,
	}
	data, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/collections/%s/points/delete", q.baseURL, q.collectionName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %d %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// Health checks if Qdrant is reachable.
func (q *QdrantClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", q.baseURL+"/healthz", nil)
	if err != nil {
		return err
	}
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("qdrant not reachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qdrant returned %d", resp.StatusCode)
	}
	return nil
}

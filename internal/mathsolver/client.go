package mathsolver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var ErrMathSolverUnavailable = errors.New("math solver service is currently unavailable (circuit open)")

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	
	// Circuit Breaker State
	mu           sync.Mutex
	failures     int
	lastFailure  time.Time
	circuitOpen  bool
}

type SolveRequest struct {
	Expression string                 `json:"expression,omitempty"`
	Matrix     [][]interface{}        `json:"matrix,omitempty"`
	Variables  []string               `json:"variables"`
	Options    map[string]interface{} `json:"options"`
}

type SolveResult struct {
	Result            string   `json:"result"`
	LaTeX             string   `json:"latex"`
	Steps             []string `json:"steps"`
	ComputationTimeMS int      `json:"computation_time_ms"`
}

func NewClient(url string) *Client {
	return &Client{
		BaseURL: url,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Solve(ctx context.Context, endpoint string, req SolveRequest) (*SolveResult, error) {
	c.mu.Lock()
	if c.circuitOpen {
		if time.Since(c.lastFailure) > 30*time.Second {
			c.circuitOpen = false
			c.failures = 0
		} else {
			c.mu.Unlock()
			return nil, ErrMathSolverUnavailable
		}
	}
	c.mu.Unlock()

	var result SolveResult
	var lastErr error

	for i := 0; i < 3; i++ {
		lastErr = c.doRequest(ctx, endpoint, req, &result)
		if lastErr == nil {
			return &result, nil
		}
		
		// Retry only on 503 or transient network errors
		time.Sleep(time.Duration(i+1) * 200 * time.Millisecond)
	}

	c.recordFailure()
	return nil, lastErr
}

func (c *Client) doRequest(ctx context.Context, endpoint string, reqData SolveRequest, target interface{}) error {
	body, _ := json.Marshal(reqData)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("service unavailable: %d", resp.StatusCode)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("solver error: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) recordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures++
	c.lastFailure = time.Now()
	if c.failures >= 5 {
		c.circuitOpen = true
	}
}

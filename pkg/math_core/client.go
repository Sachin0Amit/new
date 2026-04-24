package math_core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MathRequest struct {
	Input string `json:"input"`
}

type MathResponse struct {
	ProblemType string `json:"problem_type"`
	Solution    string `json:"solution"`
	FinalAnswer string `json:"final_answer"`
	Steps       []struct {
		Description string `json:"description"`
		Expression  string `json:"expression"`
	} `json:"steps"`
	Plots string `json:"plots"`
}

type Client struct {
	URL    string
	Client *http.Client
}

func NewClient(url string) *Client {
	return &Client{
		URL: url,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Solve(input string) (*MathResponse, error) {
	reqBody, err := json.Marshal(MathRequest{Input: input})
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post(c.URL+"/solve", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("math service returned status: %d", resp.StatusCode)
	}

	var mathResp MathResponse
	if err := json.NewDecoder(resp.Body).Decode(&mathResp); err != nil {
		return nil, err
	}

	return &mathResp, nil
}

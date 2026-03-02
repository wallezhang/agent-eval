// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func init() {
	Register("anthropic", func(config map[string]any) (Client, error) {
		return newAnthropicClient(config)
	})
}

type anthropicClient struct {
	apiKey    string
	baseURL   string
	model     string
	maxTokens int
	client    *http.Client
}

func newAnthropicClient(config map[string]any) (*anthropicClient, error) {
	c := &anthropicClient{
		baseURL:   "https://api.anthropic.com",
		model:     "claude-sonnet-4-20250514",
		maxTokens: 4096,
		client:    &http.Client{Timeout: 120 * time.Second},
	}

	if key, ok := config["api_key"].(string); ok {
		c.apiKey = key
	} else {
		return nil, fmt.Errorf("anthropic llm client: api_key is required")
	}

	if url, ok := config["base_url"].(string); ok {
		c.baseURL = url
	}
	if m, ok := config["model"].(string); ok {
		c.model = m
	}
	if mt, ok := config["max_tokens"].(float64); ok {
		c.maxTokens = int(mt)
	}

	return c, nil
}

func (c *anthropicClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	messages := make([]map[string]string, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	maxTokens := c.maxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	body := map[string]any{
		"model":       c.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": req.Temperature,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	var text string
	for _, block := range result.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	return &CompletionResponse{
		Content: text,
		Usage: Usage{
			InputTokens:  result.Usage.InputTokens,
			OutputTokens: result.Usage.OutputTokens,
		},
	}, nil
}

func (c *anthropicClient) Close() error {
	c.client.CloseIdleConnections()
	return nil
}

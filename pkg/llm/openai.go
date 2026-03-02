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
	Register("openai", func(config map[string]any) (Client, error) {
		return newOpenAIClient(config)
	})
}

type openAIClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func newOpenAIClient(config map[string]any) (*openAIClient, error) {
	c := &openAIClient{
		baseURL: "https://api.openai.com/v1",
		model:   "gpt-4",
		client:  &http.Client{Timeout: 120 * time.Second},
	}

	if key, ok := config["api_key"].(string); ok {
		c.apiKey = key
	} else {
		return nil, fmt.Errorf("openai llm client: api_key is required")
	}

	if url, ok := config["base_url"].(string); ok {
		c.baseURL = url
	}
	if m, ok := config["model"].(string); ok {
		c.model = m
	}

	return c, nil
}

func (c *openAIClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	messages := make([]map[string]string, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	body := map[string]any{
		"model":       c.model,
		"messages":    messages,
		"temperature": req.Temperature,
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

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
		return nil, fmt.Errorf("OpenAI API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &CompletionResponse{
		Content: result.Choices[0].Message.Content,
		Usage: Usage{
			InputTokens:  result.Usage.PromptTokens,
			OutputTokens: result.Usage.CompletionTokens,
		},
	}, nil
}

func (c *openAIClient) Close() error {
	c.client.CloseIdleConnections()
	return nil
}

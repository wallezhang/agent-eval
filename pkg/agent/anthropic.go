// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func init() {
	Register("anthropic", func(config map[string]any) (Agent, error) {
		return newAnthropicAgent(config)
	})
}

type anthropicAgent struct {
	apiKey      string
	baseURL     string
	modelName   string
	temperature float64
	maxTokens   int
	client      *http.Client
}

func newAnthropicAgent(config map[string]any) (*anthropicAgent, error) {
	a := &anthropicAgent{
		baseURL:     "https://api.anthropic.com",
		modelName:   "claude-sonnet-4-20250514",
		temperature: 0.0,
		maxTokens:   4096,
		client:      &http.Client{Timeout: 120 * time.Second},
	}

	if key, ok := config["api_key"].(string); ok {
		a.apiKey = key
	} else {
		return nil, fmt.Errorf("anthropic agent: api_key is required")
	}

	if url, ok := config["base_url"].(string); ok {
		a.baseURL = url
	}
	if m, ok := config["model"].(string); ok {
		a.modelName = m
	}
	if t, ok := config["temperature"].(float64); ok {
		a.temperature = t
	}
	if mt, ok := config["max_tokens"].(int); ok {
		a.maxTokens = mt
	}
	// YAML unmarshals integers as int, but JSON uses float64
	if mt, ok := config["max_tokens"].(float64); ok {
		a.maxTokens = int(mt)
	}

	return a, nil
}

func (a *anthropicAgent) Execute(ctx context.Context, input model.TaskInput) (*model.AgentOutput, error) {
	messages := buildAnthropicMessages(input)

	reqBody := map[string]any{
		"model":       a.modelName,
		"messages":    messages,
		"max_tokens":  a.maxTokens,
		"temperature": a.temperature,
	}
	if input.System != "" {
		reqBody["system"] = input.System
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/v1/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage map[string]any `json:"usage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var text string
	for _, block := range result.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	return &model.AgentOutput{
		Text: text,
		Metadata: map[string]any{
			"usage": result.Usage,
		},
	}, nil
}

func (a *anthropicAgent) Close() error {
	a.client.CloseIdleConnections()
	return nil
}

func buildAnthropicMessages(input model.TaskInput) []map[string]string {
	var messages []map[string]string

	if len(input.Messages) > 0 {
		for _, m := range input.Messages {
			messages = append(messages, map[string]string{
				"role":    m.Role,
				"content": m.Content,
			})
		}
	} else if input.Prompt != "" {
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": input.Prompt,
		})
	}

	return messages
}

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
	Register("openai", func(config map[string]any) (Agent, error) {
		return newOpenAIAgent(config)
	})
}

type openAIAgent struct {
	apiKey      string
	baseURL     string
	modelName   string
	temperature float64
	client      *http.Client
}

func newOpenAIAgent(config map[string]any) (*openAIAgent, error) {
	a := &openAIAgent{
		baseURL:     "https://api.openai.com/v1",
		modelName:   "gpt-4",
		temperature: 0.0,
		client:      &http.Client{Timeout: 120 * time.Second},
	}

	if key, ok := config["api_key"].(string); ok {
		a.apiKey = key
	} else {
		return nil, fmt.Errorf("openai agent: api_key is required")
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

	return a, nil
}

func (a *openAIAgent) Execute(ctx context.Context, input model.TaskInput) (*model.AgentOutput, error) {
	messages := buildOpenAIMessages(input)

	reqBody := map[string]any{
		"model":       a.modelName,
		"messages":    messages,
		"temperature": a.temperature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

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
		return nil, fmt.Errorf("OpenAI API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage map[string]any `json:"usage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &model.AgentOutput{
		Text: result.Choices[0].Message.Content,
		Metadata: map[string]any{
			"usage": result.Usage,
		},
	}, nil
}

func (a *openAIAgent) Close() error {
	a.client.CloseIdleConnections()
	return nil
}

func buildOpenAIMessages(input model.TaskInput) []map[string]string {
	var messages []map[string]string

	if input.System != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": input.System,
		})
	}

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

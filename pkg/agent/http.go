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
	Register("http", func(config map[string]any) (Agent, error) {
		return newHTTPAgent(config)
	})
}

// httpAgent sends requests to an arbitrary HTTP API.
type httpAgent struct {
	url          string
	method       string
	headers      map[string]string
	bodyTmpl     string // Go template for request body
	responsePath string // JSONPath-like dot path to extract the response text
	client       *http.Client
}

func newHTTPAgent(config map[string]any) (*httpAgent, error) {
	a := &httpAgent{
		method:       "POST",
		responsePath: "text",
		client:       &http.Client{Timeout: 60 * time.Second},
		headers:      make(map[string]string),
	}

	if url, ok := config["url"].(string); ok {
		a.url = url
	} else {
		return nil, fmt.Errorf("http agent: url is required")
	}

	if method, ok := config["method"].(string); ok {
		a.method = method
	}

	if headers, ok := config["headers"].(map[string]any); ok {
		for k, v := range headers {
			if s, ok := v.(string); ok {
				a.headers[k] = s
			}
		}
	}

	if bodyTmpl, ok := config["body_template"].(string); ok {
		a.bodyTmpl = bodyTmpl
	}

	if rp, ok := config["response_path"].(string); ok {
		a.responsePath = rp
	}

	return a, nil
}

func (a *httpAgent) Execute(ctx context.Context, input model.TaskInput) (*model.AgentOutput, error) {
	// Build request body.
	body := map[string]any{
		"prompt": input.Prompt,
	}
	if input.System != "" {
		body["system"] = input.System
	}
	if len(input.Messages) > 0 {
		body["messages"] = input.Messages
	}
	if len(input.Params) > 0 {
		body["params"] = input.Params
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, a.method, a.url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Try to parse as JSON and extract the response text.
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err == nil {
		if text, ok := result[a.responsePath]; ok {
			return &model.AgentOutput{
				Text:     fmt.Sprintf("%v", text),
				Metadata: result,
			}, nil
		}
	}

	// Fall back to raw text.
	return &model.AgentOutput{
		Text: string(respBody),
	}, nil
}

func (a *httpAgent) Close() error {
	a.client.CloseIdleConnections()
	return nil
}

// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"fmt"
)

// Client is a generic LLM client interface for use by model-based graders.
type Client interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Close() error
}

// CompletionRequest represents a request to an LLM.
type CompletionRequest struct {
	Messages    []Message
	Temperature float64
	MaxTokens   int
}

// CompletionResponse holds the LLM response.
type CompletionResponse struct {
	Content string
	Usage   Usage
}

// Message represents a single conversation message.
type Message struct {
	Role    string
	Content string
}

// Usage tracks token consumption.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// Factory creates an LLM Client from config.
type Factory func(config map[string]any) (Client, error)

var registry = make(map[string]Factory)

// Register adds an LLM client factory.
func Register(name string, factory Factory) {
	registry[name] = factory
}

// Create instantiates an LLM client.
func Create(name string, config map[string]any) (Client, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown llm provider: %q", name)
	}
	return factory(config)
}

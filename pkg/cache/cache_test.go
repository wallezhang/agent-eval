// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wallezhang/agent-eval/pkg/model"
)

type mockAgent struct {
	callCount int
}

func (m *mockAgent) Execute(_ context.Context, _ model.TaskInput) (*model.AgentOutput, error) {
	m.callCount++
	return &model.AgentOutput{
		Text: "mock response",
		Metadata: map[string]any{
			"usage": map[string]any{
				"input_tokens":  100,
				"output_tokens": 50,
			},
		},
	}, nil
}

func (m *mockAgent) Close() error { return nil }

func TestCachedAgent_CacheHit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")
	mock := &mockAgent{}
	cached := Wrap(mock, "test", map[string]any{"model": "test"}, dir, time.Hour)

	input := model.TaskInput{Prompt: "hello"}

	// First call: cache miss.
	out1, err := cached.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("first Execute: %v", err)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount)
	}
	if out1.Text != "mock response" {
		t.Errorf("got text=%q, want %q", out1.Text, "mock response")
	}

	// Second call: cache hit.
	out2, err := cached.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("second Execute: %v", err)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 call (cache hit), got %d", mock.callCount)
	}
	if out2.Text != "mock response" {
		t.Errorf("got text=%q, want %q", out2.Text, "mock response")
	}
	if hit, ok := out2.Metadata["cache_hit"].(bool); !ok || !hit {
		t.Error("expected cache_hit metadata to be true")
	}
}

func TestCachedAgent_DifferentInputs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")
	mock := &mockAgent{}
	cached := Wrap(mock, "test", map[string]any{}, dir, time.Hour)

	input1 := model.TaskInput{Prompt: "hello"}
	input2 := model.TaskInput{Prompt: "world"}

	// Two different inputs should both call the agent.
	_, _ = cached.Execute(context.Background(), input1)
	_, _ = cached.Execute(context.Background(), input2)

	if mock.callCount != 2 {
		t.Errorf("expected 2 calls, got %d", mock.callCount)
	}

	// Same inputs should use cache.
	_, _ = cached.Execute(context.Background(), input1)
	_, _ = cached.Execute(context.Background(), input2)

	if mock.callCount != 2 {
		t.Errorf("expected still 2 calls (cache hits), got %d", mock.callCount)
	}
}

func TestCachedAgent_TTLExpiry(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")
	mock := &mockAgent{}
	// Use very short TTL.
	cached := Wrap(mock, "test", map[string]any{}, dir, 1*time.Millisecond)

	input := model.TaskInput{Prompt: "hello"}

	_, _ = cached.Execute(context.Background(), input)
	if mock.callCount != 1 {
		t.Fatalf("expected 1 call, got %d", mock.callCount)
	}

	// Wait for TTL to expire.
	time.Sleep(10 * time.Millisecond)

	_, _ = cached.Execute(context.Background(), input)
	if mock.callCount != 2 {
		t.Errorf("expected 2 calls (TTL expired), got %d", mock.callCount)
	}
}

func TestCachedAgent_CacheKeyDeterministic(t *testing.T) {
	cached := &CachedAgent{
		agentType: "openai",
		agentCfg:  map[string]any{"model": "gpt-4"},
	}

	input := model.TaskInput{Prompt: "test prompt"}
	key1 := cached.cacheKey(input)
	key2 := cached.cacheKey(input)

	if key1 != key2 {
		t.Errorf("cache keys not deterministic: %q != %q", key1, key2)
	}

	// Different input should produce different key.
	input2 := model.TaskInput{Prompt: "different prompt"}
	key3 := cached.cacheKey(input2)
	if key1 == key3 {
		t.Error("different inputs produced same cache key")
	}
}

func TestCachedAgent_CacheDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")
	mock := &mockAgent{}
	cached := Wrap(mock, "test", map[string]any{}, dir, time.Hour)

	_, _ = cached.Execute(context.Background(), model.TaskInput{Prompt: "hello"})

	// Verify cache directory was created.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("cache directory was not created")
	}
}

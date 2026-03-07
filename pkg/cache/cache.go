// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wallezhang/agent-eval/pkg/agent"
	"github.com/wallezhang/agent-eval/pkg/model"
)

// CachedAgent wraps an Agent and caches its responses on disk.
type CachedAgent struct {
	inner     agent.Agent
	cacheDir  string
	ttl       time.Duration
	agentType string
	agentCfg  map[string]any
}

// cacheEntry represents a cached agent response.
type cacheEntry struct {
	Output    model.AgentOutput `json:"output"`
	CachedAt  time.Time         `json:"cached_at"`
	CacheKey  string            `json:"cache_key"`
}

// Wrap wraps an agent with disk-based caching.
func Wrap(inner agent.Agent, agentType string, agentCfg map[string]any, cacheDir string, ttl time.Duration) *CachedAgent {
	return &CachedAgent{
		inner:     inner,
		cacheDir:  cacheDir,
		ttl:       ttl,
		agentType: agentType,
		agentCfg:  agentCfg,
	}
}

// Execute returns a cached response if available, otherwise delegates to the inner agent.
func (c *CachedAgent) Execute(ctx context.Context, input model.TaskInput) (*model.AgentOutput, error) {
	key := c.cacheKey(input)
	path := c.cachePath(key)

	// Try to read from cache.
	if entry, err := c.readCache(path); err == nil {
		if c.ttl <= 0 || time.Since(entry.CachedAt) < c.ttl {
			output := entry.Output
			if output.Metadata == nil {
				output.Metadata = make(map[string]any)
			}
			output.Metadata["cache_hit"] = true
			return &output, nil
		}
	}

	// Cache miss: call the inner agent.
	output, err := c.inner.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	// Write to cache.
	if writeErr := c.writeCache(path, key, *output); writeErr != nil {
		// Non-fatal: log but don't fail.
		_ = writeErr
	}

	return output, nil
}

// Close closes the inner agent.
func (c *CachedAgent) Close() error {
	return c.inner.Close()
}

func (c *CachedAgent) cacheKey(input model.TaskInput) string {
	h := sha256.New()
	h.Write([]byte(c.agentType))
	if cfgBytes, err := json.Marshal(c.agentCfg); err == nil {
		h.Write(cfgBytes)
	}
	if inputBytes, err := json.Marshal(input); err == nil {
		h.Write(inputBytes)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *CachedAgent) cachePath(key string) string {
	// Use first 2 chars as subdirectory to avoid too many files in one dir.
	return filepath.Join(c.cacheDir, key[:2], key+".json")
}

func (c *CachedAgent) readCache(path string) (*cacheEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (c *CachedAgent) writeCache(path, key string, output model.AgentOutput) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	entry := cacheEntry{
		Output:   output,
		CachedAt: time.Now(),
		CacheKey: key,
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

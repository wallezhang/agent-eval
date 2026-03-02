// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func TestNewCommandAgent_HasPromptTemplate(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		expected bool
	}{
		{
			name:     "no template variable",
			args:     []any{"--verbose", "--output", "json"},
			expected: false,
		},
		{
			name:     "template in one arg",
			args:     []any{"--prompt", "{{.Prompt}}"},
			expected: true,
		},
		{
			name:     "template embedded in arg",
			args:     []any{"prefix-{{.Prompt}}-suffix"},
			expected: true,
		},
		{
			name:     "no args",
			args:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]any{"command": "echo"}
			if tt.args != nil {
				config["args"] = tt.args
			}
			a, err := newCommandAgent(config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.hasPromptTemplate != tt.expected {
				t.Errorf("hasPromptTemplate = %v, want %v", a.hasPromptTemplate, tt.expected)
			}
		})
	}
}

func TestCommandAgent_Execute_StdinFallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	// Without {{.Prompt}} in args, prompt should be passed via stdin.
	a, err := newCommandAgent(map[string]any{
		"command": "cat",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := a.Execute(context.Background(), model.TaskInput{Prompt: "hello from stdin"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Text != "hello from stdin" {
		t.Errorf("output = %q, want %q", out.Text, "hello from stdin")
	}
}

func TestCommandAgent_Execute_PromptInArgs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	// With {{.Prompt}} in args, prompt should be substituted into args.
	a, err := newCommandAgent(map[string]any{
		"command": "echo",
		"args":    []any{"{{.Prompt}}"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := a.Execute(context.Background(), model.TaskInput{Prompt: "hello from args"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Text != "hello from args" {
		t.Errorf("output = %q, want %q", out.Text, "hello from args")
	}
}

func TestCommandAgent_Execute_PromptEmbeddedInArg(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	a, err := newCommandAgent(map[string]any{
		"command": "echo",
		"args":    []any{"prefix-{{.Prompt}}-suffix"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := a.Execute(context.Background(), model.TaskInput{Prompt: "MIDDLE"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Text != "prefix-MIDDLE-suffix" {
		t.Errorf("output = %q, want %q", out.Text, "prefix-MIDDLE-suffix")
	}
}

func TestCommandAgent_Execute_MultipleArgsWithTemplate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	// printf "%s|%s" arg1 arg2 — verifies both args get substituted.
	a, err := newCommandAgent(map[string]any{
		"command": "printf",
		"args":    []any{"%s|%s", "{{.Prompt}}", "{{.Prompt}}"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := a.Execute(context.Background(), model.TaskInput{Prompt: "X"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Text != "X|X" {
		t.Errorf("output = %q, want %q", out.Text, "X|X")
	}
}

func TestCommandAgent_Execute_TemplateDoesNotMutateOriginalArgs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	a, err := newCommandAgent(map[string]any{
		"command": "echo",
		"args":    []any{"{{.Prompt}}"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Execute twice to verify the original args slice is not mutated.
	_, err = a.Execute(context.Background(), model.TaskInput{Prompt: "first"})
	if err != nil {
		t.Fatalf("first Execute failed: %v", err)
	}

	if a.args[0] != "{{.Prompt}}" {
		t.Fatalf("original args mutated: got %q, want %q", a.args[0], "{{.Prompt}}")
	}

	out, err := a.Execute(context.Background(), model.TaskInput{Prompt: "second"})
	if err != nil {
		t.Fatalf("second Execute failed: %v", err)
	}
	if out.Text != "second" {
		t.Errorf("output = %q, want %q", out.Text, "second")
	}
}

func TestNewCommandAgent_WorkingDir(t *testing.T) {
	a, err := newCommandAgent(map[string]any{
		"command":     "echo",
		"working_dir": "/tmp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.workingDir != "/tmp" {
		t.Errorf("workingDir = %q, want %q", a.workingDir, "/tmp")
	}
}

func TestNewCommandAgent_WorkingDirDefault(t *testing.T) {
	a, err := newCommandAgent(map[string]any{
		"command": "echo",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.workingDir != "" {
		t.Errorf("workingDir = %q, want empty", a.workingDir)
	}
}

func TestCommandAgent_Execute_WorkingDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	// Create a temp dir and run pwd in it to verify working_dir takes effect.
	dir := t.TempDir()

	a, err := newCommandAgent(map[string]any{
		"command":     "pwd",
		"working_dir": dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := a.Execute(context.Background(), model.TaskInput{Prompt: ""})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Resolve symlinks for macOS where /tmp -> /private/tmp.
	resolvedDir, _ := filepath.EvalSymlinks(dir)
	resolvedOut, _ := filepath.EvalSymlinks(out.Text)
	if resolvedOut != resolvedDir {
		t.Errorf("output = %q, want %q", out.Text, resolvedDir)
	}
}

func TestCommandAgent_Execute_WorkingDirRelativeCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	// Create a temp dir with a script, then run it via relative path using working_dir.
	dir := t.TempDir()
	script := filepath.Join(dir, "hello.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho hello"), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	a, err := newCommandAgent(map[string]any{
		"command":     "./hello.sh",
		"working_dir": dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := a.Execute(context.Background(), model.TaskInput{Prompt: ""})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Text != "hello" {
		t.Errorf("output = %q, want %q", out.Text, "hello")
	}
}

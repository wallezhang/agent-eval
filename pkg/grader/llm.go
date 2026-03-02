// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grader

import (
	"context"
	"fmt"
	"strings"

	"github.com/wallezhang/agent-eval/pkg/llm"
	"github.com/wallezhang/agent-eval/pkg/model"
)

func init() {
	Register("llm", func(config map[string]any) (Grader, error) {
		return newLLMGrader(config)
	})
}

// llmGrader uses an LLM to evaluate agent output against a rubric.
type llmGrader struct {
	client llm.Client
	rubric string
	model  string
}

func newLLMGrader(config map[string]any) (*llmGrader, error) {
	g := &llmGrader{}

	if rubric, ok := config["rubric"].(string); ok {
		g.rubric = rubric
	} else {
		return nil, fmt.Errorf("llm grader: rubric is required")
	}

	// Determine LLM provider.
	provider := "openai"
	if p, ok := config["provider"].(string); ok {
		provider = p
	}

	clientConfig := make(map[string]any)
	for k, v := range config {
		clientConfig[k] = v
	}

	client, err := llm.Create(provider, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("llm grader: creating client: %w", err)
	}
	g.client = client

	return g, nil
}

func (g *llmGrader) Type() string { return "llm" }

func (g *llmGrader) Grade(ctx context.Context, input GradeInput) (*model.GradeResult, error) {
	prompt := buildGradingPrompt(input, g.rubric)

	resp, err := g.client.Complete(ctx, llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.0,
	})
	if err != nil {
		return nil, fmt.Errorf("llm grading failed: %w", err)
	}

	score, pass, reason := parseLLMGradeResponse(resp.Content)

	return &model.GradeResult{
		GraderType: g.Type(),
		Score:      score,
		Pass:       pass,
		Reason:     reason,
	}, nil
}

func buildGradingPrompt(input GradeInput, rubric string) string {
	var b strings.Builder
	b.WriteString("You are an evaluation grader. Evaluate the following agent output against the given criteria.\n\n")

	b.WriteString("## Task\n")
	b.WriteString(fmt.Sprintf("Task ID: %s\n", input.Task.ID))
	if input.Task.Input.Prompt != "" {
		b.WriteString(fmt.Sprintf("Input prompt: %s\n", input.Task.Input.Prompt))
	}
	b.WriteString("\n")

	if input.Task.Expected != nil && input.Task.Expected.Text != "" {
		b.WriteString("## Expected Output\n")
		b.WriteString(input.Task.Expected.Text)
		b.WriteString("\n\n")
	}

	b.WriteString("## Agent Output\n")
	b.WriteString(input.AgentOutput.Text)
	b.WriteString("\n\n")

	b.WriteString("## Evaluation Rubric\n")
	b.WriteString(rubric)
	b.WriteString("\n\n")

	b.WriteString("## Instructions\n")
	b.WriteString("Evaluate the agent output. Respond with EXACTLY this format:\n")
	b.WriteString("SCORE: <number between 0.0 and 1.0>\n")
	b.WriteString("PASS: <true or false>\n")
	b.WriteString("REASON: <brief explanation>\n")

	return b.String()
}

func parseLLMGradeResponse(response string) (score float64, pass bool, reason string) {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SCORE:") {
			fmt.Sscanf(strings.TrimPrefix(line, "SCORE:"), "%f", &score)
		} else if strings.HasPrefix(line, "PASS:") {
			passStr := strings.TrimSpace(strings.TrimPrefix(line, "PASS:"))
			pass = strings.EqualFold(passStr, "true")
		} else if strings.HasPrefix(line, "REASON:") {
			reason = strings.TrimSpace(strings.TrimPrefix(line, "REASON:"))
		}
	}
	if reason == "" {
		reason = response
	}
	return
}

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
	Register("pairwise", func(config map[string]any) (Grader, error) {
		return newPairwiseGrader(config)
	})
}

// pairwiseGrader uses an LLM to compare the agent output against a reference output.
type pairwiseGrader struct {
	client    llm.Client
	criteria  string
	reference string
}

func newPairwiseGrader(config map[string]any) (*pairwiseGrader, error) {
	g := &pairwiseGrader{}

	if criteria, ok := config["criteria"].(string); ok {
		g.criteria = criteria
	} else {
		g.criteria = "Which response is better overall in terms of accuracy, completeness, and helpfulness?"
	}

	if ref, ok := config["reference"].(string); ok {
		g.reference = ref
	}

	provider := "openai"
	if p, ok := config["provider"].(string); ok {
		provider = p
	}

	client, err := llm.Create(provider, config)
	if err != nil {
		return nil, fmt.Errorf("pairwise grader: creating client: %w", err)
	}
	g.client = client

	return g, nil
}

func (g *pairwiseGrader) Type() string { return "pairwise" }

func (g *pairwiseGrader) Grade(ctx context.Context, input GradeInput) (*model.GradeResult, error) {
	reference := g.reference
	if reference == "" && input.Task.Expected != nil {
		reference = input.Task.Expected.Text
	}

	if reference == "" {
		return &model.GradeResult{
			GraderType: g.Type(),
			Score:      0,
			Pass:       false,
			Reason:     "no reference output for pairwise comparison",
		}, nil
	}

	prompt := buildPairwisePrompt(input, reference, g.criteria)

	resp, err := g.client.Complete(ctx, llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.0,
	})
	if err != nil {
		return nil, fmt.Errorf("pairwise grading failed: %w", err)
	}

	score, pass, reason := parsePairwiseResponse(resp.Content)

	return &model.GradeResult{
		GraderType: g.Type(),
		Score:      score,
		Pass:       pass,
		Reason:     reason,
	}, nil
}

func buildPairwisePrompt(input GradeInput, reference, criteria string) string {
	var b strings.Builder
	b.WriteString("You are comparing two responses. Determine which is better.\n\n")

	b.WriteString("## Task\n")
	if input.Task.Input.Prompt != "" {
		b.WriteString(input.Task.Input.Prompt)
		b.WriteString("\n\n")
	}

	b.WriteString("## Response A (Reference)\n")
	b.WriteString(reference)
	b.WriteString("\n\n")

	b.WriteString("## Response B (Agent)\n")
	b.WriteString(input.AgentOutput.Text)
	b.WriteString("\n\n")

	b.WriteString("## Criteria\n")
	b.WriteString(criteria)
	b.WriteString("\n\n")

	b.WriteString("## Instructions\n")
	b.WriteString("Compare Response B against Response A. Respond with EXACTLY:\n")
	b.WriteString("VERDICT: <A_BETTER | B_BETTER | TIE>\n")
	b.WriteString("REASON: <brief explanation>\n")

	return b.String()
}

func parsePairwiseResponse(response string) (score float64, pass bool, reason string) {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "VERDICT:") {
			verdict := strings.TrimSpace(strings.TrimPrefix(line, "VERDICT:"))
			switch strings.ToUpper(verdict) {
			case "B_BETTER":
				score = 1.0
				pass = true
			case "TIE":
				score = 0.5
				pass = true
			case "A_BETTER":
				score = 0.0
				pass = false
			}
		} else if strings.HasPrefix(line, "REASON:") {
			reason = strings.TrimSpace(strings.TrimPrefix(line, "REASON:"))
		}
	}
	if reason == "" {
		reason = fmt.Sprintf("raw response: %s", response)
	}
	return
}

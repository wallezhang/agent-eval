// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// Reporter generates evaluation reports.
type Reporter interface {
	Generate(run *model.EvalRun, outputDir string) error
}

// GenerateAll runs all reporters for the given format setting.
func GenerateAll(run *model.EvalRun, format, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	reporters := selectReporters(format)
	for _, r := range reporters {
		if err := r.Generate(run, outputDir); err != nil {
			return fmt.Errorf("generating report: %w", err)
		}
	}

	return nil
}

func selectReporters(format string) []Reporter {
	switch format {
	case "json":
		return []Reporter{&JSONReporter{}}
	case "html":
		return []Reporter{&HTMLReporter{}}
	case "table":
		return []Reporter{&TableReporter{}}
	case "all":
		return []Reporter{&TableReporter{}, &JSONReporter{}, &HTMLReporter{}}
	default:
		return []Reporter{&TableReporter{}}
	}
}

// runFileName generates a file name for a run output.
func runFileName(run *model.EvalRun, ext string) string {
	return filepath.Join(run.SuiteName + "-" + run.ID[:8] + ext)
}

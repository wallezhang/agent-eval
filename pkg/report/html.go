// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/wallezhang/agent-eval/pkg/model"
)

//go:embed templates/report.html.tmpl
var templateFS embed.FS

// HTMLReporter generates an HTML report using an embedded template.
type HTMLReporter struct{}

func (r *HTMLReporter) Generate(run *model.EvalRun, outputDir string) error {
	tmplData, err := templateFS.ReadFile("templates/report.html.tmpl")
	if err != nil {
		return fmt.Errorf("reading template: %w", err)
	}

	funcMap := template.FuncMap{
		"pct": func(f float64) string {
			return fmt.Sprintf("%.1f%%", f*100)
		},
		"score": func(f float64) string {
			return fmt.Sprintf("%.3f", f)
		},
		"shortID": func(s string) string {
			if len(s) > 8 {
				return s[:8]
			}
			return s
		},
		"statusClass": func(s model.TrialStatus) string {
			switch s {
			case model.TrialStatusPassed:
				return "passed"
			case model.TrialStatusFailed:
				return "failed"
			case model.TrialStatusError:
				return "error"
			default:
				return "pending"
			}
		},
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	filename := runFileName(run, ".html")
	path := filepath.Join(outputDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating HTML report: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, run); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	fmt.Printf("HTML report written to: %s\n", path)
	return nil
}

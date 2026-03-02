// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// JSONReporter outputs evaluation results as a JSON file.
type JSONReporter struct{}

func (r *JSONReporter) Generate(run *model.EvalRun, outputDir string) error {
	filename := runFileName(run, ".json")
	path := filepath.Join(outputDir, filename)

	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing JSON report: %w", err)
	}

	fmt.Printf("JSON report written to: %s\n", path)
	return nil
}

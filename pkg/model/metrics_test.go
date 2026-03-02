// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"math"
	"testing"
)

func TestComputePassAtK(t *testing.T) {
	tests := []struct {
		name     string
		n, c, k  int
		expected float64
	}{
		{"all pass", 10, 10, 1, 1.0},
		{"none pass", 10, 0, 1, 0.0},
		{"k=1 half pass", 10, 5, 1, 0.5},
		{"k=2 half pass", 10, 5, 2, 0.7777777777777778},
		{"k=3 all pass", 3, 3, 3, 1.0},
		{"k greater than n", 5, 3, 6, 0.0},
		{"n zero", 0, 0, 1, 0.0},
		{"k zero", 10, 5, 0, 0.0},
		{"single trial pass", 1, 1, 1, 1.0},
		{"single trial fail", 1, 0, 1, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputePassAtK(tt.n, tt.c, tt.k)
			if math.Abs(got-tt.expected) > 1e-9 {
				t.Errorf("ComputePassAtK(%d, %d, %d) = %f, want %f", tt.n, tt.c, tt.k, got, tt.expected)
			}
		})
	}
}

func TestComputePassPowerK(t *testing.T) {
	tests := []struct {
		name     string
		n, c, k  int
		expected float64
	}{
		{"all pass", 10, 10, 1, 1.0},
		{"none pass", 10, 0, 1, 0.0},
		{"k=1 half pass", 10, 5, 1, 0.5},
		{"k=2 half pass", 10, 5, 2, 0.2222222222222222},
		{"k greater than c", 10, 3, 5, 0.0},
		{"k greater than n", 5, 3, 6, 0.0},
		{"k equals c and n", 5, 5, 5, 1.0},
		{"n zero", 0, 0, 1, 0.0},
		{"k zero", 10, 5, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputePassPowerK(tt.n, tt.c, tt.k)
			if math.Abs(got-tt.expected) > 1e-9 {
				t.Errorf("ComputePassPowerK(%d, %d, %d) = %f, want %f", tt.n, tt.c, tt.k, got, tt.expected)
			}
		})
	}
}

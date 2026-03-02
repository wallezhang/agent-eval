// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import "math"

// ComputePassAtK computes the pass@k metric: the probability that at least one
// of k samples passes, given n total trials with c correct ones.
//
// Formula: 1 - C(n-c, k) / C(n, k)
// where C(a, b) is the binomial coefficient "a choose b".
//
// This is the unbiased estimator from the Codex paper (Chen et al., 2021).
func ComputePassAtK(n, c, k int) float64 {
	if n <= 0 || k <= 0 || k > n {
		return 0
	}
	if c >= n {
		return 1
	}
	if c <= 0 {
		return 0
	}

	// Use log-space to avoid overflow for large n:
	// C(n-c, k) / C(n, k) = product_{i=0}^{k-1} (n-c-i) / (n-i)
	logProb := 0.0
	for i := 0; i < k; i++ {
		logProb += math.Log(float64(n-c-i)) - math.Log(float64(n-i))
	}
	return 1.0 - math.Exp(logProb)
}

// ComputePassPowerK computes the pass^k metric: the probability that all k
// samples pass, given n total trials with c correct ones.
//
// Formula: C(c, k) / C(n, k)
func ComputePassPowerK(n, c, k int) float64 {
	if n <= 0 || k <= 0 || k > n {
		return 0
	}
	if c <= 0 {
		return 0
	}
	if k > c {
		return 0
	}
	if c >= n {
		return 1
	}

	// C(c, k) / C(n, k) = product_{i=0}^{k-1} (c-i) / (n-i)
	logProb := 0.0
	for i := 0; i < k; i++ {
		logProb += math.Log(float64(c-i)) - math.Log(float64(n-i))
	}
	return math.Exp(logProb)
}

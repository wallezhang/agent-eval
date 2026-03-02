// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"sync"

	"github.com/wallezhang/agent-eval/pkg/model"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

// trialFunc executes a single work item and returns a Trial result.
type trialFunc func(ctx context.Context, item workItem) (*model.Trial, error)

// scheduler manages concurrent execution of trials with rate limiting.
type scheduler struct {
	concurrency int
	rps         int
}

func newScheduler(concurrency, rps int) *scheduler {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &scheduler{
		concurrency: concurrency,
		rps:         rps,
	}
}

// Run executes all work items concurrently, respecting concurrency and rate limits.
func (s *scheduler) Run(ctx context.Context, items []workItem, fn trialFunc) ([]*model.Trial, error) {
	results := make([]*model.Trial, len(items))
	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(s.concurrency)

	var limiter *rate.Limiter
	if s.rps > 0 {
		limiter = rate.NewLimiter(rate.Limit(s.rps), s.rps)
	}

	for i, item := range items {
		i, item := i, item
		g.Go(func() error {
			if limiter != nil {
				if err := limiter.Wait(ctx); err != nil {
					return err
				}
			}

			trial, err := fn(ctx, item)
			if err != nil {
				// Record the error as part of the trial rather than failing the whole run.
				trial = &model.Trial{
					TaskID: item.task.ID,
					Index:  item.trialIndex,
					Status: model.TrialStatusError,
					Error:  err.Error(),
				}
			}

			mu.Lock()
			results[i] = trial
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

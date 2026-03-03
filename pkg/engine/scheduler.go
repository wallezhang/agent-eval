// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

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
	logger      *log.Logger
}

func newScheduler(concurrency, rps int, logger *log.Logger) *scheduler {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &scheduler{
		concurrency: concurrency,
		rps:         rps,
		logger:      logger,
	}
}

// Run executes all work items concurrently, respecting concurrency and rate limits.
func (s *scheduler) Run(ctx context.Context, items []workItem, fn trialFunc) ([]*model.Trial, error) {
	total := len(items)
	results := make([]*model.Trial, total)
	var mu sync.Mutex
	var completed int64

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

			done := atomic.AddInt64(&completed, 1)
			s.logger.Printf("[%d/%d] Task %q trial #%d: %s (score=%.2f, %dms)",
				done, total, item.task.ID, item.trialIndex+1,
				trial.Status, trial.Score, trial.DurationMS)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

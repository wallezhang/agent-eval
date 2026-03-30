// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandlerSSE_NotFound(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/runs/nonexistent/sse", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerSSE_StreamsEvents(t *testing.T) {
	srv, _ := newTestServerWithProject(t)

	// Start a run in RunManager
	srv.runManager.Start("sse-test-run", "proj")

	// Create a real HTTP server so SSE streaming works
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// Send events from a goroutine, then finish the run
	go func() {
		// Small delay to let the SSE connection establish
		time.Sleep(50 * time.Millisecond)

		srv.runManager.SendEvent("sse-test-run", SSEEvent{
			Type: "progress",
			Data: map[string]any{"completed": 1, "total": 3},
		})
		srv.runManager.SendEvent("sse-test-run", SSEEvent{
			Type: "task_complete",
			Data: map[string]any{"task_id": "t1", "score": 1.0},
		})

		// Small delay to let events be consumed
		time.Sleep(50 * time.Millisecond)
		srv.runManager.Finish("sse-test-run")
	}()

	// Make SSE request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/projects/proj/runs/sse-test-run/sse", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("making SSE request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("expected Content-Type text/event-stream, got %q", ct)
	}

	// Read SSE event lines
	scanner := bufio.NewScanner(resp.Body)
	var eventLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "data:") {
			eventLines = append(eventLines, line)
		}
	}

	if err := scanner.Err(); err != nil && !strings.Contains(err.Error(), "context") {
		t.Fatalf("scanner error: %v", err)
	}

	// We should have at least 2 event lines and 2 data lines (one pair per event)
	if len(eventLines) < 4 {
		t.Fatalf("expected at least 4 event/data lines, got %d: %v", len(eventLines), eventLines)
	}

	// Verify we got progress and task_complete events
	joined := strings.Join(eventLines, "\n")
	if !strings.Contains(joined, "event: progress") {
		t.Errorf("missing progress event in: %v", eventLines)
	}
	if !strings.Contains(joined, "event: task_complete") {
		t.Errorf("missing task_complete event in: %v", eventLines)
	}
}

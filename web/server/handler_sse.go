// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	run, ok := s.runManager.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "active run not found")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case evt, ok := <-run.EventChan:
			if !ok {
				return // channel closed — run finished
			}
			data, err := json.Marshal(evt.Data)
			if err != nil {
				s.logger.Printf("WARNING: failed to marshal SSE event: %v", err)
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, data)
			flusher.Flush()
		}
	}
}

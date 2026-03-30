// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wallezhang/agent-eval/web/server"
)

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	port := flag.Int("port", 8080, "server listen port")
	home := flag.String("home", defaultHome(), "agent-eval home directory")
	flag.Parse()

	srv, err := server.New(*home, nil)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting agent-eval web server %s on %s (home=%s)", version, addr, *home)

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func defaultHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".agent-eval"
	}
	return filepath.Join(home, ".agent-eval")
}

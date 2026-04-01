// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wallezhang/agent-eval/web/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web UI server",
	Long:  `Start the agent-eval web UI server for managing projects, configs, runs, and results through a browser.`,
	RunE:  runServer,
}

func init() {
	serverCmd.Flags().IntP("port", "p", 8080, "Server listen port")
	serverCmd.Flags().String("home", defaultHome(), "Agent-eval home directory for project registry")
	rootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, _ []string) error {
	port, _ := cmd.Flags().GetInt("port")
	home, _ := cmd.Flags().GetString("home")

	logger := log.New(os.Stdout, "[agent-eval-web] ", log.LstdFlags)

	srv, err := server.New(home, logger)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	addr := fmt.Sprintf(":%d", port)
	logger.Printf("Starting agent-eval web server %s on http://localhost%s (home=%s)", version, addr, home)

	if err := http.ListenAndServe(addr, srv); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func defaultHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".agent-eval"
	}
	return filepath.Join(home, ".agent-eval")
}

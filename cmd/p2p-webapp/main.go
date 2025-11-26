// CRC: crc-CommandRouter.md, Spec: main.md
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zot/p2p-webapp/internal/bundle"
	"github.com/zot/p2p-webapp/internal/commands"
	"github.com/zot/p2p-webapp/internal/config"
	"github.com/zot/p2p-webapp/internal/ipfs"
	"github.com/zot/p2p-webapp/internal/peer"
	"github.com/zot/p2p-webapp/internal/server"
)

var (
	noOpen  bool
	linger  bool
	verbose int
	port    int
	dir     string
)

// CRC: crc-CommandRouter.md
var rootCmd = &cobra.Command{
	Use:   "p2p-webapp",
	Short: "A Go application to host peer-to-peer applications",
	Long: `p2p-webapp is a Go application that hosts peer-to-peer applications.
It proxies opinionated IPFS and libp2p operations for managed peers
and provides a TypeScript library for easy communication with the Go application.

Copyright (C) 2025, Bill Burdick
MIT Licensed
Project URL: https://github.com/zot/p2p-webapp


Default mode (no --dir flag):
  Serves directly from the bundled site without extracting to filesystem.

With --dir flag:
  Serves from the specified directory which must contain:
  - html/: website to serve (must contain index.html)
  - ipfs/: content to make available in IPFS (optional)
  - storage/: server storage (peer keys, etc.)`,
	RunE: runServe,
}

func init() {
	rootCmd.Flags().BoolVar(&noOpen, "noopen", false, "Do not open browser automatically")
	rootCmd.Flags().BoolVar(&linger, "linger", false, "Keep server running after all WebSocket connections close")
	rootCmd.Flags().CountVarP(&verbose, "verbose", "v", "Verbose output (can be specified multiple times: -v, -vv, -vvv)")
	rootCmd.Flags().IntVarP(&port, "port", "p", 0, "Port to listen on (default: auto-select starting from 10000)")
	rootCmd.Flags().StringVar(&dir, "dir", "", "Directory to serve from (if not specified, serves from bundled site)")

	rootCmd.AddCommand(commands.ExtractCmd)
	rootCmd.AddCommand(commands.BundleCmd)
	rootCmd.AddCommand(commands.LsCmd)
	rootCmd.AddCommand(commands.CpCmd)
	rootCmd.AddCommand(commands.CatCmd)
	rootCmd.AddCommand(commands.PsCmd)
	rootCmd.AddCommand(commands.KillCmd)
	rootCmd.AddCommand(commands.KillAllCmd)
	rootCmd.AddCommand(commands.VersionCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var srv *server.Server
	var storagePath string
	var cfg *config.Config
	var err error

	if dir != "" {
		// Directory mode: serve from filesystem
		if err := validateDirectoryStructure(dir); err != nil {
			return err
		}

		// Load configuration from directory
		cfg, err = config.LoadFromDir(dir)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		storagePath = filepath.Join(dir, "storage")

		// Create IPFS node
		ipfsNode, err := ipfs.NewNode(ctx, storagePath, 0) // Random port
		if err != nil {
			return fmt.Errorf("failed to create IPFS node: %w", err)
		}
		defer func() {
			if verbose >= 3 {
				fmt.Println("[DEBUG] Closing IPFS node...")
			}
			if err := ipfsNode.Close(); err != nil {
				fmt.Printf("Warning: failed to close IPFS node: %v\n", err)
			}
			if verbose >= 3 {
				fmt.Println("[DEBUG] IPFS node closed")
			}
		}()

		fmt.Printf("Peer ID: %s\n", ipfsNode.PeerID())

		// Merge command-line flags with configuration
		cfg.Merge(port, noOpen, linger, verbose)

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		// Create peer manager
		peerManager, err := peer.NewManager(ctx, ipfsNode.Host(), ipfsNode.Peer(), cfg.Behavior.Verbosity, cfg.P2P.FileUpdateNotifyTopic, cfg.P2P.IPFSGetTimeout.Duration, cfg.P2P.StreamTimeout.Duration)
		if err != nil {
			return fmt.Errorf("failed to create peer manager: %w", err)
		}

		// Create HTTP server from directory
		htmlDir := filepath.Join(dir, "html")
		srv = server.NewServerFromDir(ctx, peerManager, cfg, htmlDir)
	} else {
		// Bundle mode: serve from bundled site
		bundleReader, err := bundle.GetBundleReader()
		if err != nil {
			return fmt.Errorf("failed to read bundle: %w", err)
		}
		if bundleReader == nil {
			return fmt.Errorf("binary is not bundled\nUse --dir to serve from a directory, or use a bundled binary")
		}

		// Load configuration from bundle
		cfg, err = config.LoadFromZIP(bundleReader)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Merge command-line flags with configuration
		cfg.Merge(port, noOpen, linger, verbose)

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		// Create temporary storage directory in current directory
		storagePath = ".p2p-webapp-storage"
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory: %w", err)
		}

		// Create IPFS node
		ipfsNode, err := ipfs.NewNode(ctx, storagePath, 0) // Random port
		if err != nil {
			return fmt.Errorf("failed to create IPFS node: %w", err)
		}
		defer func() {
			if verbose >= 3 {
				fmt.Println("[DEBUG] Closing IPFS node...")
			}
			if err := ipfsNode.Close(); err != nil {
				fmt.Printf("Warning: failed to close IPFS node: %v\n", err)
			}
			if verbose >= 3 {
				fmt.Println("[DEBUG] IPFS node closed")
			}
		}()

		fmt.Printf("Peer ID: %s\n", ipfsNode.PeerID())

		// Create peer manager
		peerManager, err := peer.NewManager(ctx, ipfsNode.Host(), ipfsNode.Peer(), cfg.Behavior.Verbosity, cfg.P2P.FileUpdateNotifyTopic, cfg.P2P.IPFSGetTimeout.Duration, cfg.P2P.StreamTimeout.Duration)
		if err != nil {
			return fmt.Errorf("failed to create peer manager: %w", err)
		}

		// Create HTTP server from bundle
		srv = server.NewServerFromBundle(ctx, peerManager, cfg, bundleReader)
	}

	// Start server
	if err := srv.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer func() {
		if verbose >= 3 {
			fmt.Println("[DEBUG] Calling srv.Stop() from defer...")
		}
		if err := srv.Stop(); err != nil {
			fmt.Printf("Warning: srv.Stop() returned error: %v\n", err)
		}
	}()

	// Open browser unless configured not to
	if cfg.Behavior.AutoOpenBrowser {
		if err := srv.OpenBrowser(); err != nil {
			fmt.Printf("Failed to open browser: %v\n", err)
			fmt.Printf("Please open http://localhost:%d manually\n", srv.Port())
		}
	} else {
		fmt.Printf("Server running at http://localhost:%d\n", srv.Port())
	}

	// Wait for interrupt signal or server context cancellation (auto-exit)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case <-sigCh:
		fmt.Println("\nShutting down...")
		// Defers will handle cleanup in correct order
	case <-srv.Done():
		// Server context was cancelled (auto-exit triggered)
		fmt.Println("Server context cancelled")
		// Defers will handle cleanup in correct order
	}

	return nil
}

func validateDirectoryStructure(baseDir string) error {
	// Check html directory
	htmlDir := filepath.Join(baseDir, "html")
	if _, err := os.Stat(htmlDir); os.IsNotExist(err) {
		return fmt.Errorf("html directory not found in %s", baseDir)
	}

	// Check for index.html
	indexPath := filepath.Join(htmlDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return fmt.Errorf("html/index.html not found in %s", baseDir)
	}

	// Create ipfs directory if it doesn't exist
	ipfsDir := filepath.Join(baseDir, "ipfs")
	if _, err := os.Stat(ipfsDir); os.IsNotExist(err) {
		if err := os.Mkdir(ipfsDir, 0755); err != nil {
			return fmt.Errorf("failed to create ipfs directory: %w", err)
		}
	}

	// Create storage directory if it doesn't exist
	storageDir := filepath.Join(baseDir, "storage")
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		if err := os.Mkdir(storageDir, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory: %w", err)
		}
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

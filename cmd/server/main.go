package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/cRotermund/skills-mcp-go/pkg/skills"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	logger := log.New(os.Stderr, "skills-server: ", log.LstdFlags)
	root, err := configuredSkillsDirectory()
	if err != nil {
		logger.Fatal(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		logger.Fatalf("create skills directory: %v", err)
	}

	scanner := skills.NewScanner()
	initial, err := scanner.Scan(root)
	if err != nil {
		logger.Fatal(err)
	}
	logDiagnostics(logger, initial.Diagnostics)
	repository, err := skills.NewRepository(initial.Skills)
	if err != nil {
		logger.Fatal(err)
	}

	watcher, err := skills.NewWatcher(root, scanner, logger)
	if err != nil {
		logger.Fatalf("start skill watcher: %v", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go func() {
		if err := watcher.Run(ctx, func(result skills.ScanResult) {
			logDiagnostics(logger, result.Diagnostics)
			if err := repository.Replace(result.Skills); err != nil {
				logger.Printf("replace skill registry: %v", err)
			}
		}); err != nil {
			logger.Printf("skill watcher stopped: %v", err)
		}
	}()

	mcpServer := NewMCPServer(repository)
	if err := server.ServeStdio(mcpServer, server.WithErrorLogger(logger)); err != nil {
		logger.Printf("stdio server stopped: %v", err)
	}
}

func configuredSkillsDirectory() (string, error) {
	root := os.Getenv("SKILLS_DIR")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve user home: %w", err)
		}
		root = filepath.Join(home, ".config", "agentic", "skills")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve skills directory: %w", err)
	}
	return filepath.Clean(root), nil
}

func logDiagnostics(logger *log.Logger, diagnostics []skills.Diagnostic) {
	for _, diagnostic := range diagnostics {
		logger.Print(diagnostic.Error())
	}
}

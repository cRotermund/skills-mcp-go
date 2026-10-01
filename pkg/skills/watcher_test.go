package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherRescansAddedSkill(t *testing.T) {
	root := t.TempDir()
	writeSkillFixture(t, root, "initial", "1.0.0", "Initial")
	scanner := NewScanner()
	initial, err := scanner.Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	updates := make(chan ScanResult, 1)
	watcher, err := NewWatcher(root, scanner, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- watcher.Run(ctx, func(result ScanResult) { updates <- result }) }()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("watcher did not stop")
		}
	}()

	if err := os.MkdirAll(filepath.Join(root, "added"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "added", "SKILL.md"), []byte("---\nname: added\nversion: 1.0.0\ndescription: Added\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(3 * time.Second)
	for {
		select {
		case result := <-updates:
			if len(result.Skills) == 2 {
				if len(initial.Skills) != 1 {
					t.Fatalf("initial=%d", len(initial.Skills))
				}
				return
			}
		case <-deadline:
			t.Fatal("watcher did not report a rescan containing the new skill")
		}
	}
}

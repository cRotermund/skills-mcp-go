package skills

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	root     string
	scanner  Scanner
	watcher  *fsnotify.Watcher
	interval time.Duration
	logger   *log.Logger
}

func NewWatcher(root string, scanner Scanner, logger *log.Logger) (*Watcher, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	watcher := &Watcher{root: root, scanner: scanner, watcher: fsWatcher, interval: 100 * time.Millisecond, logger: logger}
	if err := watcher.addDirectories(); err != nil {
		_ = fsWatcher.Close()
		return nil, err
	}
	return watcher, nil
}

func (w *Watcher) Run(ctx context.Context, onScan func(ScanResult)) error {
	defer w.watcher.Close()
	var timer *time.Timer
	var timerC <-chan time.Time
	trigger := func() {
		if timer == nil {
			timer = time.NewTimer(w.interval)
		} else {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(w.interval)
		}
		timerC = timer.C
	}
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if err := w.addDirectoryTree(event.Name); err != nil {
						w.log("watch new directory %s: %v", event.Name, err)
					}
				}
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				trigger()
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			w.log("filesystem watcher: %v", err)
		case <-timerC:
			result, err := w.scanner.Scan(w.root)
			if err != nil {
				w.log("rescan skills: %v", err)
			} else {
				onScan(result)
			}
			timerC = nil
		}
	}
}

func (w *Watcher) addDirectories() error {
	return w.addDirectoryTree(w.root)
}

func (w *Watcher) addDirectoryTree(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return filepath.SkipDir
		}
		return w.watcher.Add(path)
	})
}

func (w *Watcher) log(format string, args ...any) {
	if w.logger != nil {
		w.logger.Printf(format, args...)
	}
}

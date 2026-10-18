//go:build !android && !ios && !mobile && !nacl
// +build !android,!ios,!mobile,!nacl

// This file is copied from fyne.io/fyne/app/settings_desktop.go to avoid an import loop

package internal

import (
	"os"
	"path/filepath"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/fsnotify/fsnotify"
)

func watchFileAddTarget(watcher *fsnotify.Watcher, path string) {
	dir := filepath.Dir(path)
	ensureDirExists(dir)

	err := watcher.Add(dir)
	if err != nil {
		granio.LogError("Settings watch error:", err)
	}
}

func ensureDirExists(dir string) {
	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		return
	}

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		granio.LogError("Unable to create settings storage:", err)
	}
}

func watchFile(path string, callback func()) *fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		granio.LogError("Failed to watch settings file:", err)
		return nil
	}

	go func() {
		for event := range watcher.Events {
			if event.Op&fsnotify.Remove != 0 { // if it was deleted then watch again
				watcher.Remove(path) // fsnotify returns false positives, see https://github.com/fsnotify/fsnotify/issues/268

				watchFileAddTarget(watcher, path)
			} else {
				callback()
			}
		}

		err = watcher.Close()
		if err != nil {
			granio.LogError("Settings un-watch error:", err)
		}
	}()

	watchFileAddTarget(watcher, path)
	return watcher
}

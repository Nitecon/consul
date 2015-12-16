package runner

import (
	"github.com/howeyc/fsnotify"
	"os"
	"path/filepath"
	"strings"
)

func watchFolder(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fatal(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if isWatchedFile(ev.Name) {
					watcherLog("sending event %s", ev)
					startChannel <- ev.String()
				}
			case err := <-watcher.Error:
				watcherLog("error: %s", err)
			}
		}
	}()

	watcherLog("Watching %s", path)
	err = watcher.Watch(path)

	if err != nil {
		fatal(err)
	}
}

func watch() {
	root := root()
	/* This was added from https://github.com/jsimnz/fresh/commit/9d8e121f043d783c52bcb4a6f2644304f37f18b9 */
	ignorePathsArr := strings.Split(settings["ignore_dirs"], ",")

	watcherLog("Ignore Paths: %v", ignorePathsArr)

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !isTmpDir(path) && !isIgnoredDir(path) {
			for _, igp := range ignorePathsArr{
				absIgp, _ := filepath.Abs(igp)
				absPath, _ := filepath.Abs(path)
				if strings.Contains(absPath, absIgp){
					watcherLog("Ignoring %s", path)
					return filepath.SkipDir
				}
			}
			if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
				watcherLog("Ignoring %s", path)
				return filepath.SkipDir
			}
			watchFolder(path)
		}

		return err
	})
}

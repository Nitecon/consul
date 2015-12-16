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
	ignorePaths := make(map[string]bool)
	for _, ipath := range ignorePathsArr{
		if strings.HasPrefix(ipath, " ") || strings.HasSuffix(ipath, " "){
			ipath = strings.Trim(ipath, " ")
			absolutePath, _ := filepath.Abs(ipath)
			ignorePaths[absolutePath] = true
		}
	}
	watcherLog("Ignore Paths: %v", ignorePaths)

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !isTmpDir(path) {
			if _, ignore := ignorePaths[info.Name()]; (len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".")) || ignore {
				return filepath.SkipDir
			}
			watchFolder(path)
		}

		return err
	})
}

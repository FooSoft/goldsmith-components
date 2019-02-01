package devserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Builder interface {
	Build(sourceDir, targetDir, cacheDir string)
}

func DevServe(builder Builder, port int, sourceDir, targetDir, cacheDir string, watchDirs ...string) {
	dirs := append(watchDirs, sourceDir)
	build(dirs, func() { builder.Build(sourceDir, targetDir, cacheDir) })

	httpAddr := fmt.Sprintf(":%d", port)
	httpHandler := http.FileServer(http.Dir(targetDir))

	log.Fatal(http.ListenAndServe(httpAddr, httpHandler))
}

func build(dirs []string, callback func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	var mutex sync.Mutex
	timestamp := time.Now()
	dirty := true

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				mutex.Lock()
				timestamp = time.Now()
				dirty = true
				mutex.Unlock()

				if event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Stat(event.Name)
					if os.IsNotExist(err) {
						continue
					}
					if err != nil {
						log.Fatal(err)
					}

					if info.IsDir() {
						watch(event.Name, watcher)
					} else {
						watcher.Add(event.Name)
					}
				}
			case err := <-watcher.Errors:
				log.Fatal(err)
			}
		}
	}()

	go func() {
		for range time.Tick(10 * time.Millisecond) {
			if dirty && time.Now().Sub(timestamp) > 100*time.Millisecond {
				mutex.Lock()
				dirty = false
				mutex.Unlock()

				callback()
			}
		}
	}()

	for _, dir := range dirs {
		watch(dir, watcher)
	}
}

func watch(dir string, watcher *fsnotify.Watcher) {
	watcher.Add(dir)

	items, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		fullPath := path.Join(dir, item.Name())
		if item.IsDir() {
			watch(fullPath, watcher)
		} else {
			watcher.Add(fullPath)
		}
	}
}

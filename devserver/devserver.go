/*
 * Copyright (c) 2016 Alex Yatskov <alex@foosoft.net>
 * Author: Alex Yatskov <alex@foosoft.net>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

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
	Build(srcDir, dstDir string)
}

func DevServe(builder Builder, port int, srcDir, dstDir string, watchDirs ...string) {
	dirs := append(watchDirs, srcDir)
	build(dirs, func() { builder.Build(srcDir, dstDir) })

	httpAddr := fmt.Sprintf(":%d", port)
	httpHandler := http.FileServer(http.Dir(dstDir))

	log.Fatal(http.ListenAndServe(httpAddr, httpHandler))
}

func build(dirs []string, callback func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	var mtx sync.Mutex
	timestamp := time.Now()
	dirty := true

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				mtx.Lock()
				timestamp = time.Now()
				dirty = true
				mtx.Unlock()

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
				mtx.Lock()
				dirty = false
				mtx.Unlock()

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

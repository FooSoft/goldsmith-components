/*
 * Copyright (c) 2015 Alex Yatskov <alex@foosoft.net>
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

package index

import (
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/FooSoft/goldsmith"
)

type filter func(path string)

type index struct {
	key     string
	dirs    map[string][]Entry
	dirsMtx sync.Mutex
}

type Entry struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	dir     bool
}

func New(key string) goldsmith.Plugin {
	return &index{
		key:  key,
		dirs: make(map[string][]Entry),
	}
}

func (i *index) Process(ctx goldsmith.Context, f goldsmith.File) error {
	entries, err := i.list(path.Dir(f.Path()))
	if err != nil {
		return err
	}

	f.SetValue(i.key, entries)
	return nil
}

func (i *index) list(dir string) ([]Entry, error) {
	i.dirsMtx.Lock()
	defer i.dirsMtx.Unlock()

	if items, ok := i.dirs[dir]; ok {
		return items, nil
	}

	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, item := range items {
		entry := Entry{item.Name(), item.Size(), item.Mode(), item.ModTime(), item.IsDir()}
		entries = append(entries, entry)
	}

	i.dirs[dir] = entries
	return entries, nil
}

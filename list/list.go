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

package list

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/FooSoft/goldsmith"
)

type list struct {
	key     string
	dirs    map[string]Entries
	dirsMtx sync.Mutex
}

func New(key string) goldsmith.Plugin {
	return &list{key: key, dirs: make(map[string]Entries)}
}

func (*list) Accept(ctx goldsmith.Context, f goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(f.Path())) {
	case ".html", ".htm":
		return true
	default:
		return false
	}
}

func (l *list) Process(ctx goldsmith.Context, f goldsmith.File) error {
	relDir := path.Dir(f.Path())
	absDir := path.Join(ctx.SrcDir(), relDir)

	entries, err := l.scan(absDir)
	if err != nil {
		return err
	}

	f.SetValue(l.key, entries)
	ctx.DispatchFile(f)

	return nil
}

func (l *list) scan(dir string) ([]Entry, error) {
	l.dirsMtx.Lock()
	defer l.dirsMtx.Unlock()

	if items, ok := l.dirs[dir]; ok {
		return items, nil
	}

	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var entries Entries
	for _, item := range items {
		entry := Entry{item.Name(), item.Size(), item.Mode(), item.ModTime(), item.IsDir()}
		entries = append(entries, entry)
	}

	sort.Sort(entries)
	l.dirs[dir] = entries

	return entries, nil
}

type Entry struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	Dir     bool
}

type Entries []Entry

func (e Entries) Len() int {
	return len(e)
}

func (e Entries) Less(i, j int) bool {
	e1, e2 := e[i], e[j]

	if e1.Dir && !e2.Dir {
		return true
	}
	if !e1.Dir && e2.Dir {
		return false
	}

	return strings.Compare(e1.Name, e2.Name) == -1
}

func (e Entries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

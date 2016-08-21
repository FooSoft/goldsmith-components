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
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
)

type index struct {
	filename string
	key      string
	meta     map[string]interface{}

	dirs    map[string]*dirSummary
	handled map[string]bool
	dirsMtx sync.Mutex
}

func New(filename, key string, meta map[string]interface{}) goldsmith.Plugin {
	return &index{
		filename: filename,
		key:      key,
		meta:     meta,
		handled:  make(map[string]bool),
		dirs:     make(map[string]*dirSummary),
	}
}

func (*index) Name() string {
	return "index"
}

func (i *index) Process(ctx goldsmith.Context, f goldsmith.File) error {
	i.dirsMtx.Lock()
	defer i.dirsMtx.Unlock()

	curr := f.Path()
	leaf := true

	for {
		if handled, _ := i.handled[curr]; handled {
			break
		}

		i.handled[curr] = true

		dir := path.Dir(curr)
		base := path.Base(curr)

		summary, ok := i.dirs[dir]
		if !ok {
			summary = new(dirSummary)
			i.dirs[dir] = summary
		}

		if leaf {
			if base == i.filename {
				summary.index = f
			} else {
				ctx.DispatchFile(f)
			}
		}

		entry := DirEntry{Name: base, Path: curr, IsDir: !leaf, File: f}
		summary.entries = append(summary.entries, entry)

		if dir == "." {
			break
		}

		curr = dir
		leaf = false
	}

	return nil
}

func (i *index) Finalize(ctx goldsmith.Context) error {
	for name, summary := range i.dirs {
		sort.Sort(summary.entries)

		f := summary.index
		if f == nil {
			f = goldsmith.NewFileFromData(path.Join(name, i.filename), make([]byte, 0))
			for name, value := range i.meta {
				f.SetValue(name, value)
			}
		}

		f.SetValue(i.key, summary.entries)
		ctx.DispatchFile(f)
	}

	return nil
}

type dirSummary struct {
	entries DirEntries
	index   goldsmith.File
}

type DirEntry struct {
	Name  string
	Path  string
	IsDir bool
	File  goldsmith.File
}

type DirEntries []DirEntry

func (d DirEntries) Len() int {
	return len(d)
}

func (d DirEntries) Less(i, j int) bool {
	d1, d2 := d[i], d[j]

	if d1.IsDir && !d2.IsDir {
		return true
	}
	if !d1.IsDir && d2.IsDir {
		return false
	}

	return strings.Compare(d1.Name, d2.Name) == -1
}

func (d DirEntries) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

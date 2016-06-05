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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/FooSoft/goldsmith"
)

type index struct {
	indexFile string
	key       string

	dirs    map[string]dirSummary
	dirsMtx sync.Mutex
}

func New(indexFile, key string, meta map[string]interface{}) goldsmith.Plugin {
	return &index{
		indexFile: indexFile,
		key:       key,
		dirs:      make(map[string]dirSummary),
	}
}

func (i *index) Process(ctx goldsmith.Context, f goldsmith.File) error {
	defer ctx.DispatchFile(f)

	relDir := path.Dir(f.Path())
	absDir := path.Join(ctx.SrcDir(), relDir)

	i.dirsMtx.Lock()
	defer i.dirsMtx.Unlock()

	if _, ok := i.dirs[relDir]; ok {
		return nil
	}

	fi, err := ioutil.ReadDir(absDir)
	if err != nil {
		return err
	}

	var ds dirSummary
	for _, info := range fi {
		item := DirItem{info.Name(), info.Size(), info.Mode(), info.ModTime(), info.IsDir()}
		ds.items = append(ds.items, item)
		if path.Base(item.Name) == i.indexFile {
			ds.hasIndex = true
		}
	}

	sort.Sort(ds.items)
	i.dirs[relDir] = ds

	return err
}

func (i *index) Finalize(ctx goldsmith.Context) error {
	for dn, ds := range i.dirs {
		if ds.hasIndex {
			continue
		}

		f := goldsmith.NewFileFromData(path.Join(dn, i.indexFile), make([]byte, 0))
		f.SetValue(i.key, ds.items)
		ctx.DispatchFile(f)
	}

	return nil
}

type dirSummary struct {
	hasIndex bool
	items    DirItems
}

type DirItem struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	Dir     bool
}

type DirItems []DirItem

func (d DirItems) Len() int {
	return len(d)
}

func (d DirItems) Less(i, j int) bool {
	d1, d2 := d[i], d[j]

	if d1.Dir && !d2.Dir {
		return true
	}
	if !d1.Dir && d2.Dir {
		return false
	}

	return strings.Compare(d1.Name, d2.Name) == -1
}

func (d DirItems) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

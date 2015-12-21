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

package collection

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
)

type Comparer func(i, j goldsmith.File) (less bool)

type collection struct {
	srcKey, dstKey string

	comp    Comparer
	cols    map[string]fileGroup
	colsMtx sync.Mutex

	files    []goldsmith.File
	filesMtx sync.Mutex
}

func New(srcKey, dstKey string, comp Comparer) goldsmith.Plugin {
	return &collection{
		srcKey: srcKey,
		dstKey: dstKey,
		comp:   comp,
		cols:   make(map[string]fileGroup),
	}
}

func (c *collection) Accept(ctx goldsmith.Context, f goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(f.Path())) {
	case ".html", ".htm":
		return true
	default:
		return false
	}
}

func (c *collection) Process(ctx goldsmith.Context, f goldsmith.File) error {
	defer func() {
		c.filesMtx.Lock()
		c.files = append(c.files, f)
		c.filesMtx.Unlock()
	}()

	meta := f.Meta()
	meta[c.dstKey] = c.cols

	col, ok := meta[c.srcKey]
	if !ok {
		return nil
	}

	colStr, ok := col.(string)
	if !ok {
		return nil
	}

	c.colsMtx.Lock()
	{
		fg, ok := c.cols[colStr]
		fg.files = append(fg.files, f)
		if !ok {
			fg.comp = c.comp
			c.cols[colStr] = fg
		}
	}
	c.colsMtx.Unlock()

	return nil
}

func (c *collection) Finalize(ctx goldsmith.Context) error {
	for _, files := range c.cols {
		sort.Sort(files)
	}

	for _, f := range c.files {
		ctx.DispatchFile(f)
	}

	return nil
}

type fileGroup struct {
	files []goldsmith.File
	comp  Comparer
}

func (f fileGroup) Len() int {
	return len(f.files)
}

func (f fileGroup) Swap(i, j int) {
	f.files[i], f.files[j] = f.files[j], f.files[i]
}

func (f fileGroup) Less(i, j int) bool {
	if f.comp == nil {
		return strings.Compare(f.files[i].Path(), f.files[j].Path()) < 0
	}

	return f.comp(f.files[i], f.files[j])
}

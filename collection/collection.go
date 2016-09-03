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
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
)

type comparer func(i, j goldsmith.File) (less bool)

type collection struct {
	srcKey, dstKey string

	comp    comparer
	cols    map[string][]goldsmith.File
	colsMtx sync.Mutex

	files    []goldsmith.File
	filesMtx sync.Mutex
}

func New() *collection {
	return &collection{
		srcKey: "collection",
		dstKey: "collections",
		comp:   nil,
		cols:   make(map[string][]goldsmith.File),
	}
}

func (c *collection) CollectionSrcKey(srcKey string) *collection {
	c.srcKey = srcKey
	return c
}

func (c *collection) CollectionsDstKey(dstKey string) *collection {
	c.dstKey = dstKey
	return c
}

func (c *collection) Comparer(comp comparer) *collection {
	c.comp = comp
	return c
}

func (*collection) Name() string {
	return "collection"
}

func (*collection) Initialize(ctx goldsmith.Context) ([]string, error) {
	return []string{"**/*.html", "**/*.htm"}, nil
}

func (c *collection) Process(ctx goldsmith.Context, f goldsmith.File) error {
	defer func() {
		f.SetValue(c.dstKey, c.cols)

		c.filesMtx.Lock()
		c.files = append(c.files, f)
		c.filesMtx.Unlock()
	}()

	col, ok := f.Value(c.srcKey)
	if !ok {
		return nil
	}

	colStr, ok := col.(string)
	if !ok {
		return nil
	}

	c.colsMtx.Lock()
	{
		files, _ := c.cols[colStr]
		files = append(files, f)
		c.cols[colStr] = files
	}
	c.colsMtx.Unlock()

	return nil
}

func (c *collection) Finalize(ctx goldsmith.Context) error {
	for _, files := range c.cols {
		fg := &fileGroup{files, c.comp}
		sort.Sort(fg)
	}

	for _, f := range c.files {
		ctx.DispatchFile(f)
	}

	return nil
}

type fileGroup struct {
	Files []goldsmith.File
	comp  comparer
}

func (f fileGroup) Len() int {
	return len(f.Files)
}

func (f fileGroup) Swap(i, j int) {
	f.Files[i], f.Files[j] = f.Files[j], f.Files[i]
}

func (f fileGroup) Less(i, j int) bool {
	if f.comp == nil {
		return strings.Compare(f.Files[i].Path(), f.Files[j].Path()) < 0
	}

	return f.comp(f.Files[i], f.Files[j])
}

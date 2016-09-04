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
	collKey, groupsKey string

	comp   comparer
	groups map[string][]goldsmith.File
	files  []goldsmith.File

	mtx sync.Mutex
}

func New() *collection {
	return &collection{
		collKey:   "Collection",
		groupsKey: "Groups",
		comp:      nil,
		groups:    make(map[string][]goldsmith.File),
	}
}

func (c *collection) CollectionKey(collKey string) *collection {
	c.collKey = collKey
	return c
}

func (c *collection) GroupsKey(groupsKey string) *collection {
	c.groupsKey = groupsKey
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
	c.mtx.Lock()
	defer func() {
		f.SetValue(c.groupsKey, c.groups)
		c.files = append(c.files, f)
		c.mtx.Unlock()
	}()

	coll, ok := f.Value(c.collKey)
	if !ok {
		return nil
	}

	var collStrs []string
	switch t := coll.(type) {
	case string:
		collStrs = append(collStrs, t)
	case []string:
		collStrs = append(collStrs, t...)
	}

	for _, collStr := range collStrs {
		files, _ := c.groups[collStr]
		files = append(files, f)
		c.groups[collStr] = files
	}

	return nil
}

func (c *collection) Finalize(ctx goldsmith.Context) error {
	for _, files := range c.groups {
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

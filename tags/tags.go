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

package tags

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
)

type tags struct {
	baseDir string
	key     string
	meta    map[string]interface{}

	info    map[string]TagInfo
	infoMtx sync.Mutex

	files    []goldsmith.File
	filesMtx sync.Mutex
}

type TagInfo struct {
	Files    files
	SafeName string
	RawName  string
	Path     string
}

type TagState struct {
	Index string
	Set   []string
	Info  map[string]TagInfo
}

func New(meta map[string]interface{}) *tags {
	return &tags{
		baseDir: "tags",
		key:     "Tags",
		meta:    meta,
		info:    make(map[string]TagInfo),
	}
}

func (t *tags) BaseDir(dir string) *tags {
	t.baseDir = dir
	return t
}

func (t *tags) Key(key string) *tags {
	t.key = key
	return t
}

func (*tags) Name() string {
	return "tags"
}

func (*tags) Initialize(ctx goldsmith.Context) ([]string, error) {
	return []string{"**/*.html", "**/*.htm"}, nil
}

func (t *tags) Process(ctx goldsmith.Context, f goldsmith.File) error {
	defer func() {
		t.filesMtx.Lock()
		t.files = append(t.files, f)
		t.filesMtx.Unlock()
	}()

	tagData, ok := f.Value(t.key)
	if !ok {
		f.SetValue(t.key, TagState{Info: t.info})
		return nil
	}

	tags, ok := tagData.([]interface{})
	if !ok {
		f.SetValue(t.key, TagState{Info: t.info})
		return nil
	}

	var tagStrs []string
	for _, tag := range tags {
		tagStr, ok := tag.(string)
		if !ok {
			continue
		}

		tagStrs = append(tagStrs, tagStr)

		t.infoMtx.Lock()
		{
			item, ok := t.info[tagStr]
			item.Files = append(item.Files, f)
			if !ok {
				item.SafeName = safeTag(tagStr)
				item.RawName = tagStr
				item.Path = t.tagPagePath(tagStr)
			}

			t.info[tagStr] = item
		}
		t.infoMtx.Unlock()
	}

	sort.Strings(tagStrs)
	f.SetValue(t.key, TagState{Info: t.info, Set: tagStrs})

	return nil
}

func (t *tags) Finalize(ctx goldsmith.Context) error {
	for _, meta := range t.info {
		sort.Sort(meta.Files)
	}

	for _, f := range t.buildPages(ctx, t.info) {
		ctx.DispatchFile(f)
	}

	for _, f := range t.files {
		ctx.DispatchFile(f)
	}

	return nil
}

func (t *tags) buildPages(ctx goldsmith.Context, info map[string]TagInfo) (files []goldsmith.File) {
	for tag := range info {
		f := goldsmith.NewFileFromData(t.tagPagePath(tag), nil)
		f.SetValue(t.key, TagState{Index: tag, Info: t.info})
		for name, value := range t.meta {
			f.SetValue(name, value)
		}

		files = append(files, f)
	}

	return
}

func (t *tags) tagPagePath(tag string) string {
	return filepath.Join(t.baseDir, safeTag(tag), "index.html")
}

func safeTag(tag string) string {
	return strings.ToLower(strings.Replace(tag, " ", "-", -1))
}

type files []goldsmith.File

func (f files) Len() int {
	return len(f)
}

func (f files) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f files) Less(i, j int) bool {
	return strings.Compare(f[i].Path(), f[j].Path()) < 0
}

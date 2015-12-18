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

type files []goldsmith.File

func (f files) Len() int           { return len(f) }
func (f files) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f files) Less(i, j int) bool { return strings.Compare(f[i].Path(), f[j].Path()) < 0 }

type tagInfo struct {
	Files    files
	SafeName string
	Path     string
}

type tagState struct {
	Index string
	Set   []string
	Info  tagInfoMap
}

type tagInfoMap map[string]tagInfo

type tags struct {
	basePath       string
	srcKey, dstKey string
	meta           map[string]interface{}
	info           tagInfoMap
	mtx            sync.Mutex
}

func New(basePath, srcKey, dstKey string, meta map[string]interface{}) goldsmith.Plugin {
	return &tags{
		basePath: basePath,
		srcKey:   srcKey,
		dstKey:   dstKey,
		meta:     meta,
		info:     make(tagInfoMap),
	}
}

func (*tags) Name() string {
	return "Tags"
}

func (*tags) Accept(f goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(f.Path())) {
	case ".html", ".htm":
		return true
	default:
		return false
	}
}

func (t *tags) Process(ctx goldsmith.Context, f goldsmith.File) error {
	meta := f.Meta()

	tagData, ok := meta[t.srcKey]
	if !ok {
		meta[t.dstKey] = tagState{Info: t.info}
		return nil
	}

	tags, ok := tagData.([]interface{})
	if !ok {
		meta[t.dstKey] = tagState{Info: t.info}
		return nil
	}

	var tagStrs []string
	for _, tag := range tags {
		tagStr, ok := tag.(string)
		if !ok {
			continue
		}

		tagStrs = append(tagStrs, tagStr)

		t.mtx.Lock()
		{
			item, ok := t.info[tagStr]
			item.Files = append(item.Files, f)
			if !ok {
				item.SafeName = safeTag(tagStr)
				item.Path = t.tagPagePath(tagStr)
			}

			t.info[tagStr] = item
		}
		t.mtx.Unlock()
	}

	sort.Strings(tagStrs)
	meta[t.dstKey] = tagState{Info: t.info, Set: tagStrs}

	return nil
}

func (t *tags) Finalize(ctx goldsmith.Context, files []goldsmith.File) error {
	for _, meta := range t.info {
		sort.Sort(meta.Files)
	}

	for _, f := range t.buildPages(ctx, t.info) {
		ctx.AddFile(f)
	}

	return nil
}

func (t *tags) buildPages(ctx goldsmith.Context, info tagInfoMap) (files []goldsmith.File) {
	for tag := range info {
		f := goldsmith.NewFileFromData(t.tagPagePath(tag), nil)

		meta := f.Meta()
		meta[t.dstKey] = tagState{Index: tag, Info: t.info}
		for key, value := range t.meta {
			meta[key] = value
		}

		files = append(files, f)
	}

	return
}

func (t *tags) tagPagePath(tag string) string {
	return filepath.Join(t.basePath, safeTag(tag), "index.html")
}

func safeTag(tag string) string {
	return strings.ToLower(strings.Replace(tag, " ", "-", -1))
}

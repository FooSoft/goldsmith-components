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

	"github.com/FooSoft/goldsmith"
)

type tagInfo struct {
	Files    []*goldsmith.File
	SafeName string
	Path     string
}

type tagInfoMap map[string]*tagInfo

type tags struct {
	srcKey, dstKey string
	meta           map[string]interface{}
}

func New(srcKey, dstKey string, meta map[string]interface{}) (goldsmith.Chainer, error) {
	return &tags{srcKey, dstKey, meta}, nil
}

func (*tags) Filter(path string) bool {
	if ext := filepath.Ext(path); ext != ".html" {
		return true
	}

	return false
}

func buildMeta(tag string, tags []string, info tagInfoMap) map[string]interface{} {
	meta := make(map[string]interface{})

	if len(tag) > 0 {
		meta["index"] = tag
	}

	if tags != nil {
		var tagsAlpha []string
		copy(tagsAlpha, tags)
		sort.Strings(tagsAlpha)
		meta["set"] = tags

	}

	if info != nil {
		meta["meta"] = info
	}

	return meta
}

func (t *tags) buildTags(input, output chan *goldsmith.File) tagInfoMap {
	info := make(tagInfoMap)

	for file := range input {
		data, ok := file.Meta[t.srcKey]
		if !ok {
			output <- file
			continue
		}

		tags, ok := data.([]interface{})
		if !ok {
			output <- file
			continue
		}

		var tagStrs []string
		for _, tag := range tags {
			tagStr, ok := tag.(string)
			if !ok {
				continue
			}

			item, ok := info[tagStr]
			if !ok {
				item = &tagInfo{
					SafeName: safeTag(tagStr),
					Path:     t.tagPagePath(tagStr),
				}
			}

			item.Files = append(item.Files, file)
			tagStrs = append(tagStrs, tagStr)
		}

		file.Meta[t.dstKey] = buildMeta("", tagStrs, info)
		output <- file
	}

	return info
}

func (t *tags) buildPages(ctx goldsmith.Context, info tagInfoMap, output chan *goldsmith.File) {
	for tag := range info {
		file := ctx.NewFile(t.tagPagePath(tag))
		for key, value := range t.meta {
			file.Meta[key] = value
		}

		file.Meta[t.dstKey] = buildMeta(tag, nil, info)
		output <- file
	}
}

func (t *tags) tagPagePath(tag string) string {
	return filepath.Join(t.srcKey, safeTag(tag), "index.html")
}

func safeTag(tag string) string {
	return strings.Replace(tag, " ", "-", -1)
}

func (t *tags) Chain(ctx goldsmith.Context, input, output chan *goldsmith.File) {
	defer close(output)
	info := t.buildTags(input, output)
	t.buildPages(ctx, info, output)
}

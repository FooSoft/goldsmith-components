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
	"strings"

	"github.com/FooSoft/goldsmith"
)

type tagInfo struct {
	Files    []*goldsmith.File
	safeName string
	path     string
}

type tagInfoMap map[string]tagInfo

type tags struct {
	srcKey, dstKey string
	meta           map[string]interface{}
}

func New(srcKey, dstKey string, meta map[string]interface{}) (goldsmith.Chainer, error) {
	return &tags{srcKey, dstKey, meta}, nil
}

func (*tags) Filter(path string) bool {
	if ext := filepath.Ext(path); ext == ".html" {
		return true
	}

	return false
}

func (t *tags) buildTagData(input, output chan *goldsmith.File) tagInfoMap {
	tf := make(tagInfoMap)

	for file := range input {
		if data, ok := file.Meta[t.srcKey]; ok {
			tags, ok := data.([]interface{})
			if !ok {
				continue
			}

		tagLoop:
			for _, tag := range tags {
				tagStr, ok := tag.(string)
				if !ok {
					continue
				}

				tagInfo, ok := tf[tagStr]
				if !ok {
					tagInfo.safeName = safeTag(tagStr)
					tagInfo.path = t.buildTagPagePath(tagStr)
				}

				for _, taggedFile := range tagInfo.Files {
					if taggedFile == file {
						continue tagLoop
					}
				}

				tagInfo.Files = append(tagInfo.Files, file)
				tf[tagStr] = tagInfo
			}

			file.Meta[t.dstKey] = map[string]interface{}{
				"all": tf,
			}
		}

		output <- file
	}

	return tf
}

func (t *tags) buildIndexPage(ctx goldsmith.Context, tf tagInfoMap, output chan *goldsmith.File) {
	file, err := ctx.NewFile(t.buildTagIndexPagePath())
	if err != nil {
		panic(err)
	}

	if t.meta != nil {
		file.Meta = t.meta
	}

	file.Meta[t.dstKey] = map[string]interface{}{
		"all": tf,
	}

	output <- file
}

func (t *tags) buildTagPages(ctx goldsmith.Context, tf tagInfoMap, output chan *goldsmith.File) {
	for tag := range tf {
		file, err := ctx.NewFile(t.buildTagPagePath(tag))
		if err != nil {
			panic(err)
		}

		if t.meta != nil {
			file.Meta = t.meta
		}

		file.Meta[t.dstKey] = map[string]interface{}{
			"all": tf,
			"tag": tag,
		}

		output <- file
	}
}

func (t *tags) buildTagPagePath(tag string) string {
	return filepath.Join(t.srcKey, safeTag(tag), "index.html")
}

func (t *tags) buildTagIndexPagePath() string {
	return filepath.Join(t.srcKey, "index.html")
}

func safeTag(tag string) string {
	return strings.Replace(tag, " ", "-", -1)
}

func (t *tags) Chain(ctx goldsmith.Context, input, output chan *goldsmith.File) {
	defer close(output)
	tf := t.buildTagData(input, output)
	t.buildIndexPage(ctx, tf, output)
	t.buildTagPages(ctx, tf, output)
}

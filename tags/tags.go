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

	"github.com/FooSoft/goldsmith"
)

type tags struct {
	srcKey, dstKey string
	outputDir      string
	meta           map[string]interface{}
}

type meta struct {
	All     map[string][]string
	Current string
}

func New(srcKey, dstKey, outputDir string, meta map[string]interface{}) goldsmith.Config {
	return goldsmith.Config{
		Chainer: &tags{srcKey, dstKey, outputDir, meta},
		Globs:   []string{"*.html", "*.html"},
	}
}

func (t *tags) buildMeta(input, output chan *goldsmith.File) meta {
	m := meta{All: make(map[string][]string)}

	for file := range input {
		values, _ := file.Meta[t.srcKey]
		for _, value := range values.([]interface{}) {
			paths, _ := m.All[value.(string)]

			for _, path := range paths {
				if path == file.Path {
					continue
				}
			}

			paths = append(paths, file.Path)
			m.All[value.(string)] = paths
		}

		output <- file
	}

	return m
}

func (t *tags) buildIndex(ctx goldsmith.Context, m meta, output chan *goldsmith.File) {
	path := filepath.Join(t.outputDir, "index.html")

	file, err := ctx.NewFile(path)
	if err != nil {
		file.Err = err
	}

	if t.meta != nil {
		file.Meta = t.meta
	}

	file.Meta[t.dstKey] = meta{All: m.All}
	output <- file
}

func (t *tags) buildPages(ctx goldsmith.Context, m meta, output chan *goldsmith.File) {
	for tag := range m.All {
		path := filepath.Join(t.outputDir, tag, "index.html")

		file, err := ctx.NewFile(path)
		if err != nil {
			file.Err = err
		}

		if t.meta != nil {
			file.Meta = t.meta
		}

		file.Meta[t.dstKey] = m
		output <- file
	}
}

func (t *tags) ChainMultiple(ctx goldsmith.Context, input, output chan *goldsmith.File) {
	defer close(output)

	meta := t.buildMeta(input, output)
	t.buildIndex(ctx, meta, output)
	t.buildPages(ctx, meta, output)
}

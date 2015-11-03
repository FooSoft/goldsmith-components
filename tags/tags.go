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

import "github.com/FooSoft/goldsmith"

type tags struct {
	key, dir string
	meta     map[string]interface{}
	paths    map[string][]string
}

func New(key, dir string, meta map[string]interface{}) goldsmith.Config {
	return goldsmith.Config{
		Chainer: &tags{key, dir, meta, make(map[string][]string)},
		Globs:   []string{"*.html", "*.html"},
	}
}

func (t *tags) ChainMultiple(ctx goldsmith.Context, input chan *goldsmith.File, output chan *goldsmith.File) {
	defer close(output)

	for file := range input {
		values, _ := file.Meta[t.key]
		for _, value := range values.([]string) {
			paths, _ := t.paths[value]

			for _, path := range paths {
				if path == file.Path {
					continue
				}
			}

			paths = append(paths, file.Path)
			t.paths[value] = paths
		}

		output <- file
	}
}

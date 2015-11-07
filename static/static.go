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

package static

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/FooSoft/goldsmith"
)

type static struct {
	src, dst string
	paths    []string
}

func New(src, dst string) (goldsmith.Chainer, error) {
	if filepath.IsAbs(dst) {
		return nil, fmt.Errorf("absolute paths are not supported: %s", dst)
	}

	var paths []string
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			paths = append(paths, path)
		}
		return err
	})

	if err != nil {
		return nil, err
	}

	return &static{src, dst, paths}, nil
}

func (s *static) Chain(ctx goldsmith.Context, input, output chan *goldsmith.File) {
	defer close(output)

	for file := range input {
		output <- file
	}

	for _, path := range s.paths {
		relPath, err := filepath.Rel(s.src, path)
		if err != nil {
			panic(err)
		}

		file, err := ctx.NewFileStatic(filepath.Join(s.dst, relPath))
		if err != nil {
			panic(err)
		}

		var f *os.File
		if f, file.Err = os.Open(path); file.Err == nil {
			_, file.Err = file.Buff.ReadFrom(f)
			f.Close()
		}

		output <- file
	}
}

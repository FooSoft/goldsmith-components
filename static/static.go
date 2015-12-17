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
	"os"
	"path/filepath"

	"github.com/FooSoft/goldsmith"
)

type static struct {
	src, dst string
}

func New(src, dst string) goldsmith.Plugin {
	return &static{src, dst}
}

func (*static) Accept(file *goldsmith.File) bool {
	return false
}

func (s *static) Initialize(ctx goldsmith.Context) error {
	var paths []string
	err := filepath.Walk(s.src, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			paths = append(paths, path)
		}

		return err
	})

	if err != nil {
		return err
	}

	for _, path := range paths {
		srcRelPath, err := filepath.Rel(s.src, path)
		if err != nil {
			panic(err)
		}

		dstRelPath := filepath.Join(s.dst, srcRelPath)
		file := goldsmith.NewFile(dstRelPath)

		var f *os.File
		if f, file.Err = os.Open(path); file.Err == nil {
			_, file.Err = file.Buff.ReadFrom(f)
			f.Close()
		}

		ctx.AddFile(file)
	}

	return nil
}

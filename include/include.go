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

package include

import (
	"os"
	"path/filepath"

	"github.com/FooSoft/goldsmith"
)

type include struct {
	src, dst string
}

func New(src, dst string) goldsmith.Plugin {
	return &include{src, dst}
}

func (*include) Name() string {
	return "include"
}

func (i *include) Initialize(ctx goldsmith.Context) error {
	var paths []string
	err := filepath.Walk(i.src, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			paths = append(paths, path)
		}

		return err
	})

	if err != nil {
		return err
	}

	for _, path := range paths {
		srcRelPath, err := filepath.Rel(i.src, path)
		if err != nil {
			panic(err)
		}

		dstRelPath := filepath.Join(i.dst, srcRelPath)
		f, err := goldsmith.NewFileFromAsset(dstRelPath, path)
		if err != nil {
			return err
		}

		ctx.DispatchFile(f)
	}

	return nil
}

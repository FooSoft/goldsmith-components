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

package markdown

import (
	"bytes"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/russross/blackfriday"
)

type MarkdownType int

const (
	MarkdownCommon MarkdownType = iota
	MarkdownBasic
)

type markdown struct {
	mdType MarkdownType
}

func NewCommon() (goldsmith.Chainer, error) {
	return &markdown{MarkdownCommon}, nil
}

func NewBasic() (goldsmith.Chainer, error) {
	return &markdown{MarkdownBasic}, nil
}

func (*markdown) Accept(file *goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(file.Path)) {
	case ".md", ".markdown":
		return true
	default:
		return false
	}
}

func (m *markdown) Chain(ctx goldsmith.Context, input, output chan *goldsmith.File) {
	var wg sync.WaitGroup

	defer func() {
		wg.Wait()
		close(output)
	}()

	for file := range input {
		wg.Add(1)
		go func(f *goldsmith.File) {
			defer func() {
				output <- f
				wg.Done()
			}()

			var data []byte
			switch m.mdType {
			case MarkdownCommon:
				data = blackfriday.MarkdownCommon(f.Buff.Bytes())
			case MarkdownBasic:
				data = blackfriday.MarkdownBasic(f.Buff.Bytes())
			}

			f.Buff = *bytes.NewBuffer(data)
			f.Path = strings.TrimSuffix(f.Path, path.Ext(f.Path)) + ".html"
		}(file)
	}
}

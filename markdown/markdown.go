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

func NewCommon() goldsmith.Plugin {
	return &markdown{MarkdownCommon}
}

func NewBasic() goldsmith.Plugin {
	return &markdown{MarkdownBasic}
}

func (*markdown) Accept(file *goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(file.Path)) {
	case ".md", ".markdown":
		return true
	default:
		return false
	}
}

func (m *markdown) Process(ctx goldsmith.Context, file *goldsmith.File) bool {
	var data []byte
	switch m.mdType {
	case MarkdownCommon:
		data = blackfriday.MarkdownCommon(file.Buff.Bytes())
	case MarkdownBasic:
		data = blackfriday.MarkdownBasic(file.Buff.Bytes())
	}

	file.Buff = *bytes.NewBuffer(data)
	file.Path = strings.TrimSuffix(file.Path, path.Ext(file.Path)) + ".html"

	return true
}

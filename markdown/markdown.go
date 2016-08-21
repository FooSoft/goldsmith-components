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
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/russross/blackfriday"
)

type mdType int

const (
	mdCommon mdType = iota
	mdBasic
)

type markdown struct {
	mdType mdType
}

func NewCommon() goldsmith.Plugin {
	return &markdown{mdCommon}
}

func NewBasic() goldsmith.Plugin {
	return &markdown{mdBasic}
}

func (*markdown) Name() string {
	return "markdown"
}

func (*markdown) Initialize(ctx goldsmith.Context) ([]string, error) {
	return []string{"**/*.md", "**/*.markdown"}, nil
}

func (m *markdown) Process(ctx goldsmith.Context, f goldsmith.File) error {
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(f); err != nil {
		return err
	}

	var data []byte
	switch m.mdType {
	case mdCommon:
		data = blackfriday.MarkdownCommon(buff.Bytes())
	case mdBasic:
		data = blackfriday.MarkdownBasic(buff.Bytes())
	}

	name := strings.TrimSuffix(f.Path(), path.Ext(f.Path())) + ".html"
	nf := goldsmith.NewFileFromData(name, data)
	nf.CopyValues(f)
	ctx.DispatchFile(nf)

	return nil
}

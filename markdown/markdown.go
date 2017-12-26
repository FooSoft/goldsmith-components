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

type summary struct {
	Title string
	Intro string
}

type wrapper struct {
	blackfriday.Renderer
	summary summary
}

type markdown struct {
	htmlFlags     int
	markdownFlags int
	summaryKey    string
}

func New() *markdown {
	htmlFlags := blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_DASHES |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES

	markdownFlags := blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
		blackfriday.EXTENSION_DEFINITION_LISTS

	return &markdown{htmlFlags, markdownFlags, "summary"}
}

func (m *markdown) HtmlFlags(flags int) *markdown {
	m.htmlFlags = flags
	return m
}

func (m *markdown) MarkdownFlags(flags int) *markdown {
	m.markdownFlags = flags
	return m
}

func (m *markdown) SummaryKey(key string) *markdown {
	m.summaryKey = key
	return m
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

	wrapper := &wrapper{
		Renderer: blackfriday.HtmlRenderer(m.htmlFlags, "", ""),
	}

	data := blackfriday.Markdown(buff.Bytes(), wrapper, m.markdownFlags)
	name := strings.TrimSuffix(f.Path(), path.Ext(f.Path())) + ".html"

	nf := goldsmith.NewFileFromData(name, data)
	nf.CopyValues(f)
	nf.SetValue(m.summaryKey, wrapper.summary)
	ctx.DispatchFile(nf)

	return nil
}

func (w *wrapper) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	if len(w.summary.Title) == 0 && level == 1 {
		marker := out.Len()
		if text() {
			w.summary.Title = string(out.Bytes()[marker:out.Len()])
		}
		out.Truncate(marker)
	}

	w.Renderer.Header(out, text, level, id)
}

func (w *wrapper) Paragraph(out *bytes.Buffer, text func() bool) {
	if len(w.summary.Intro) == 0 {
		marker := out.Len()
		if text() {
			w.summary.Intro = string(out.Bytes()[marker:out.Len()])
		}
		out.Truncate(marker)
	}

	w.Renderer.Paragraph(out, text)
}

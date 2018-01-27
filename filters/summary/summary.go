/*
 * Copyright (c) 2018 Alex Yatskov <alex@foosoft.net>
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

package summary

import (
	"html/template"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/PuerkitoBio/goquery"
)

type summary struct {
	key         string
	titlePath   string
	summaryPath string
}

func New() *summary {
	return &summary{
		key:         "Summary",
		titlePath:   "h1",
		summaryPath: "p",
	}
}

func (s *summary) Key(key string) *summary {
	s.key = key
	return s
}

func (s *summary) TitlePath(path string) *summary {
	s.titlePath = path
	return s
}

func (s *summary) SummaryPath(path string) *summary {
	s.summaryPath = path
	return s
}

func (*summary) Name() string {
	return "summary"
}

func (*summary) Initialize(ctx goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (s *summary) Process(ctx goldsmith.Context, f goldsmith.File) error {
	defer ctx.DispatchFile(f)

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return err
	}

	meta := make(map[string]template.HTML)
	if match := doc.Find(s.titlePath); match.Length() > 0 {
		if html, err := match.Html(); err == nil {
			meta["Title"] = template.HTML(html)
		}
	}

	if match := doc.Find(s.summaryPath); match.Length() > 0 {
		if html, err := match.Html(); err == nil {
			meta["Summary"] = template.HTML(html)
		}
	}

	f.SetValue(s.key, meta)
	return nil
}

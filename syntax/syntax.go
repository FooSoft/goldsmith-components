/*
 * Copyright (c) 2016 Alex Yatskov <alex@foosoft.net>
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

package syntax

import (
	"bytes"
	"errors"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type syntax struct {
	style   string
	numbers bool
}

func New() goldsmith.Plugin {
	return &syntax{
		style:   "github",
		numbers: false,
	}
}

func (s *syntax) Style(style string) *syntax {
	s.style = style
	return s
}

func (s *syntax) LineNumbers(numbers bool) *syntax {
	s.numbers = numbers
	return s
}

func (*syntax) Name() string {
	return "syntax"
}

func (s *syntax) Initialize(ctx goldsmith.Context) ([]string, error) {
	return []string{"**/*.html", "**/*.htm"}, nil
}

func (s *syntax) Process(ctx goldsmith.Context, f goldsmith.File) error {
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return err
	}

	var errs []error
	doc.Find("[class*=language-]").Each(func(i int, sel *goquery.Selection) {
		segs := strings.SplitN(sel.AttrOr("class", ""), "-", 2)
		if len(segs) < 2 {
			errs = append(errs, errors.New("failed to determine language"))
			return
		}

		lexer := lexers.Get(segs[1])
		if lexer == nil {
			lexer = lexers.Fallback
		}

		iterator, err := lexer.Tokenise(nil, sel.Text())
		if err != nil {
			errs = append(errs, err)
			return
		}

		style := styles.Get(s.style)
		if style == nil {
			style = styles.Fallback
		}

		var options []html.Option
		if s.numbers {
			options = append(options, html.WithLineNumbers())
		}

		formatter := html.New(options...)
		var buff bytes.Buffer
		if err := formatter.Format(&buff, style, iterator); err != nil {
			errs = append(errs, err)
			return
		}

		sel.SetHtml(string(buff.Bytes()))
	})

	if len(errs) > 0 {
		return errs[0]
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	nf := goldsmith.NewFileFromData(f.Path(), []byte(html))
	ctx.DispatchFile(nf)

	return nil
}

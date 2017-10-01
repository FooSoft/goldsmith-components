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
	"fmt"

	"github.com/FooSoft/goldsmith"
	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type Placement int

const (
	PlaceInside Placement = iota
	PlaceInline
	PlaceOuter
)

type syntax struct {
	style     string
	numbers   bool
	prefix    string
	placement Placement
}

func New() *syntax {
	return &syntax{
		style:     "github",
		numbers:   false,
		prefix:    "language-",
		placement: PlaceInside,
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

func (s *syntax) Prefix(prefix string) *syntax {
	s.prefix = prefix
	return s
}

func (s *syntax) Placement(placement Placement) *syntax {
	s.placement = placement
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
	doc.Find(fmt.Sprintf("[class*=%s]", s.prefix)).Each(func(i int, sel *goquery.Selection) {
		class := sel.AttrOr("class", "")
		language := class[len(s.prefix):len(class)]
		lexer := lexers.Get(language)
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

		html := string(buff.Bytes())
		switch s.placement {
		case PlaceInside:
			sel.SetHtml(html)
		case PlaceInline:
			sel.ReplaceWithHtml(html)
		case PlaceOuter:
			sel.Closest("pre").ReplaceWithHtml(html)
		}
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

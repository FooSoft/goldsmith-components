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

package abs

import (
	"fmt"
	"net/url"

	"github.com/FooSoft/goldsmith"
	"github.com/PuerkitoBio/goquery"
)

type abs struct {
	attrs   []string
	baseUrl *url.URL
}

func New() *abs {
	return &abs{attrs: []string{"href", "src"}}
}

func (a *abs) BaseUrl(root string) *abs {
	a.baseUrl, _ = url.Parse(root)
	return a
}

func (a *abs) Attr(attrs ...string) *abs {
	a.attrs = attrs
	return a
}

func (*abs) Name() string {
	return "abs"
}

func (*abs) Initialize(ctx goldsmith.Context) ([]string, error) {
	return []string{"**/*.html", "**/*.htm"}, nil
}

func (a *abs) Process(ctx goldsmith.Context, f goldsmith.File) error {
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return err
	}

	for _, attr := range a.attrs {
		path := fmt.Sprintf("*[%s]", attr)
		doc.Find(path).Each(func(index int, sel *goquery.Selection) {
			baseUrl, err := url.Parse(f.Path())
			val, _ := sel.Attr(attr)

			currUrl, err := url.Parse(val)
			if err != nil || currUrl.IsAbs() {
				return
			}

			currUrl = baseUrl.ResolveReference(currUrl)
			if a.baseUrl != nil {
				currUrl = a.baseUrl.ResolveReference(currUrl)
			}

			sel.SetAttr(attr, currUrl.String())
		})
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	nf := goldsmith.NewFileFromData(f.Path(), []byte(html))
	nf.CopyValues(f)
	ctx.DispatchFile(nf)

	return nil
}

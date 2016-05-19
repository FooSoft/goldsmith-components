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
	"path/filepath"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/PuerkitoBio/goquery"
)

type abs struct {
	attrs   []string
	rootUrl *url.URL
}

func New() goldsmith.Plugin {
	return &abs{attrs: []string{"href", "src"}}
}

func NewCustom(root string, attrs []string) goldsmith.Plugin {
	rootUrl, _ := url.Parse(root)
	return &abs{attrs, rootUrl}
}

func (*abs) Accept(ctx goldsmith.Context, f goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(f.Path())) {
	case ".html", ".htm":
		return true
	default:
		return false
	}
}

func (a *abs) Process(ctx goldsmith.Context, f goldsmith.File) error {
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return err
	}

	for _, attr := range a.attrs {
		path := fmt.Sprintf("*[%s]", attr)
		doc.Find(path).Each(func(index int, sel *goquery.Selection) {
			val, _ := sel.Attr(attr)

			currUrl, err := url.Parse(val)
			if err != nil || currUrl.IsAbs() {
				return
			}

			baseUrl, err := url.Parse(filepath.Base(f.Path()))
			currUrl = baseUrl.ResolveReference(currUrl)
			if a.rootUrl != nil {
				currUrl = a.rootUrl.ResolveReference(currUrl)
			}

			sel.SetAttr(attr, currUrl.String())
		})
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	nf := goldsmith.NewFileFromData(f.Path(), []byte(html))
	ctx.DispatchFile(nf)

	return nil
}

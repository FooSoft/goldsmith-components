// Copyright (c) 2016-2018 Alex Yatskov <alex@foosoft.net>
//
// Abs converts relative file references in HTML documents to absolute paths.
package abs

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/PuerkitoBio/goquery"
)

type Abs interface {
	// BaseUrl sets the base path to which relative URLs are joined; the
	// default empty string implies server root path.
	BaseUrl(root string) Abs

	// Attrs sets the attributes which are scanned for relative URLs; the
	// default set includes "href" and "src" attributes.
	Attrs(attrs ...string) Abs

	Name() string
	Initialize(ctx goldsmith.Context) ([]goldsmith.Filter, error)
	Process(ctx goldsmith.Context, f goldsmith.File) error
}

// New creates a new instance of the Abs plugin.
func New() Abs {
	return &abs{attrs: []string{"href", "src"}}
}

type abs struct {
	attrs   []string
	baseUrl *url.URL
}

func (a *abs) BaseUrl(root string) Abs {
	a.baseUrl, _ = url.Parse(root)
	return a
}

func (a *abs) Attrs(attrs ...string) Abs {
	a.attrs = attrs
	return a
}

func (*abs) Name() string {
	return "abs"
}

func (*abs) Initialize(ctx goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
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
			if err != nil {
				return
			}

			if !currUrl.IsAbs() {
				currUrl = baseUrl.ResolveReference(currUrl)
			}
			if a.baseUrl != nil {
				currUrl.Path = filepath.Join(a.baseUrl.Path, currUrl.Path)
			}

			sel.SetAttr(attr, currUrl.String())
		})
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	nf := goldsmith.NewFileFromData(f.Path(), []byte(html))
	nf.InheritValues(f)
	ctx.DispatchFile(nf)

	return nil
}

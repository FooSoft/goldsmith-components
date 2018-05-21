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

// Abs chainable plugin context.
type Abs interface {
	// BaseURL sets the base path to which relative URLs are joined. The
	// default value is root ("/").
	BaseURL(root string) Abs

	// Attrs sets the attributes which are scanned for relative URLs. The
	// default attributes include "href" and "src".
	Attrs(attrs ...string) Abs

	// Name implements goldsmith.Plugin.
	Name() string

	// Initialize implements goldsmith.Initializer.
	Initialize(ctx goldsmith.Context) ([]goldsmith.Filter, error)

	// Process implements goldsmith.Processor.
	Process(ctx goldsmith.Context, f goldsmith.File) error
}

// New creates a new instance of the Abs plugin.
func New() Abs {
	return &abs{attrs: []string{"href", "src"}}
}

type abs struct {
	attrs   []string
	baseURL *url.URL
}

func (a *abs) BaseURL(root string) Abs {
	a.baseURL, _ = url.Parse(root)
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
			baseURL, err := url.Parse(f.Path())
			val, _ := sel.Attr(attr)

			currURL, err := url.Parse(val)
			if err != nil {
				return
			}

			if !currURL.IsAbs() {
				currURL = baseURL.ResolveReference(currURL)
			}
			if a.baseURL != nil {
				currURL.Path = filepath.Join(a.baseURL.Path, currURL.Path)
			}

			sel.SetAttr(attr, currURL.String())
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

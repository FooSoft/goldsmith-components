// Package abs converts relative file references in HTML documents to absolute paths.
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
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor

	// BaseURL sets the base path to which relative URLs are joined (default: "/").
	BaseURL(root string) Abs

	// Attrs sets the attributes which are scanned for relative URLs (default: "href", "src").
	Attrs(attrs ...string) Abs
}

// New creates a new instance of the Abs plugin.
func New() Abs {
	return &absPlugin{attrs: []string{"href", "src"}}
}

type absPlugin struct {
	attrs   []string
	baseURL *url.URL
}

func (plugin *absPlugin) BaseURL(root string) Abs {
	plugin.baseURL, _ = url.Parse(root)
	return plugin
}

func (plugin *absPlugin) Attrs(attrs ...string) Abs {
	plugin.attrs = attrs
	return plugin
}

func (*absPlugin) Name() string {
	return "abs"
}

func (*absPlugin) Initialize(context goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (plugin *absPlugin) Process(context goldsmith.Context, f goldsmith.File) error {
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return err
	}

	for _, attr := range plugin.attrs {
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
			if plugin.baseURL != nil {
				currURL.Path = filepath.Join(plugin.baseURL.Path, currURL.Path)
			}

			sel.SetAttr(attr, currURL.String())
		})
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	nf := goldsmith.NewFileFromData(f.Path(), []byte(html), f.ModTime())
	nf.InheritValues(f)
	context.DispatchFile(nf)

	return nil
}

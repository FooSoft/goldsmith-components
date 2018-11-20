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

	// BaseUrl sets the base path to which relative URLs are joined (default: "/").
	BaseUrl(root string) Abs

	// Attributes sets the attributes which are scanned for relative URLs (default: "href", "src").
	Attributes(attributes ...string) Abs
}

// New creates a new instance of the Abs plugin.
func New() Abs {
	return &abs{attributes: []string{"href", "src"}}
}

type abs struct {
	attributes []string
	baseUrl    *url.URL
}

func (a *abs) BaseUrl(root string) Abs {
	a.baseUrl, _ = url.Parse(root)
	return a
}

func (a *abs) Attributes(attrs ...string) Abs {
	a.attributes = attrs
	return a
}

func (*abs) Name() string {
	return "abs"
}

func (*abs) Initialize(context *goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (a *abs) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	if outputFile := context.RetrieveCachedFile(inputFile.Path(), inputFile); outputFile != nil {
		context.DispatchFile(outputFile)
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(inputFile)
	if err != nil {
		return err
	}

	for _, attribute := range a.attributes {
		cssPath := fmt.Sprintf("*[%s]", attribute)
		doc.Find(cssPath).Each(func(index int, sel *goquery.Selection) {
			baseUrl, err := url.Parse(inputFile.Path())
			value, _ := sel.Attr(attribute)

			currUrl, err := url.Parse(value)
			if err != nil {
				return
			}

			if !currUrl.IsAbs() {
				currUrl = baseUrl.ResolveReference(currUrl)
			}
			if a.baseUrl != nil {
				currUrl.Path = filepath.Join(a.baseUrl.Path, currUrl.Path)
			}

			sel.SetAttr(attribute, currUrl.String())
		})
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	outputFile := goldsmith.NewFileFromData(inputFile.Path(), []byte(html))
	outputFile.InheritValues(inputFile)
	context.DispatchAndCacheFile(outputFile)

	return nil
}

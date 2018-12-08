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

	// BaseUrl sets the base directory to which relative URLs are joined (default: "/").
	BaseUrl(baseDir string) Abs

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

func (a *abs) BaseUrl(baseDir string) Abs {
	a.baseUrl, _ = url.Parse(baseDir)
	return a
}

func (a *abs) Attributes(attributes ...string) Abs {
	a.attributes = attributes
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
		outputFile.InheritValues(inputFile)
		context.DispatchFile(outputFile)
		return nil
	}

	baseUrl, err := url.Parse(inputFile.Path())
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(inputFile)
	if err != nil {
		return err
	}

	for _, attribute := range a.attributes {
		cssPath := fmt.Sprintf("*[%s]", attribute)
		doc.Find(cssPath).Each(func(index int, selection *goquery.Selection) {
			value, exists := selection.Attr(attribute)
			if !exists {
				return
			}

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

			selection.SetAttr(attribute, currUrl.String())
		})
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	outputFile := goldsmith.NewFileFromData(inputFile.Path(), []byte(html))
	outputFile.InheritValues(inputFile)
	context.DispatchAndCacheFile(outputFile, inputFile)
	return nil
}

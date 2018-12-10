// Package absolute converts relative file references in HTML documents to absolute paths.
package absolute

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/PuerkitoBio/goquery"
)

// Absolute chainable plugin context.
type Absolute interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor

	// BaseUrl sets the base directory to which relative URLs are joined (default: "/").
	BaseUrl(baseDir string) Absolute

	// Attributes sets the attributes which are scanned for relative URLs (default: "href", "src").
	Attributes(attributes ...string) Absolute
}

// New creates absolute new instance of the Abs plugin.
func New() Absolute {
	return &absolute{attributes: []string{"href", "src"}}
}

type absolute struct {
	attributes []string
	baseUrl    *url.URL
}

func (absolute *absolute) BaseUrl(baseDir string) Absolute {
	absolute.baseUrl, _ = url.Parse(baseDir)
	return absolute
}

func (absolute *absolute) Attributes(attributes ...string) Absolute {
	absolute.attributes = attributes
	return absolute
}

func (*absolute) Name() string {
	return "absolute"
}

func (*absolute) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return extension.New(".html", ".htm"), nil
}

func (absolute *absolute) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	if outputFile := context.RetrieveCachedFile(inputFile.Path(), inputFile); outputFile != nil {
		outputFile.Meta = inputFile.Meta
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

	for _, attribute := range absolute.attributes {
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

			if absolute.baseUrl != nil {
				currUrl.Path = filepath.Join(absolute.baseUrl.Path, currUrl.Path)
			}

			selection.SetAttr(attribute, currUrl.String())
		})
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	outputFile := context.CreateFileFromData(inputFile.Path(), []byte(html))
	outputFile.Meta = inputFile.Meta
	context.DispatchAndCacheFile(outputFile, inputFile)
	return nil
}

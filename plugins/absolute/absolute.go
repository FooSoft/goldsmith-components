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
type Absolute struct {
	attributes []string
	baseUrl    *url.URL
}

// New creates absolute new instance of the Abs plugin.
func New() *Absolute {
	return &Absolute{attributes: []string{"href", "src"}}
}

// BaseUrl sets the base directory to which relative URLs are joined (default: "/").
func (absolute *Absolute) BaseUrl(baseDir string) *Absolute {
	absolute.baseUrl, _ = url.Parse(baseDir)
	return absolute
}

// Attributes sets the attributes which are scanned for relative URLs (default: "href", "src").
func (absolute *Absolute) Attributes(attributes ...string) *Absolute {
	absolute.attributes = attributes
	return absolute
}

func (*Absolute) Name() string {
	return "absolute"
}

func (*Absolute) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return extension.New(".html", ".htm"), nil
}

func (plugin *Absolute) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
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

	for _, attribute := range plugin.attributes {
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

			if plugin.baseUrl != nil {
				currUrl.Path = filepath.Join(plugin.baseUrl.Path, currUrl.Path)
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

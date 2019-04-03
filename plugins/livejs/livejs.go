// Package livejs injects code to reload the current page when it is modified.
package livejs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/PuerkitoBio/goquery"
)

// LiveJs chainable context.
type LiveJs struct {
	html string
}

// New creates a new instance of the LiveJs plugin.
func New() *LiveJs {
	return new(LiveJs)
}

func (*LiveJs) Name() string {
	return "livejs"
}

func (plugin *LiveJs) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("unable to get livejs path")
	}

	baseDir := path.Dir(filename)
	jsPath := path.Join(baseDir, "live.js")

	js, err := ioutil.ReadFile(jsPath)
	if err != nil {
		return nil, err
	}

	plugin.html = fmt.Sprintf("\n<!-- begin livejs code -->\n<script>\n%s\n</script>\n<!-- end livejs code -->\n", js)
	return wildcard.New("**/*.html", "**/*.htm"), nil
}

func (plugin *LiveJs) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	if outputFile := context.RetrieveCachedFile(inputFile.Path(), inputFile); outputFile != nil {
		outputFile.Meta = inputFile.Meta
		context.DispatchFile(outputFile)
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(inputFile)
	if err != nil {
		return err
	}

	doc.Find("body").AppendHtml(plugin.html)

	html, err := doc.Html()
	if err != nil {
		return err
	}

	outputFile := context.CreateFileFromData(inputFile.Path(), []byte(html))
	outputFile.Meta = inputFile.Meta
	context.DispatchAndCacheFile(outputFile)
	return nil
}

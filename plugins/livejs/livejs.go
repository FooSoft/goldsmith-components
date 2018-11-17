// Package livejs injects code to reload the current page when it is modified.
package livejs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/PuerkitoBio/goquery"
)

// LiveJs chainable context.
type LiveJs interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor
}

// New creates a new instance of the LiveJs plugin.
func New() LiveJs {
	return new(livejs)
}

type livejs struct {
	js string
}

func (*livejs) Name() string {
	return "livejs"
}

func (l *livejs) Initialize(ctx *goldsmith.Context) ([]goldsmith.Filter, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("unable to get livejs path")
	}

	baseDir := path.Dir(filename)
	jsPath := path.Join(baseDir, "live.js")

	data, err := ioutil.ReadFile(jsPath)
	if err != nil {
		return nil, err
	}

	l.js = fmt.Sprintf("\n<!-- begin livejs code -->\n<script>\n%s\n</script>\n<!-- end livejs code -->\n", data)
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (l *livejs) Process(ctx *goldsmith.Context, f *goldsmith.File) error {
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return err
	}

	doc.Find("body").AppendHtml(l.js)

	html, err := doc.Html()
	if err != nil {
		return err
	}

	nf := goldsmith.NewFileFromData(f.Path(), []byte(html))
	ctx.DispatchFile(nf)

	return nil
}

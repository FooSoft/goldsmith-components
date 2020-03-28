// Package markdown renders Markdown documents to HTML with the "goldmark"
// processor. Note that unlike other static site generators, Markdown processing
// does not automatically parse frontmatter; you will need to use the "frontmatter"
// plugin to extract any metadata which may be present in your source content.
package markdown

import (
	"bytes"
	"path"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Markdown chainable context.
type Markdown struct {
	md goldmark.Markdown
}

// New creates a new instance of the Markdown plugin.
func New() *Markdown {
	return new(Markdown)
}

// WithGoldmark allows you to provide your own instance of Goldmark with
// a custom selection of plugins. The default includes GFM and Typographer.
func (plugin *Markdown) WithGoldmark(md goldmark.Markdown) *Markdown {
	plugin.md = md
	return plugin
}

func (*Markdown) Name() string {
	return "markdown"
}

func (plugin *Markdown) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.md", "**/*.markdown"), nil
}

func (plugin *Markdown) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	md := plugin.md
	if md == nil {
		plugin.md = goldmark.New(
			goldmark.WithExtensions(extension.GFM, extension.Typographer),
			goldmark.WithParserOptions(parser.WithAutoHeadingID()),
			goldmark.WithRendererOptions(html.WithUnsafe()),
		)
		md = plugin.md
	}


	outputPath := strings.TrimSuffix(inputFile.Path(), path.Ext(inputFile.Path())) + ".html"
	if outputFile := context.RetrieveCachedFile(outputPath, inputFile); outputFile != nil {
		outputFile.Meta = inputFile.Meta
		context.DispatchFile(outputFile)
		return nil
	}

	var dataIn bytes.Buffer
	if _, err := dataIn.ReadFrom(inputFile); err != nil {
		return err
	}

	var dataOut bytes.Buffer
	if err := md.Convert(dataIn.Bytes(), &dataOut); err != nil {
		return err
	}

	outputFile := context.CreateFileFromData(outputPath, dataOut.Bytes())
	outputFile.Meta = inputFile.Meta
	context.DispatchAndCacheFile(outputFile, inputFile)
	return nil
}

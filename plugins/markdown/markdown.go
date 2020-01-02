// Package markdown renders Markdown documents to HTML with the "blackfriday"
// processor. You can specify which Markdown extensions and HTML features to
// use by directly passing the blackfriday flags to this plugin. Note that
// unlike other static site generators, Markdown processing does not
// automatically parse frontmatter; you will need to use the "frontmatter"
// plugin to extract any metadata which may be present in your source content.
package markdown

import (
	"bytes"
	"path"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/russross/blackfriday"
)

// Markdown chainable context.
type Markdown struct {
	htmlFlags     blackfriday.HTMLFlags
	markdownFlags blackfriday.Extensions
}

// New creates a new instance of the Markdown plugin.
func New() *Markdown {
	htmlFlags := blackfriday.UseXHTML |
		blackfriday.Smartypants |
		blackfriday.SmartypantsFractions |
		blackfriday.SmartypantsDashes |
		blackfriday.SmartypantsLatexDashes

	markdownFlags := blackfriday.NoIntraEmphasis |
		blackfriday.Tables |
		blackfriday.FencedCode |
		blackfriday.Autolink |
		blackfriday.Strikethrough |
		blackfriday.SpaceHeadings |
		blackfriday.HeadingIDs |
		blackfriday.BackslashLineBreak |
		blackfriday.DefinitionLists

	return &Markdown{htmlFlags: htmlFlags, markdownFlags: markdownFlags}
}

// HtmlFlags sets the HTML flags used by the blackfriday markdown processor;
// see https://github.com/russross/blackfriday/blob/master/html.go for options.
func (plugin *Markdown) HtmlFlags(flags blackfriday.HTMLFlags) *Markdown {
	plugin.htmlFlags = flags
	return plugin
}

// MarkdownFlags sets the markdown flags used by the blackfriday markdown processor;
// see https://github.com/russross/blackfriday/blob/master/markdown.go for options.
func (plugin *Markdown) MarkdownFlags(flags blackfriday.Extensions) *Markdown {
	plugin.markdownFlags = flags
	return plugin
}

func (*Markdown) Name() string {
	return "markdown"
}

func (plugin *Markdown) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.md", "**/*.markdown"), nil
}

func (plugin *Markdown) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	outputPath := strings.TrimSuffix(inputFile.Path(), path.Ext(inputFile.Path())) + ".html"
	if outputFile := context.RetrieveCachedFile(outputPath, inputFile); outputFile != nil {
		outputFile.Meta = inputFile.Meta
		context.DispatchFile(outputFile)
		return nil
	}

	var buff bytes.Buffer
	if _, err := buff.ReadFrom(inputFile); err != nil {
		return err
	}

	var (
		renderer = blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{Flags: plugin.htmlFlags})
		data     = blackfriday.Run(buff.Bytes(), blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(plugin.markdownFlags))
	)

	outputFile := context.CreateFileFromData(outputPath, data)
	outputFile.Meta = inputFile.Meta
	context.DispatchAndCacheFile(outputFile, inputFile)
	return nil
}

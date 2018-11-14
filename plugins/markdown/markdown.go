// Package markdown renders markdown documents to HTML with the blackfriday markdown processor.
package markdown

import (
	"bytes"
	"path"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/russross/blackfriday"
)

// Markdown chainable context.
type Markdown interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor

	// HTMLFlags sets the HTML flags used by the blackfriday markdown processor;
	// see https://github.com/russross/blackfriday/blob/master/html.go for options.
	HTMLFlags(flags int) Markdown

	// MarkdownFlags sets the markdown flags used by the blackfriday markdown processor;
	// see https://github.com/russross/blackfriday/blob/master/markdown.go for options.
	MarkdownFlags(flags int) Markdown
}

// New creates a new instance of the Markdown plugin.
func New() Markdown {
	htmlFlags := blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_DASHES |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES

	markdownFlags := blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
		blackfriday.EXTENSION_DEFINITION_LISTS

	return &markdown{htmlFlags: htmlFlags, markdownFlags: markdownFlags}
}

type markdown struct {
	htmlFlags     int
	markdownFlags int
}

func (m *markdown) HTMLFlags(flags int) Markdown {
	m.htmlFlags = flags
	return m
}

func (m *markdown) MarkdownFlags(flags int) Markdown {
	m.markdownFlags = flags
	return m
}

func (*markdown) Name() string {
	return "markdown"
}

func (*markdown) Initialize(ctx *goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".md", ".markdown")}, nil
}

func (m *markdown) Process(ctx *goldsmith.Context, f *goldsmith.File) error {
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(f); err != nil {
		return err
	}

	renderer := blackfriday.HtmlRenderer(m.htmlFlags, "", "")
	data := blackfriday.Markdown(buff.Bytes(), renderer, m.markdownFlags)
	name := strings.TrimSuffix(f.Path(), path.Ext(f.Path())) + ".html"

	nf := goldsmith.NewFileFromData(name, data, f.ModTime())
	nf.InheritValues(f)
	ctx.DispatchFile(nf)
	ctx.CacheFile(f, nf)

	return nil
}

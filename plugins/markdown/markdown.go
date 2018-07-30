package markdown

import (
	"bytes"
	"path"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/russross/blackfriday"
)

type Markdown interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor

	HtmlFlags(flags int) Markdown
	MarkdownFlags(flags int) Markdown
}

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

func (m *markdown) HtmlFlags(flags int) Markdown {
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

func (*markdown) Initialize(ctx goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".md", ".markdown")}, nil
}

func (m *markdown) Process(ctx goldsmith.Context, f goldsmith.File) error {
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(f); err != nil {
		return err
	}

	renderer := blackfriday.HtmlRenderer(m.htmlFlags, "", "")
	data := blackfriday.Markdown(buff.Bytes(), renderer, m.markdownFlags)
	name := strings.TrimSuffix(f.Path(), path.Ext(f.Path())) + ".html"

	nf := goldsmith.NewFileFromData(name, data)
	nf.InheritValues(f)
	ctx.DispatchFile(nf)

	return nil
}

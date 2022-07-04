package markdown

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"foosoft.net/projects/goldsmith"
	"foosoft.net/projects/goldsmith-components/harness"
)

func Test(self *testing.T) {
	harness.Validate(
		self,
		func(gs *goldsmith.Goldsmith) {
			gs.Chain(NewWithGoldmark(goldmark.New(
				goldmark.WithExtensions(extension.GFM, extension.Typographer, extension.DefinitionList),
				goldmark.WithParserOptions(parser.WithAutoHeadingID()),
				goldmark.WithRendererOptions(html.WithUnsafe()),
			)))
		},
	)
}

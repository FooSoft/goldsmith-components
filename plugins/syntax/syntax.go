package syntax

import (
	"bytes"
	"fmt"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type Placement int

const (
	PlaceInside Placement = iota
	PlaceInline
)

type Syntax struct {
	style     string
	numbers   bool
	prefix    string
	placement Placement
}

func New() *Syntax {
	return &Syntax{
		style:     "github",
		numbers:   false,
		prefix:    "language-",
		placement: PlaceInside,
	}
}

func (plugin *Syntax) Style(style string) *Syntax {
	plugin.style = style
	return plugin
}

func (plugin *Syntax) LineNumbers(numbers bool) *Syntax {
	plugin.numbers = numbers
	return plugin
}

func (plugin *Syntax) Prefix(prefix string) *Syntax {
	plugin.prefix = prefix
	return plugin
}

func (plugin *Syntax) Placement(placement Placement) *Syntax {
	plugin.placement = placement
	return plugin
}

func (*Syntax) Name() string {
	return "syntax"
}

func (*Syntax) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return extension.New(".html", ".htm"), nil
}

func (plugin *Syntax) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	if outputFile := context.RetrieveCachedFile(inputFile.Path(), inputFile); outputFile != nil {
		outputFile.Meta = inputFile.Meta
		context.DispatchFile(outputFile)
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(inputFile)
	if err != nil {
		return err
	}

	var errs []error
	doc.Find(fmt.Sprintf("[class*=%s]", plugin.prefix)).Each(func(i int, sel *goquery.Selection) {
		class := sel.AttrOr("class", "")
		language := class[len(plugin.prefix):len(class)]
		lexer := lexers.Get(language)
		if lexer == nil {
			lexer = lexers.Fallback
		}

		iterator, err := lexer.Tokenise(nil, sel.Text())
		if err != nil {
			errs = append(errs, err)
			return
		}

		style := styles.Get(plugin.style)
		if style == nil {
			style = styles.Fallback
		}

		var options []html.Option
		if plugin.numbers {
			options = append(options, html.WithLineNumbers())
		}

		formatter := html.New(options...)
		var buff bytes.Buffer
		if err := formatter.Format(&buff, style, iterator); err != nil {
			errs = append(errs, err)
			return
		}

		switch plugin.placement {
		case PlaceInside:
			sel.SetHtml(string(buff.Bytes()))
		case PlaceInline:
			if docCode, err := goquery.NewDocumentFromReader(&buff); err == nil {
				selPre := docCode.Find("pre")
				if style, exists := selPre.Attr("style"); exists {
					sel.SetAttr("style", style)
				}

				if htmlPre, err := selPre.Html(); err == nil {
					sel.SetHtml(htmlPre)
				}
			}
		}
	})

	if len(errs) > 0 {
		return errs[0]
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

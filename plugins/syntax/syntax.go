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

type Syntax interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor

	Style(style string) Syntax
	LineNumbers(numbers bool) Syntax
	Prefix(prefix string) Syntax
	Placement(placement Placement) Syntax
}

func New() Syntax {
	return &syntax{
		style:     "github",
		numbers:   false,
		prefix:    "language-",
		placement: PlaceInside,
	}
}

type syntax struct {
	style     string
	numbers   bool
	prefix    string
	placement Placement
}

func (s *syntax) Style(style string) Syntax {
	s.style = style
	return s
}

func (s *syntax) LineNumbers(numbers bool) Syntax {
	s.numbers = numbers
	return s
}

func (s *syntax) Prefix(prefix string) Syntax {
	s.prefix = prefix
	return s
}

func (s *syntax) Placement(placement Placement) Syntax {
	s.placement = placement
	return s
}

func (*syntax) Name() string {
	return "syntax"
}

func (*syntax) Initialize(context *goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (s *syntax) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
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
	doc.Find(fmt.Sprintf("[class*=%s]", s.prefix)).Each(func(i int, sel *goquery.Selection) {
		class := sel.AttrOr("class", "")
		language := class[len(s.prefix):len(class)]
		lexer := lexers.Get(language)
		if lexer == nil {
			lexer = lexers.Fallback
		}

		iterator, err := lexer.Tokenise(nil, sel.Text())
		if err != nil {
			errs = append(errs, err)
			return
		}

		style := styles.Get(s.style)
		if style == nil {
			style = styles.Fallback
		}

		var options []html.Option
		if s.numbers {
			options = append(options, html.WithLineNumbers())
		}

		formatter := html.New(options...)
		var buff bytes.Buffer
		if err := formatter.Format(&buff, style, iterator); err != nil {
			errs = append(errs, err)
			return
		}

		switch s.placement {
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

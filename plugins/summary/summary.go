package summary

import (
	"html/template"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/PuerkitoBio/goquery"
)

type Summary interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor

	SummaryKey(key string) Summary
	TitlePath(path string) Summary
	SummaryPath(path string) Summary
}

func New() Summary {
	return &summary{
		summaryKey:  "Summary",
		titlePath:   "h1",
		summaryPath: "p",
	}
}

type summary struct {
	summaryKey  string
	titlePath   string
	summaryPath string
}

func (s *summary) SummaryKey(key string) Summary {
	s.summaryKey = key
	return s
}

func (s *summary) TitlePath(path string) Summary {
	s.titlePath = path
	return s
}

func (s *summary) SummaryPath(path string) Summary {
	s.summaryPath = path
	return s
}

func (*summary) Name() string {
	return "summary"
}

func (*summary) Initialize(context *goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (s *summary) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	if outputFile := context.RetrieveCachedFile(inputFile.Path(), inputFile.Path()); outputFile != nil {
		context.DispatchFile(outputFile, false)
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(inputFile)
	if err != nil {
		return err
	}

	meta := make(map[string]template.HTML)
	if match := doc.Find(s.titlePath); match.Length() > 0 {
		if html, err := match.Html(); err == nil {
			meta["Title"] = template.HTML(html)
		}
	}

	if match := doc.Find(s.summaryPath); match.Length() > 0 {
		if html, err := match.Html(); err == nil {
			meta["Summary"] = template.HTML(html)
		}
	}

	for key, value := range meta {
		inputFile.SetValue(key, value)
	}

	context.DispatchFile(inputFile, true)
	return nil
}

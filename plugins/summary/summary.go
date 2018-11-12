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

	Key(key string) Summary
	TitlePath(path string) Summary
	SummaryPath(path string) Summary
}

func New() Summary {
	return &summary{
		key:         "Summary",
		titlePath:   "h1",
		summaryPath: "p",
	}
}

type summary struct {
	key         string
	titlePath   string
	summaryPath string
}

func (s *summary) Key(key string) Summary {
	s.key = key
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

func (*summary) Initialize(ctx *goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (s *summary) Process(ctx *goldsmith.Context, f *goldsmith.File) error {
	defer ctx.DispatchFile(f)

	doc, err := goquery.NewDocumentFromReader(f)
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

	f.SetValue(s.key, meta)
	return nil
}

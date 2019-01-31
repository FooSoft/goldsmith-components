package summary

import (
	"html/template"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/PuerkitoBio/goquery"
)

type Summary struct {
	summaryKey  string
	titlePath   string
	summaryPath string
}

func New() *Summary {
	return &Summary{
		summaryKey:  "Summary",
		titlePath:   "h1",
		summaryPath: "p",
	}
}

func (plugin *Summary) SummaryKey(key string) *Summary {
	plugin.summaryKey = key
	return plugin
}

func (plugin *Summary) TitlePath(path string) *Summary {
	plugin.titlePath = path
	return plugin
}

func (plugin *Summary) SummaryPath(path string) *Summary {
	plugin.summaryPath = path
	return plugin
}

func (*Summary) Name() string {
	return "summary"
}

func (*Summary) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return extension.New(".html", ".htm"), nil
}

func (plugin *Summary) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	doc, err := goquery.NewDocumentFromReader(inputFile)
	if err != nil {
		return err
	}

	meta := make(map[string]template.HTML)
	if match := doc.Find(plugin.titlePath); match.Length() > 0 {
		if html, err := match.Html(); err == nil {
			meta["Title"] = template.HTML(html)
		}
	}

	if match := doc.Find(plugin.summaryPath); match.Length() > 0 {
		if html, err := match.Html(); err == nil {
			meta["Summary"] = template.HTML(html)
		}
	}

	for key, value := range meta {
		inputFile.Meta[key] = value
	}

	context.DispatchFile(inputFile)
	return nil
}

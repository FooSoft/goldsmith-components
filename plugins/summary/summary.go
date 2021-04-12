// Package summary generates a summary and title for HTML files using CSS
// selectors. This plugin is useful when combined with other plugins such as
// "collection" to create blog post previews.
package summary

import (
	"html/template"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/PuerkitoBio/goquery"
)

// Summary chainable context.
type Summary struct {
	titleKey    string
	summaryKey  string
	titlePath   string
	summaryPath string
}

// New creates a new instance of the Summary plugin.
func New() *Summary {
	return &Summary{
		titleKey:    "Title",
		summaryKey:  "Summary",
		titlePath:   "h1",
		summaryPath: "p",
	}
}

// TitleKey sets the metadata key used to store the file title (default: "Title").
func (plugin *Summary) TitleKey(key string) *Summary {
	plugin.titleKey = key
	return plugin
}

// SummaryKey sets the metadata key used to store the file summary (default: "Summary").
func (plugin *Summary) SummaryKey(key string) *Summary {
	plugin.summaryKey = key
	return plugin
}

// TitlePath sets CSS path used to retrieve the file title (default: "h1").
func (plugin *Summary) TitlePath(path string) *Summary {
	plugin.titlePath = path
	return plugin
}

// SummaryPath sets the CSS path used to retrieve the file summary (default: "p").
func (plugin *Summary) SummaryPath(path string) *Summary {
	plugin.summaryPath = path
	return plugin
}

func (*Summary) Name() string {
	return "summary"
}

func (*Summary) Initialize(context *goldsmith.Context) error {
	context.Filter(wildcard.New("**/*.html", "**/*.htm"))
	return nil
}

func (plugin *Summary) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	doc, err := goquery.NewDocumentFromReader(inputFile)
	if err != nil {
		return err
	}

	meta := make(map[string]template.HTML)
	if match := doc.Find(plugin.titlePath); match.Length() > 0 {
		if html, err := match.Html(); err == nil {
			meta[plugin.titleKey] = template.HTML(html)
		}
	}

	if match := doc.Find(plugin.summaryPath); match.Length() > 0 {
		if html, err := match.Html(); err == nil {
			meta[plugin.summaryKey] = template.HTML(html)
		}
	}

	for key, value := range meta {
		inputFile.Meta[key] = value
	}

	context.DispatchFile(inputFile)
	return nil
}

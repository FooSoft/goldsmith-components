package document

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/harness"
	"github.com/PuerkitoBio/goquery"
)

func process(file *goldsmith.File, doc *goquery.Document) error {
	doc.Find("h1").SetAttr("style", "color: red;")
	return nil
}

func Test(t *testing.T) {
	type Processor func(*goquery.Document) error

	harness.Validate(
		t,
		func(gs *goldsmith.Goldsmith) {
			gs.Chain(New(process))
		},
	)
}

package document

import (
	"testing"

	"foosoft.net/projects/goldsmith"
	"foosoft.net/projects/goldsmith-components/harness"
	"github.com/PuerkitoBio/goquery"
)

func process(file *goldsmith.File, doc *goquery.Document) error {
	doc.Find("h1").SetAttr("style", "color: red;")
	return nil
}

func Test(self *testing.T) {
	type Processor func(*goquery.Document) error

	harness.Validate(
		self,
		func(gs *goldsmith.Goldsmith) {
			gs.Chain(New(process))
		},
	)
}

package rule

import (
	"testing"

	"foosoft.net/projects/goldsmith"
	"foosoft.net/projects/goldsmith-components/harness"
	"foosoft.net/projects/goldsmith-components/plugins/frontmatter"
	"foosoft.net/projects/goldsmith-components/plugins/layout"
)

func Test(self *testing.T) {
	harness.Validate(
		self,
		func(gs *goldsmith.Goldsmith) {
			gs.
				Chain(New()).
				Chain(frontmatter.New()).
				Chain(layout.New().DefaultLayout("page"))
		},
	)
}

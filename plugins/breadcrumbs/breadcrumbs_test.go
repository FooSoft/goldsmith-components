package breadcrumbs

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
				Chain(frontmatter.New()).
				Chain(New()).
				Chain(layout.New())
		},
	)
}

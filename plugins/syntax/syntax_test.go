package syntax

import (
	"testing"

	"foosoft.net/projects/goldsmith"
	"foosoft.net/projects/goldsmith-components/harness"
	"foosoft.net/projects/goldsmith-components/plugins/frontmatter"
	"foosoft.net/projects/goldsmith-components/plugins/layout"
	"foosoft.net/projects/goldsmith-components/plugins/markdown"
)

func Test(self *testing.T) {
	harness.Validate(
		self,
		func(gs *goldsmith.Goldsmith) {
			gs.
				Chain(frontmatter.New()).
				Chain(markdown.New()).
				Chain(New()).
				Chain(layout.New())
		},
	)
}

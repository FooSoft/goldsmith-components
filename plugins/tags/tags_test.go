package tags

import (
	"testing"

	"foosoft.net/projects/goldsmith"
	"foosoft.net/projects/goldsmith-components/harness"
	"foosoft.net/projects/goldsmith-components/plugins/frontmatter"
	"foosoft.net/projects/goldsmith-components/plugins/layout"
	"foosoft.net/projects/goldsmith-components/plugins/markdown"
)

func Test(self *testing.T) {
	meta := map[string]interface{}{
		"Layout": "tag",
	}

	harness.Validate(
		self,
		func(gs *goldsmith.Goldsmith) {
			gs.
				Chain(frontmatter.New()).
				Chain(markdown.New()).
				Chain(New().IndexMeta(meta)).
				Chain(layout.New())
		},
	)
}

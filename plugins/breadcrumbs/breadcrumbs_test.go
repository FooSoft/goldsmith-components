package breadcrumbs

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/harness"
	"github.com/FooSoft/goldsmith-components/plugins/frontmatter"
	"github.com/FooSoft/goldsmith-components/plugins/layout"
	"github.com/FooSoft/goldsmith-components/plugins/markdown"
)

func Test(t *testing.T) {
	harness.Validate(
		t,
		"",
		func(gs *goldsmith.Goldsmith) {
			gs.
				Chain(frontmatter.New()).
				Chain(markdown.New()).
				Chain(New()).
				Chain(layout.New())
		},
	)
}

package pager

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/harness"
	"github.com/FooSoft/goldsmith-components/plugins/collection"
	"github.com/FooSoft/goldsmith-components/plugins/frontmatter"
	"github.com/FooSoft/goldsmith-components/plugins/layout"
)

func Test(t *testing.T) {
	lister := func(file *goldsmith.File) interface{} {
		if groupsRaw, ok := file.Meta["Groups"]; ok {
			if groups, ok := groupsRaw.(map[string][]*goldsmith.File); ok {
				if group, ok := groups["group"]; ok {
					return group
				}
			}
		}

		return nil
	}

	harness.Validate(
		t,
		func(gs *goldsmith.Goldsmith) {
			gs.
				Chain(frontmatter.New()).
				Chain(collection.New()).
				Chain(New(lister).ItemsPerPage(4)).
				Chain(layout.New())
		},
	)
}

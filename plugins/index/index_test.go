package index

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/harness"
	"github.com/FooSoft/goldsmith-components/plugins/layout"
)

func Test(t *testing.T) {
	meta := map[string]interface{}{
		"Layout": "index",
	}

	harness.Validate(
		t,
		func(gs *goldsmith.Goldsmith) {
			gs.
				Chain(New(meta)).
				Chain(layout.New())
		},
	)
}

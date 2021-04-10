package forward

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/operator"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/FooSoft/goldsmith-components/harness"
	"github.com/FooSoft/goldsmith-components/plugins/layout"
)

func Test(t *testing.T) {
	meta := map[string]interface{}{
		"Layout": "forward",
	}

	pathMap := map[string]string{
		"/file_old.html": "/file_new.html",
	}

	harness.Validate(
		t,
		func(gs *goldsmith.Goldsmith) {
			gs.
				FilterPush(operator.Not(wildcard.New("*.gohtml"))).
				Chain(New(meta).PathMap(pathMap)).
				FilterPop().
				Chain(layout.New())
		},
	)
}

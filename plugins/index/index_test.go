package index

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/operator"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/FooSoft/goldsmith-components/harness"
	"github.com/FooSoft/goldsmith-components/plugins/layout"
)

func Test(self *testing.T) {
	props := map[string]interface{}{
		"Layout": "index",
	}

	harness.Validate(
		self,
		func(gs *goldsmith.Goldsmith) {
			gs.
				FilterPush(operator.Not(wildcard.New("*.gohtml"))).
				Chain(New(props)).
				FilterPop().
				Chain(layout.New())
		},
	)
}

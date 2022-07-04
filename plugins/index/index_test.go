package index

import (
	"testing"

	"foosoft.net/projects/goldsmith"
	"foosoft.net/projects/goldsmith-components/filters/operator"
	"foosoft.net/projects/goldsmith-components/filters/wildcard"
	"foosoft.net/projects/goldsmith-components/harness"
	"foosoft.net/projects/goldsmith-components/plugins/layout"
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

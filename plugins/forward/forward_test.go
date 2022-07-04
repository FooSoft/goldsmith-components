package forward

import (
	"testing"

	"foosoft.net/projects/goldsmith"
	"foosoft.net/projects/goldsmith-components/filters/operator"
	"foosoft.net/projects/goldsmith-components/filters/wildcard"
	"foosoft.net/projects/goldsmith-components/harness"
	"foosoft.net/projects/goldsmith-components/plugins/layout"
)

func Test(self *testing.T) {
	meta := map[string]interface{}{
		"Layout": "forward",
	}

	pathMap := map[string]string{
		"file_old.html": "file_new.html",
	}

	harness.Validate(
		self,
		func(gs *goldsmith.Goldsmith) {
			gs.
				FilterPush(operator.Not(wildcard.New("*.gohtml"))).
				Chain(New(meta).PathMap(pathMap)).
				FilterPop().
				Chain(layout.New())
		},
	)
}

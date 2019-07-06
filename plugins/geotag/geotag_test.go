package geotag

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/operator"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/FooSoft/goldsmith-components/harness"
	"github.com/FooSoft/goldsmith-components/plugins/index"
	"github.com/FooSoft/goldsmith-components/plugins/layout"
)

func Test(t *testing.T) {
	meta := map[string]interface{}{
		"Layout": "index",
	}

	lookuper, err := NewLookuperGeonamesFile("testdata/cities500.txt")
	if err != nil {
		t.Fatal(err)
	}

	harness.Validate(
		t,
		func(gs *goldsmith.Goldsmith) {
			gs.
				Chain(New().Lookuper(lookuper)).
				FilterPush(operator.Not(wildcard.New("*.gohtml"))).
				Chain(index.New(meta)).
				FilterPop().
				Chain(layout.New())
		},
	)
}

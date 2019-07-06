package exif

import (
	"os"
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/operator"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/FooSoft/goldsmith-components/harness"
	"github.com/FooSoft/goldsmith-components/plugins/index"
	"github.com/FooSoft/goldsmith-components/plugins/layout"
)

func Test(t *testing.T) {
	fp, err := os.Open("testdata/cities500.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer fp.Close()

	meta := map[string]interface{}{
		"Layout": "index",
	}

	harness.Validate(
		t,
		func(gs *goldsmith.Goldsmith) {
			if _, err := fp.Seek(0, os.SEEK_SET); err != nil {
				t.Fatal(err)
			}

			gs.
				Chain(New().Lookup(fp)).
				FilterPush(operator.Not(wildcard.New("*.gohtml"))).
				Chain(index.New(meta)).
				FilterPop().
				Chain(layout.New())
		},
	)
}

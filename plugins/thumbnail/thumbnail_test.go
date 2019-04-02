package thumbnail

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/harness"
)

func Test(t *testing.T) {
	harness.Validate(
		t,
		func(gs *goldsmith.Goldsmith) {
			gs.Chain(New())
		},
	)
}

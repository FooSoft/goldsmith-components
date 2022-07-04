package wildcard

import (
	"testing"

	"foosoft.net/projects/goldsmith"
	"foosoft.net/projects/goldsmith-components/harness"
)

func Test(t *testing.T) {
	harness.Validate(
		t,
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(New("**/*.txt", "*.md"))
		},
	)
}

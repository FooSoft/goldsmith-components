package condition

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/harness"
)

func TestEnabled(t *testing.T) {
	harness.ValidateCase(
		t,
		"enabled",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(New(true))
		},
	)
}

func TestDisabled(t *testing.T) {
	harness.ValidateCase(
		t,
		"disabled",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(New(false))
		},
	)
}

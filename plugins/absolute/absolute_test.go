package absolute

import (
	"testing"

	"foosoft.net/projects/goldsmith"
	"foosoft.net/projects/goldsmith-components/harness"
)

func Test(self *testing.T) {
	harness.Validate(
		self,
		func(gs *goldsmith.Goldsmith) {
			gs.Chain(New().BaseUrl("https://foosoft.net"))
		},
	)
}

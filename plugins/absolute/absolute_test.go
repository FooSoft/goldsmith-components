package absolute

import (
	"testing"

	"github.com/FooSoft/goldsmith-components/harness"
)

func Test(t *testing.T) {
	harness.Validate(t, "", New())
}

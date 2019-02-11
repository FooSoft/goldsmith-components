package absolute

import (
	"testing"

	"github.com/FooSoft/goldsmith-components/harness"
)

func TestDefault(t *testing.T) {
	harness.Validate(t, "default", New())
}

func TestBaseUrl(t *testing.T) {
	harness.Validate(t, "base_url", New().BaseUrl("/base"))
}

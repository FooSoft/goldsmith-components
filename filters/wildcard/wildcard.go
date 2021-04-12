package wildcard

import (
	"github.com/FooSoft/goldsmith"
	"github.com/bmatcuk/doublestar"
)

type Wildcard struct {
	wildcards []string
}

func New(wildcards ...string) *Wildcard {
	return &Wildcard{wildcards}
}

func (*Wildcard) Name() string {
	return "wildcard"
}

func (filter *Wildcard) Accept(file *goldsmith.File) bool {
	filePath := file.Path()

	for _, wildcard := range filter.wildcards {
		if matched, _ := doublestar.PathMatch(wildcard, filePath); matched {
			return true
		}
	}

	return false
}

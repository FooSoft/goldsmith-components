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

func (filter *Wildcard) Accept(file *goldsmith.File) (bool, error) {
	filePath := file.Path()

	for _, wildcard := range filter.wildcards {
		matched, err := doublestar.PathMatch(wildcard, filePath)
		if err != nil {
			return false, err
		}

		if matched {
			return true, nil
		}
	}

	return false, nil
}

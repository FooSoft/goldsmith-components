package wildcard

import (
	"github.com/FooSoft/goldsmith"
	"github.com/bmatcuk/doublestar"
)

type Wildcard interface {
	goldsmith.Filter
}

func New(wildcards ...string) Wildcard {
	return &wildcard{wildcards}
}

type wildcard struct {
	wildcards []string
}

func (*wildcard) Name() string {
	return "wildcard"
}

func (e *wildcard) Accept(ctx *goldsmith.Context, file *goldsmith.File) (bool, error) {
	filePath := file.Path()

	for _, wildcard := range e.wildcards {
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

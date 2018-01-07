package wildcard

import (
	"github.com/FooSoft/goldsmith"
	"github.com/bmatcuk/doublestar"
)

type wildcard struct {
	wildcards []string
}

func New(wildcards ...string) *wildcard {
	return &wildcard{wildcards}
}

func (*wildcard) Name() string {
	return "wildcard"
}

func (e *wildcard) Accept(ctx goldsmith.Context, f goldsmith.File) (bool, error) {
	filePath := f.Path()

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

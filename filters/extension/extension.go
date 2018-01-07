package extension

import (
	"path/filepath"

	"github.com/FooSoft/goldsmith"
)

type extension struct {
	extensions []string
}

func New(extensions ...string) *extension {
	return &extension{extensions}
}

func (*extension) Name() string {
	return "extension"
}

func (e *extension) Accept(ctx goldsmith.Context, f goldsmith.File) (bool, error) {
	fileExt := filepath.Ext(f.Path())

	for _, extension := range e.extensions {
		if extension == fileExt {
			return true, nil
		}
	}

	return false, nil
}

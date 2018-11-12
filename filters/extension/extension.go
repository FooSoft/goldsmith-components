package extension

import (
	"path/filepath"

	"github.com/FooSoft/goldsmith"
)

type Extension interface {
	goldsmith.Filter
}

func New(extensions ...string) Extension {
	return &extension{extensions}
}

type extension struct {
	extensions []string
}

func (*extension) Name() string {
	return "extension"
}

func (e *extension) Accept(ctx *goldsmith.Context, file *goldsmith.File) (bool, error) {
	fileExt := filepath.Ext(file.Path())

	for _, extension := range e.extensions {
		if extension == fileExt {
			return true, nil
		}
	}

	return false, nil
}

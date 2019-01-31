package extension

import (
	"path/filepath"

	"github.com/FooSoft/goldsmith"
)

type Extension struct {
	extensions []string
}

func New(extensions ...string) *Extension {
	return &Extension{extensions}
}

func (*Extension) Name() string {
	return "extension"
}

func (filter *Extension) Accept(ctx *goldsmith.Context, file *goldsmith.File) (bool, error) {
	fileExt := filepath.Ext(file.Path())

	for _, extension := range filter.extensions {
		if extension == fileExt {
			return true, nil
		}
	}

	return false, nil
}

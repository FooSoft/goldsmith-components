package not

import (
	"github.com/FooSoft/goldsmith"
)

type Not interface {
	goldsmith.Filter
}

func New(filter goldsmith.Filter) Not {
	return &not{filter}
}

type not struct {
	filter goldsmith.Filter
}

func (*not) Name() string {
	return "not"
}

func (n *not) Accept(ctx *goldsmith.Context, file *goldsmith.File) (bool, error) {
	accept, err := n.filter.Accept(ctx, file)
	return !accept, err
}

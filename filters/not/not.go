package not

import (
	"github.com/FooSoft/goldsmith"
)

type not struct {
	filter goldsmith.Filter
}

func New(filter goldsmith.Filter) *not {
	return &not{filter}
}

func (*not) Name() string {
	return "not"
}

func (n *not) Accept(ctx goldsmith.Context, f goldsmith.File) (bool, error) {
	accept, err := n.filter.Accept(ctx, f)
	return !accept, err
}

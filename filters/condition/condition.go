package condition

import (
	"github.com/FooSoft/goldsmith"
)

type Condition struct {
	accept bool
}

func New(accept bool) *Condition {
	return &Condition{accept}
}

func (*Condition) Name() string {
	return "condition"
}

func (filter *Condition) Accept(ctx *goldsmith.Context, file *goldsmith.File) (bool, error) {
	return filter.accept, nil
}

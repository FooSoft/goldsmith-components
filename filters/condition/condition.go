package condition

import (
	"github.com/FooSoft/goldsmith"
)

type Condition interface {
	goldsmith.Filter
}

func New(accept bool) Condition {
	return &condition{accept}
}

type condition struct {
	accept bool
}

func (*condition) Name() string {
	return "condition"
}

func (c *condition) Accept(ctx *goldsmith.Context, file *goldsmith.File) (bool, error) {
	return c.accept, nil
}

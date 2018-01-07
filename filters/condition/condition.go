package condition

import (
	"github.com/FooSoft/goldsmith"
)

type condition struct {
	accept bool
}

func New(accept bool) *condition {
	return &condition{accept}
}

func (*condition) Name() string {
	return "condition"
}

func (c *condition) Accept(ctx goldsmith.Context, f goldsmith.File) (bool, error) {
	return c.accept, nil
}

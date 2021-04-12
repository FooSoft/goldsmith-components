package condition

import (
	"github.com/FooSoft/goldsmith"
)

type Condition struct {
	accept bool
}

func New(accept bool) *Condition {
	return &Condition{accept: accept}
}

func (*Condition) Name() string {
	return "condition"
}

func (filter *Condition) Accept(file *goldsmith.File) bool {
	return filter.accept
}

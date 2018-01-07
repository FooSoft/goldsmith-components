package and

import (
	"github.com/FooSoft/goldsmith"
)

type and struct {
	filters []goldsmith.Filter
}

func New(filters ...goldsmith.Filter) *and {
	return &and{filters}
}

func (*and) Name() string {
	return "and"
}

func (a *and) Accept(ctx goldsmith.Context, f goldsmith.File) (bool, error) {
	for _, filter := range a.filters {
		accept, err := filter.Accept(ctx, f)
		if err != nil {
			return false, err
		}
		if !accept {
			return false, nil
		}
	}

	return true, nil
}

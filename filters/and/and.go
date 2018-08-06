package and

import (
	"github.com/FooSoft/goldsmith"
)

type And interface {
	goldsmith.Filter
}

func New(filters ...goldsmith.Filter) And {
	return &and{filters}
}

type and struct {
	filters []goldsmith.Filter
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

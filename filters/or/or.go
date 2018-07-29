package or

import (
	"github.com/FooSoft/goldsmith"
)

type or struct {
	filters []goldsmith.Filter
}

func New(filters ...goldsmith.Filter) *or {
	return &or{filters}
}

func (*or) Name() string {
	return "or"
}

func (o *or) Accept(ctx goldsmith.Context, f goldsmith.File) (bool, error) {
	for _, filter := range o.filters {
		accept, err := filter.Accept(ctx, f)
		if err != nil {
			return false, err
		}
		if accept {
			return true, nil
		}
	}

	return false, nil
}
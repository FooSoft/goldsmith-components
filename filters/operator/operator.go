package operator

import (
	"github.com/FooSoft/goldsmith"
)

type Operator interface {
	goldsmith.Filter
}

func And(filters ...goldsmith.Filter) Operator {
	return &operatorAnd{filters}
}

type operatorAnd struct {
	filters []goldsmith.Filter
}

func (*operatorAnd) Name() string {
	return "operator"
}

func (operator *operatorAnd) Accept(context *goldsmith.Context, file *goldsmith.File) (bool, error) {
	for _, filter := range operator.filters {
		accept, err := filter.Accept(context, file)
		if err != nil {
			return false, err
		}
		if !accept {
			return false, nil
		}
	}

	return true, nil
}

func Not(filter goldsmith.Filter) Operator {
	return &operatorNot{filter}
}

type operatorNot struct {
	filter goldsmith.Filter
}

func (*operatorNot) Name() string {
	return "operator"
}

func (operator *operatorNot) Accept(context *goldsmith.Context, file *goldsmith.File) (bool, error) {
	accept, err := operator.filter.Accept(context, file)
	return !accept, err
}

func Or(filters ...goldsmith.Filter) Operator {
	return &operatorOr{filters}
}

type operatorOr struct {
	filters []goldsmith.Filter
}

func (*operatorOr) Name() string {
	return "operator"
}

func (operator *operatorOr) Accept(context *goldsmith.Context, file *goldsmith.File) (bool, error) {
	for _, filter := range operator.filters {
		accept, err := filter.Accept(context, file)
		if err != nil {
			return false, err
		}
		if accept {
			return true, nil
		}
	}

	return false, nil
}

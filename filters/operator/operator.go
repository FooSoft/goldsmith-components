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

func (filter *operatorAnd) Accept(file *goldsmith.File) bool {
	for _, filter := range filter.filters {
		if !filter.Accept(file) {
			return false
		}
	}

	return true
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

func (filter *operatorNot) Accept(file *goldsmith.File) bool {
	return !filter.filter.Accept(file)
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

func (filter *operatorOr) Accept(file *goldsmith.File) bool {
	for _, filter := range filter.filters {
		if filter.Accept(file) {
			return true
		}
	}

	return false
}

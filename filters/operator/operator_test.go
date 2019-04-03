package operator

import (
	"testing"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/condition"
	"github.com/FooSoft/goldsmith-components/harness"
)

func TestAndFalse(t *testing.T) {
	harness.ValidateCase(
		t,
		"and_f",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(And(condition.New(false)))
		},
	)
}

func TestAndFalseTrue(t *testing.T) {
	harness.ValidateCase(
		t,
		"and_ft",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(And(condition.New(false), condition.New(true)))
		},
	)
}

func TestAndTrueFalse(t *testing.T) {
	harness.ValidateCase(
		t,
		"and_tf",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(And(condition.New(true), condition.New(false)))
		},
	)
}

func TestAndTrue(t *testing.T) {
	harness.ValidateCase(
		t,
		"and_t",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(And(condition.New(true)))
		},
	)
}

func TestOrFalse(t *testing.T) {
	harness.ValidateCase(
		t,
		"or_f",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(Or(condition.New(false)))
		},
	)
}

func TestOrFalseTrue(t *testing.T) {
	harness.ValidateCase(
		t,
		"or_ft",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(Or(condition.New(false), condition.New(true)))
		},
	)
}

func TestOrTrueFalse(t *testing.T) {
	harness.ValidateCase(
		t,
		"or_tf",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(Or(condition.New(true), condition.New(false)))
		},
	)
}

func TestOrTrue(t *testing.T) {
	harness.ValidateCase(
		t,
		"or_t",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(Or(condition.New(true)))
		},
	)
}

func TestNotFalse(t *testing.T) {
	harness.ValidateCase(
		t,
		"not_f",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(Not(condition.New(false)))
		},
	)
}

func TestNotTrue(t *testing.T) {
	harness.ValidateCase(
		t,
		"not_t",
		func(gs *goldsmith.Goldsmith) {
			gs.FilterPush(Not(condition.New(true)))
		},
	)
}

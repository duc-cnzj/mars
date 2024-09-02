package filters

import (
	"entgo.io/ent/dialect/sql"
)

var notMatchedFn = func(*sql.Selector) {}

type Operation string

type Condition[T any] func(T) bool

func If[T any](condition Condition[T], fn func(t T) func(*sql.Selector)) func(T) func(*sql.Selector) {
	return func(t T) func(*sql.Selector) {
		if condition(t) {
			return fn(t)
		}
		return notMatchedFn
	}
}

func IfStrEQ(field string) func(string) func(*sql.Selector) {
	return If[string](func(s string) bool {
		return s != ""
	}, func(s string) func(*sql.Selector) {
		return sql.FieldEQ(field, s)
	})
}

func IfIntEQ[T ~int | ~int32](field string) func(T) func(*sql.Selector) {
	return If[T](func(s T) bool {
		return s != 0
	}, func(s T) func(*sql.Selector) {
		return sql.FieldEQ(field, s)
	})
}

func IfOrderByDesc(field string) func(*bool) func(*sql.Selector) {
	return If[*bool](func(s *bool) bool {
		return s != nil && *s
	}, func(s *bool) func(*sql.Selector) {
		return sql.OrderByField(field, sql.OrderDesc()).ToFunc()
	})
}

func IfBool(field string) func(*bool) func(*sql.Selector) {
	return If[*bool](func(s *bool) bool {
		return s != nil
	}, func(s *bool) func(*sql.Selector) {
		return sql.FieldEQ(field, *s)
	})
}

var (
	IfEmail         = IfStrEQ("email")
	IfEnabled       = IfBool("enabled")
	IfOrderByIDDesc = IfOrderByDesc("id")
	IfNameLike      = If(func(s string) bool {
		return s != ""
	}, func(t string) func(*sql.Selector) {
		return sql.FieldContains("name", t)
	})
)

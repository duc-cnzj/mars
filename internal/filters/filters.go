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

func IfStrPtrEQ(field string) func(*string) func(*sql.Selector) {
	return If[*string](func(s *string) bool {
		return s != nil
	}, func(s *string) func(*sql.Selector) {
		return sql.FieldEQ(field, *s)
	})
}

func IfInt64PtrEQ(field string) func(*int64) func(*sql.Selector) {
	return If[*int64](func(s *int64) bool {
		return s != nil
	}, func(s *int64) func(*sql.Selector) {
		return sql.FieldEQ(field, s)
	})
}
func IfIntPtrEQ[T ~int | ~int32](field string) func(*T) func(*sql.Selector) {
	return If[*T](func(s *T) bool {
		return s != nil
	}, func(s *T) func(*sql.Selector) {
		return sql.FieldEQ(field, s)
	})
}

func IfInt64EQ(field string) func(int64) func(*sql.Selector) {
	return If[int64](func(s int64) bool {
		return s != 0
	}, func(s int64) func(*sql.Selector) {
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
		return s != nil
	}, func(s *bool) func(*sql.Selector) {
		return sql.OrderByField(field, sql.OrderDesc()).ToFunc()
	})
}
func IfOrderByAsc(field string) func(*bool) func(*sql.Selector) {
	return If[*bool](func(s *bool) bool {
		return s != nil
	}, func(s *bool) func(*sql.Selector) {
		return sql.OrderByField(field, sql.OrderAsc()).ToFunc()
	})
}

func IfBool(field string) func(*bool) func(*sql.Selector) {
	return If[*bool](func(s *bool) bool {
		return s != nil
	}, func(s *bool) func(*sql.Selector) {
		return sql.FieldEQ(field, *s)
	})
}

var IfEmail = IfStrEQ("email")
var IfEnabled = IfBool("enabled")
var IfOrderByIDDesc = IfOrderByDesc("id")
var IfNameLike = If(func(s string) bool {
	return s != ""
}, func(t string) func(*sql.Selector) {
	return sql.FieldContains("name", t)
})

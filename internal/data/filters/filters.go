package filters

import (
	"entgo.io/ent/dialect/sql"
)

// notMatchedFn 是条件不满足时的空查询过滤器（恒不追加任何 SQL 条件）。
var notMatchedFn = func(*sql.Selector) {}

// Condition 是过滤条件的判定函数：给定输入值返回是否放行。
type Condition[T any] func(T) bool

// If 依据 condition 决定是否应用 fn 生成的 SQL 过滤器，不满足时返回空过滤器。
func If[T any](condition Condition[T], fn func(t T) func(*sql.Selector)) func(T) func(*sql.Selector) {
	return func(t T) func(*sql.Selector) {
		if condition(t) {
			return fn(t)
		}
		return notMatchedFn
	}
}

// IfStrEQ 字符串相等过滤：输入非空时才追加 field = s。
func IfStrEQ(field string) func(string) func(*sql.Selector) {
	return If[string](func(s string) bool {
		return s != ""
	}, func(s string) func(*sql.Selector) {
		return sql.FieldEQ(field, s)
	})
}

// IfIntEQ 整数相等过滤：输入非零时才追加 field = s。
func IfIntEQ[T ~int | ~int32](field string) func(T) func(*sql.Selector) {
	return If[T](func(s T) bool {
		return s != 0
	}, func(s T) func(*sql.Selector) {
		return sql.FieldEQ(field, s)
	})
}

// IfOrderByDesc 倒序过滤：输入指针非 nil 且为 true 时追加 field DESC。
func IfOrderByDesc(field string) func(*bool) func(*sql.Selector) {
	return If[*bool](func(s *bool) bool {
		return s != nil && *s
	}, func(s *bool) func(*sql.Selector) {
		return sql.OrderByField(field, sql.OrderDesc()).ToFunc()
	})
}

// IfBool 布尔相等过滤：输入指针非 nil 时才追加 field = *s。
func IfBool(field string) func(*bool) func(*sql.Selector) {
	return If[*bool](func(s *bool) bool {
		return s != nil
	}, func(s *bool) func(*sql.Selector) {
		return sql.FieldEQ(field, *s)
	})
}

// IfEmail/IfEnabled/IfOrderByIDDesc/IfNameLike 是常用查询过滤器的命名实例。
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

// Code generated by ent, DO NOT EDIT.

package dbcache

import (
	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the dbcache type in the database.
	Label = "db_cache"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldKey holds the string denoting the key field in the database.
	FieldKey = "key"
	// FieldValue holds the string denoting the value field in the database.
	FieldValue = "value"
	// FieldExpiredAt holds the string denoting the expired_at field in the database.
	FieldExpiredAt = "expired_at"
	// Table holds the table name of the dbcache in the database.
	Table = "db_cache"
)

// Columns holds all SQL columns for dbcache fields.
var Columns = []string{
	FieldID,
	FieldKey,
	FieldValue,
	FieldExpiredAt,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// KeyValidator is a validator for the "key" field. It is called by the builders before save.
	KeyValidator func(string) error
)

// OrderOption defines the ordering options for the DBCache queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByKey orders the results by the key field.
func ByKey(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldKey, opts...).ToFunc()
}

// ByValue orders the results by the value field.
func ByValue(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldValue, opts...).ToFunc()
}

// ByExpiredAt orders the results by the expired_at field.
func ByExpiredAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldExpiredAt, opts...).ToFunc()
}

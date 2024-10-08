// Code generated by ent, DO NOT EDIT.

package event

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/duc-cnzj/mars/api/v5/types"
)

const (
	// Label holds the string label denoting the event type in the database.
	Label = "event"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldDeletedAt holds the string denoting the deleted_at field in the database.
	FieldDeletedAt = "deleted_at"
	// FieldAction holds the string denoting the action field in the database.
	FieldAction = "action"
	// FieldUsername holds the string denoting the username field in the database.
	FieldUsername = "username"
	// FieldMessage holds the string denoting the message field in the database.
	FieldMessage = "message"
	// FieldOld holds the string denoting the old field in the database.
	FieldOld = "old"
	// FieldNew holds the string denoting the new field in the database.
	FieldNew = "new"
	// FieldHasDiff holds the string denoting the has_diff field in the database.
	FieldHasDiff = "has_diff"
	// FieldDuration holds the string denoting the duration field in the database.
	FieldDuration = "duration"
	// FieldFileID holds the string denoting the file_id field in the database.
	FieldFileID = "file_id"
	// EdgeFile holds the string denoting the file edge name in mutations.
	EdgeFile = "file"
	// Table holds the table name of the event in the database.
	Table = "events"
	// FileTable is the table that holds the file relation/edge.
	FileTable = "events"
	// FileInverseTable is the table name for the File entity.
	// It exists in this package in order to avoid circular dependency with the "file" package.
	FileInverseTable = "files"
	// FileColumn is the table column denoting the file relation/edge.
	FileColumn = "file_id"
)

// Columns holds all SQL columns for event fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldDeletedAt,
	FieldAction,
	FieldUsername,
	FieldMessage,
	FieldOld,
	FieldNew,
	FieldHasDiff,
	FieldDuration,
	FieldFileID,
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

// Note that the variables below are initialized by the runtime
// package on the initialization of the application. Therefore,
// it should be imported in the main as follows:
//
//	import _ "github.com/duc-cnzj/mars/v5/internal/ent/runtime"
var (
	Hooks        [1]ent.Hook
	Interceptors [1]ent.Interceptor
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultAction holds the default value on creation for the "action" field.
	DefaultAction types.EventActionType
	// DefaultUsername holds the default value on creation for the "username" field.
	DefaultUsername string
	// UsernameValidator is a validator for the "username" field. It is called by the builders before save.
	UsernameValidator func(string) error
	// DefaultMessage holds the default value on creation for the "message" field.
	DefaultMessage string
	// MessageValidator is a validator for the "message" field. It is called by the builders before save.
	MessageValidator func(string) error
	// DefaultHasDiff holds the default value on creation for the "has_diff" field.
	DefaultHasDiff bool
	// DefaultDuration holds the default value on creation for the "duration" field.
	DefaultDuration string
)

// OrderOption defines the ordering options for the Event queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByDeletedAt orders the results by the deleted_at field.
func ByDeletedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDeletedAt, opts...).ToFunc()
}

// ByAction orders the results by the action field.
func ByAction(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAction, opts...).ToFunc()
}

// ByUsername orders the results by the username field.
func ByUsername(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUsername, opts...).ToFunc()
}

// ByMessage orders the results by the message field.
func ByMessage(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldMessage, opts...).ToFunc()
}

// ByOld orders the results by the old field.
func ByOld(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOld, opts...).ToFunc()
}

// ByNew orders the results by the new field.
func ByNew(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldNew, opts...).ToFunc()
}

// ByHasDiff orders the results by the has_diff field.
func ByHasDiff(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHasDiff, opts...).ToFunc()
}

// ByDuration orders the results by the duration field.
func ByDuration(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDuration, opts...).ToFunc()
}

// ByFileID orders the results by the file_id field.
func ByFileID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldFileID, opts...).ToFunc()
}

// ByFileField orders the results by file field.
func ByFileField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newFileStep(), sql.OrderByField(field, opts...))
	}
}
func newFileStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(FileInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, FileTable, FileColumn),
	)
}

// Code generated by ent, DO NOT EDIT.

package member

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the member type in the database.
	Label = "member"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldDeletedAt holds the string denoting the deleted_at field in the database.
	FieldDeletedAt = "deleted_at"
	// FieldEmail holds the string denoting the email field in the database.
	FieldEmail = "email"
	// FieldNamespaceID holds the string denoting the namespace_id field in the database.
	FieldNamespaceID = "namespace_id"
	// EdgeNamespace holds the string denoting the namespace edge name in mutations.
	EdgeNamespace = "namespace"
	// Table holds the table name of the member in the database.
	Table = "members"
	// NamespaceTable is the table that holds the namespace relation/edge.
	NamespaceTable = "members"
	// NamespaceInverseTable is the table name for the Namespace entity.
	// It exists in this package in order to avoid circular dependency with the "namespace" package.
	NamespaceInverseTable = "namespaces"
	// NamespaceColumn is the table column denoting the namespace relation/edge.
	NamespaceColumn = "namespace_id"
)

// Columns holds all SQL columns for member fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldDeletedAt,
	FieldEmail,
	FieldNamespaceID,
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
	// EmailValidator is a validator for the "email" field. It is called by the builders before save.
	EmailValidator func(string) error
)

// OrderOption defines the ordering options for the Member queries.
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

// ByEmail orders the results by the email field.
func ByEmail(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldEmail, opts...).ToFunc()
}

// ByNamespaceID orders the results by the namespace_id field.
func ByNamespaceID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldNamespaceID, opts...).ToFunc()
}

// ByNamespaceField orders the results by namespace field.
func ByNamespaceField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newNamespaceStep(), sql.OrderByField(field, opts...))
	}
}
func newNamespaceStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(NamespaceInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, NamespaceTable, NamespaceColumn),
	)
}

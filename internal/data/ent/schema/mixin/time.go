package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
)

// Time composes create/update time mixin.
type Time struct{ ent.Schema }

// Fields of the time mixin.
func (Time) Fields() []ent.Field {
	return append(
		CreateAt{}.Fields(),
		UpdateAt{}.Fields()...,
	)
}

// CreateAt adds created at time field.
type CreateAt struct{ ent.Schema }

// Fields of the create time mixin.
func (CreateAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}).
			Immutable(),
	}
}

// UpdateAt adds updated at time field.
type UpdateAt struct{ ent.Schema }

// Fields of the update time mixin.
func (UpdateAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("updated_at").
			Default(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}).
			UpdateDefault(time.Now),
	}
}

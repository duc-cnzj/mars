package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Favorite holds the schema definition for the Favorite entity.
type Favorite struct {
	ent.Schema
}

// Fields of the Favorite.
func (Favorite) Fields() []ent.Field {
	return []ent.Field{
		field.String("email"),
		field.Int("namespace_id").
			Optional(),
	}
}

// Edges of the Favorite.
func (Favorite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).
			Ref("favorites").
			Unique().
			Field("namespace_id"),
	}
}

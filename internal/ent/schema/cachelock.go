package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
)

// CacheLock holds the schema definition for the CacheLock entity.
type CacheLock struct {
	ent.Schema
}

// Fields of the CacheLock.
func (CacheLock) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty().
			Unique(),
		field.String("owner").
			NotEmpty(),
		field.Time("expired_at").
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}).
			Optional(),
	}
}

// Edges of the CacheLock.
func (CacheLock) Edges() []ent.Edge {
	return nil
}

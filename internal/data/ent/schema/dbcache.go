package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// DBCache holds the schema definition for the DBCache entity.
type DBCache struct {
	ent.Schema
}

// Fields of the DBCache.
func (DBCache) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty().
			Unique(),
		field.String("value").
			SchemaType(map[string]string{
				dialect.MySQL: "longtext",
			}).
			Optional(),
		field.Time("expired_at").
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}).
			Optional(),
	}
}

// Edges of the DBCache.
func (DBCache) Edges() []ent.Edge {
	return nil
}

func (DBCache) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "db_cache"},
	}
}

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/mixin"
)

// Event holds the schema definition for the Event entity.
type Event struct {
	ent.Schema
}

// Fields of the Event.
func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.Int32("action").GoType(types.EventActionType(0)).Default(0),
		field.String("username").MaxLen(255).Default("").Comment("用户名称"),
		field.String("message").MaxLen(255).Default(""),
		field.String("old").
			SchemaType(map[string]string{
				dialect.MySQL: "longtext",
			}),
		field.String("new").
			SchemaType(map[string]string{
				dialect.MySQL: "longtext",
			}),
		field.String("duration").Default(""),
		field.Int("file_id").Optional().Nillable(),
	}
}

// Edges of the Event.
func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("file", File.Type).
			Ref("events").
			Unique().
			Field("file_id"),
	}
}

func (Event) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("action"),
		index.Fields("username", "created_at"),
	}
}
func (Event) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
		mixin.SoftDeleteMixin{},
	}
}

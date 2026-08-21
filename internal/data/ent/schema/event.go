package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
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
		field.String("operator_email").MaxLen(255).Default("").Comment("操作人邮箱，用于普通用户我的事件归属过滤"),
		field.String("message").MaxLen(255).Default(""),
		field.String("old").
			SchemaType(map[string]string{
				dialect.MySQL: "longtext",
			}).
			Optional(),
		field.String("new").
			SchemaType(map[string]string{
				dialect.MySQL: "longtext",
			}).
			Optional(),
		field.Bool("has_diff").Default(false),
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
		index.Fields("operator_email", "created_at"),
	}
}
func (Event) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
		mixin.SoftDeleteMixin{},
	}
}

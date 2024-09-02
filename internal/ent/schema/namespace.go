package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/mixin"
)

// Namespace holds the schema definition for the Namespace entity.
type Namespace struct {
	ent.Schema
}

// Fields of the Namespace.
func (Namespace) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(100).
			Annotations(
				entsql.Annotation{
					Charset:   "utf8mb4",
					Collation: "utf8mb4_general_ci",
				},
			).
			Comment("项目空间名"),
		field.Strings("image_pull_secrets").
			Default([]string{}).
			Comment("image pull secrets"),
		field.Bool("private").
			Default(false).
			Comment("是否私有, 默认公开"),
		field.String("creator_email").
			MaxLen(50).
			Comment("创建者 email"),
		field.String("description").
			SchemaType(map[string]string{
				dialect.MySQL: "text",
			}).
			Optional().
			Comment("项目空间描述"),
	}
}

// Edges of the Namespace.
func (Namespace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("projects", Project.Type),
		edge.To("favorites", Favorite.Type),
		edge.To("members", Member.Type),
	}
}

func (Namespace) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
		mixin.SoftDeleteMixin{},
	}
}

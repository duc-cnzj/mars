package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
)

// AccessToken holds the schema definition for the AccessToken entity.
type AccessToken struct {
	ent.Schema
}

// Fields of the AccessToken.
func (AccessToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").
			MaxLen(100).
			Unique().
			NotEmpty(),
		field.String("usage").
			MaxLen(50),
		field.String("email").
			Default("").
			MaxLen(255),
		field.Time("expired_at").
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}).
			Optional(),
		field.Time("last_used_at").
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}).
			Nillable().
			Optional(),
		field.JSON("user_info", schematype.UserInfo{}).
			Optional(),
	}
}

// Edges of the AccessToken.
func (AccessToken) Edges() []ent.Edge {
	return nil
}

func (AccessToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
	}
}

func (AccessToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
		mixin.SoftDeleteMixin{},
	}
}

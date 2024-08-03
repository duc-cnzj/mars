package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/mixin"
)

// GitProject holds the schema definition for the GitProject entity.
type GitProject struct {
	ent.Schema
}

// Fields of the GitProject.
func (GitProject) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(255).Default(""),
		field.String("default_branch").MaxLen(255).Default(""),
		field.Int("git_project_id").Unique(),
		field.Bool("enabled").Default(false),
		field.Bool("global_enabled").Default(false),
		field.JSON("global_config", &mars.Config{}).
			Optional().
			Comment("全局配置"),
	}
}

// Edges of the GitProject.
func (GitProject) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("changelogs", Changelog.Type),
	}
}

func (GitProject) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
		mixin.SoftDeleteMixin{},
	}
}

package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/mixin"
)

// Repo holds the schema definition for the Repo entity.
type Repo struct {
	ent.Schema
}

// Fields of the Repo.
func (Repo) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(255).
			Match(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)).
			Comment("默认使用的名称: helm create {name}"),
		field.String("default_branch").
			MaxLen(255).
			Nillable().
			Optional(),
		field.String("git_project_name").
			Optional().
			Nillable().
			Comment("关联的 git 项目 name"),
		field.Int32("git_project_id").
			Optional().
			Nillable().
			Comment("关联的 git 项目 id"),
		field.Bool("enabled").
			Default(false),
		field.JSON("mars_config", &mars.Config{}).
			Optional().
			Comment("mars 配置"),
	}
}

// Edges of the Repo.
func (Repo) Edges() []ent.Edge {
	return nil
}

func (Repo) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
		mixin.SoftDeleteMixin{},
	}
}

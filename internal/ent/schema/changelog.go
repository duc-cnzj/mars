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

// Changelog holds the schema definition for the Changelog entity.
type Changelog struct {
	ent.Schema
}

// Fields of the Changelog.
func (Changelog) Fields() []ent.Field {
	return []ent.Field{
		field.Int("version").Default(1),
		field.String("username").MaxLen(100).Comment("修改人"),
		field.Strings("manifest").Optional(),
		field.String("config").
			Optional().
			Comment("用户提交的配置"),
		field.String("config_type").MaxLen(255).Optional(),
		field.String("git_branch").MaxLen(255).NotEmpty(),
		field.String("git_commit").MaxLen(255).NotEmpty(),
		field.Strings("docker_image").Optional(),
		field.JSON("env_values", []*types.KeyValue{}).Optional().Comment("可用的环境变量值"),
		field.JSON("extra_values", []*types.ExtraValue{}).Optional().Comment("用户表单传入的额外值"),
		field.Strings("final_extra_values").Optional().Comment("用户表单传入的额外值 + 系统默认的额外值"),
		field.String("git_commit_web_url").MaxLen(255).Optional(),
		field.String("git_commit_title").MaxLen(255).Optional(),
		field.String("git_commit_author").MaxLen(255).Optional(),
		field.Time("git_commit_date").
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}).
			Nillable().
			Optional(),
		field.Bool("config_changed").Default(false),
		field.Int("project_id").Optional(),
		field.Int("git_project_id").Optional(),
	}
}

// Edges of the Changelog.
func (Changelog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("git_project", GitProject.Type).
			Ref("changelogs").
			Unique().
			Field("git_project_id"),
		edge.From("project", Project.Type).
			Ref("changelogs").
			Unique().
			Field("project_id"),
	}
}

// Indexes of the Changelog.
func (Changelog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id", "config_changed", "deleted_at", "version"),
	}
}

// Mixin of the Changelog.
func (Changelog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
		mixin.SoftDeleteMixin{},
	}
}

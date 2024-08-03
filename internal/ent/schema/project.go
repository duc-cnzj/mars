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

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(100).
			Default("").
			Comment("项目名"),
		field.Int("git_project_id"),
		field.String("git_branch").
			MaxLen(255).
			Comment("git 分支"),
		field.String("git_commit").
			MaxLen(255).
			Comment("git commit"),
		field.String("config"),
		field.String("override_values"),
		field.Strings("docker_image").
			Comment("docker 镜像"),
		field.Strings("pod_selectors").
			Comment("pod 选择器"),
		field.Bool("atomic").
			Default(false),
		field.Int32("deploy_status").
			GoType(types.Deploy(0)).
			Default(0).
			Comment("部署状态"),
		field.JSON("env_values", []*types.KeyValue{}).
			Comment("环境变量值"),
		field.JSON("extra_values", []*types.ExtraValue{}).
			Comment("额外值"),
		field.Strings("final_extra_values").
			Comment("用户表单传入的额外值 + 系统默认的额外值"),
		field.Int("version").
			Default(1).
			Comment("版本"),
		field.String("config_type").
			MaxLen(255).
			Optional(),
		field.Strings("manifest").
			Comment("manifest"),
		field.String("git_commit_web_url").
			MaxLen(255).
			Default(""),
		field.String("git_commit_title").
			MaxLen(255).
			Default(""),
		field.String("git_commit_author").
			MaxLen(255).
			Default(""),
		field.Time("git_commit_date").
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}).
			Nillable().
			Optional(),
		field.Int("namespace_id").
			Optional(),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("changelogs", Changelog.Type),
		edge.From("namespace", Namespace.Type).
			Ref("projects").
			Unique().
			Field("namespace_id"),
	}
}

func (Project) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
		mixin.SoftDeleteMixin{},
	}
}
func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("git_project_id"),
	}
}

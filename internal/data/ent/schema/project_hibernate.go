package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
)

// ProjectHibernate 是项目定时休眠（夜间缩容）的运行态（1:1 项目）：
// 项目级开关 + 运行态（awake/sleeping）+ 休眠前副本字段的值快照。
// 副本字段映射与时间窗是仓库级配置（repos.mars_config.hibernate），不落本表。
type ProjectHibernate struct {
	ent.Schema
}

// Fields of the ProjectHibernate.
func (ProjectHibernate) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("enabled").
			Default(false).
			Comment("项目级休眠开关（所属仓库启用 hibernate 能力后才有意义）"),
		field.String("state").
			MaxLen(32).
			Default("awake").
			Comment("运行态：awake 清醒 / sleeping 已休眠"),
		field.JSON("snapshot", []*websocket_pb.ExtraValue{}).
			Optional().
			Comment("休眠前副本字段的值快照，唤醒时据此恢复"),
		field.Strings("window_names").
			Optional().
			Comment("项目选定的时间窗名子集（引用仓库 Hibernate.window_names）；启用时非空，空 = 该项目休眠已关闭"),
		field.Int("project_id").
			Optional(),
	}
}

// Edges of the ProjectHibernate.
func (ProjectHibernate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("hibernate").
			Unique().
			Field("project_id"),
	}
}

// Indexes of the ProjectHibernate.
func (ProjectHibernate) Indexes() []ent.Index {
	return []ent.Index{
		// 每项目一条：DB 级幂等兜底（Upsert 的 ON CONFLICT 依赖它）。
		index.Fields("project_id").Unique(),
	}
}

// Mixin of the ProjectHibernate.
// 不挂 SoftDeleteMixin：本表是运行态附属行、无审计保留价值，且唯一索引落在 project_id 上——
// 软删行会占着唯一键，使项目恢复后再次 Upsert 只更新到一条「看不见的」软删行，状态永久丢失。
func (ProjectHibernate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
	}
}

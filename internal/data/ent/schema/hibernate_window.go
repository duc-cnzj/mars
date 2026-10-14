package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
)

// HibernateWindow 是全局共享的休眠时间窗：管理员在后台维护若干命名时段，
// 各仓库的定时休眠配置按名字引用其中若干（repos.mars_config.hibernate.window_names）。
// 时间窗全局共享、副本字段仓库各自配置——本表只存时段本身。
type HibernateWindow struct {
	ent.Schema
}

// Fields of the HibernateWindow.
func (HibernateWindow) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(64).
			Comment("时间窗名称（全局唯一，仓库按名字引用）"),
		field.String("start").
			MaxLen(5).
			Comment("休眠开始时间（服务端本地时区，HH:MM）"),
		field.String("end").
			MaxLen(5).
			Comment("休眠结束时间（HH:MM，start>end 表示跨零点）"),
		field.String("description").
			Optional().
			Comment("备注说明"),
	}
}

// Indexes of the HibernateWindow.
func (HibernateWindow) Indexes() []ent.Index {
	return []ent.Index{
		// 仓库按名字引用时间窗，名字必须全局唯一，否则解析歧义。
		index.Fields("name").Unique(),
	}
}

// Mixin of the HibernateWindow.
// 不挂 SoftDeleteMixin：时间窗是管理员维护的全局字典行，无审计保留价值；且仓库按名字引用，
// 软删行会让名字唯一键被「看不见的行」长期占位。删除即物理删除，引用它的仓库在对账时跳过。
func (HibernateWindow) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
	}
}

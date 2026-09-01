package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
)

// User holds the schema definition for the User entity.
// 管理后台「用户管理」的本地投影：从内置管理员/命名空间成员等真实身份同步而来，
// 登录用户由 SyncLoginUser 在登录热路径实时 upsert，email 为唯一身份标识。
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			Unique().
			MaxLen(255).
			Comment("用户邮箱（唯一身份标识）"),
		field.String("name").
			MaxLen(255).
			Default("").
			Comment("展示名"),
		field.JSON("roles", []string{}).
			Default([]string{}).
			Comment("角色列表：mars_admin=管理员；空数组=普通用户"),
		field.Bool("roles_override").
			Default(false).
			Comment("角色是否已被后台手动管理接管：false=登录时按 SSO 角色同步；true=后台手动升降级后 SSO 不再覆盖"),
		field.Time("last_login").
			Nillable().
			Optional().
			Comment("最近登录时间（取最近一条登录事件）"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("last_login"),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
	}
}

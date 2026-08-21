package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Favorite holds the schema definition for the Favorite entity.
type Favorite struct {
	ent.Schema
}

// Fields of the Favorite.
func (Favorite) Fields() []ent.Field {
	return []ent.Field{
		field.String("email"),
		field.Int("namespace_id").
			Optional(),
		// 用户维度排序值：越小越靠前；新关注写 MAX+1 追加末尾，拖拽重排按序回填 0,1,2…
		// 排序只存在 favorites 行上（每用户每空间一条），天然 per-user，不串用户。
		field.Int("sort_order").
			Default(0).
			Comment("用户关注列表排序值，越小越靠前"),
	}
}

// Edges of the Favorite.
func (Favorite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).
			Ref("favorites").
			Unique().
			Field("namespace_id"),
	}
}

func (Favorite) Indexes() []ent.Index {
	return []ent.Index{
		// 每用户每空间唯一：DB 级幂等（应用层 exist 检查之外的兜底）
		index.Fields("email", "namespace_id").Unique(),
		// 关注列表按用户 + 排序值检索
		index.Fields("email", "sort_order"),
	}
}

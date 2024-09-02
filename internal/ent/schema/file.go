package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
)

// File holds the schema definition for the File entity.
type File struct {
	ent.Schema
}

// Fields of the File.
func (File) Fields() []ent.Field {
	return []ent.Field{
		field.String("upload_type").
			MaxLen(100).
			GoType(schematype.UploadType("")).
			Default(string(schematype.Local)),
		field.String("path").
			MaxLen(255).
			Comment("文件全路径"),
		field.Uint64("size").
			SchemaType(map[string]string{
				dialect.MySQL: "int",
			}).
			Default(0).
			Comment("文件大小"),
		field.String("username").
			Default("").
			MaxLen(255).
			Comment("用户名称"),
		field.String("namespace").
			MaxLen(100).
			Default(""),
		field.String("pod").
			MaxLen(100).
			Default(""),
		field.String("container").
			MaxLen(100).
			Default(""),
		field.String("container_path").
			MaxLen(255).
			Default("").
			Comment("容器中的文件路径"),
	}
}

// Edges of the File.
func (File) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("events", Event.Type),
	}
}

func (File) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateAt{},
		mixin.UpdateAt{},
		mixin.SoftDeleteMixin{},
	}
}

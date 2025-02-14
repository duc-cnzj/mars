package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
)

type Email struct {
	ent.Schema

	name    string
	comment string
}

type EmailOption func(em *Email)

func WithEmailFieldName(name string) EmailOption {
	return func(em *Email) {
		em.name = name
	}
}
func WithEmailFieldComment(comment string) EmailOption {
	return func(em *Email) {
		em.comment = comment
	}
}

func NewEmail(opt ...EmailOption) ent.Mixin {
	r := &Email{
		name: "email",
	}

	for _, v := range opt {
		v(r)
	}

	return r
}

// Fields of the update time mixin.
func (e *Email) Fields() []ent.Field {
	return []ent.Field{
		field.String(e.name).
			MaxLen(50).
			Comment(e.comment).
			Annotations(
				entsql.Annotation{
					Charset:   "utf8mb4",
					Collation: "utf8mb4_general_ci",
				},
			),
	}
}

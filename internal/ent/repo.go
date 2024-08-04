// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/ent/repo"
)

// Repo is the model entity for the Repo schema.
type Repo struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// DeletedAt holds the value of the "deleted_at" field.
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	// 默认使用的名称: helm create {name}
	Name string `json:"name,omitempty"`
	// DefaultBranch holds the value of the "default_branch" field.
	DefaultBranch *string `json:"default_branch,omitempty"`
	// 关联的 git 项目 name
	GitProjectName *string `json:"git_project_name,omitempty"`
	// 关联的 git 项目 id
	GitProjectID *int32 `json:"git_project_id,omitempty"`
	// Enabled holds the value of the "enabled" field.
	Enabled bool `json:"enabled,omitempty"`
	// mars 配置
	MarsConfig   *mars.Config `json:"mars_config,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Repo) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case repo.FieldMarsConfig:
			values[i] = new([]byte)
		case repo.FieldEnabled:
			values[i] = new(sql.NullBool)
		case repo.FieldID, repo.FieldGitProjectID:
			values[i] = new(sql.NullInt64)
		case repo.FieldName, repo.FieldDefaultBranch, repo.FieldGitProjectName:
			values[i] = new(sql.NullString)
		case repo.FieldCreatedAt, repo.FieldUpdatedAt, repo.FieldDeletedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Repo fields.
func (r *Repo) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case repo.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			r.ID = int(value.Int64)
		case repo.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				r.CreatedAt = value.Time
			}
		case repo.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				r.UpdatedAt = value.Time
			}
		case repo.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				r.DeletedAt = new(time.Time)
				*r.DeletedAt = value.Time
			}
		case repo.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				r.Name = value.String
			}
		case repo.FieldDefaultBranch:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field default_branch", values[i])
			} else if value.Valid {
				r.DefaultBranch = new(string)
				*r.DefaultBranch = value.String
			}
		case repo.FieldGitProjectName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field git_project_name", values[i])
			} else if value.Valid {
				r.GitProjectName = new(string)
				*r.GitProjectName = value.String
			}
		case repo.FieldGitProjectID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field git_project_id", values[i])
			} else if value.Valid {
				r.GitProjectID = new(int32)
				*r.GitProjectID = int32(value.Int64)
			}
		case repo.FieldEnabled:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field enabled", values[i])
			} else if value.Valid {
				r.Enabled = value.Bool
			}
		case repo.FieldMarsConfig:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field mars_config", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &r.MarsConfig); err != nil {
					return fmt.Errorf("unmarshal field mars_config: %w", err)
				}
			}
		default:
			r.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Repo.
// This includes values selected through modifiers, order, etc.
func (r *Repo) Value(name string) (ent.Value, error) {
	return r.selectValues.Get(name)
}

// Update returns a builder for updating this Repo.
// Note that you need to call Repo.Unwrap() before calling this method if this Repo
// was returned from a transaction, and the transaction was committed or rolled back.
func (r *Repo) Update() *RepoUpdateOne {
	return NewRepoClient(r.config).UpdateOne(r)
}

// Unwrap unwraps the Repo entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (r *Repo) Unwrap() *Repo {
	_tx, ok := r.config.driver.(*txDriver)
	if !ok {
		panic("ent: Repo is not a transactional entity")
	}
	r.config.driver = _tx.drv
	return r
}

// String implements the fmt.Stringer.
func (r *Repo) String() string {
	var builder strings.Builder
	builder.WriteString("Repo(")
	builder.WriteString(fmt.Sprintf("id=%v, ", r.ID))
	builder.WriteString("created_at=")
	builder.WriteString(r.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(r.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	if v := r.DeletedAt; v != nil {
		builder.WriteString("deleted_at=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(r.Name)
	builder.WriteString(", ")
	if v := r.DefaultBranch; v != nil {
		builder.WriteString("default_branch=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	if v := r.GitProjectName; v != nil {
		builder.WriteString("git_project_name=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	if v := r.GitProjectID; v != nil {
		builder.WriteString("git_project_id=")
		builder.WriteString(fmt.Sprintf("%v", *v))
	}
	builder.WriteString(", ")
	builder.WriteString("enabled=")
	builder.WriteString(fmt.Sprintf("%v", r.Enabled))
	builder.WriteString(", ")
	builder.WriteString("mars_config=")
	builder.WriteString(fmt.Sprintf("%v", r.MarsConfig))
	builder.WriteByte(')')
	return builder.String()
}

// Repos is a parsable slice of Repo.
type Repos []*Repo

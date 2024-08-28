// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/duc-cnzj/mars/api/v5/mars"
	"github.com/duc-cnzj/mars/v5/internal/ent/project"
	"github.com/duc-cnzj/mars/v5/internal/ent/repo"
)

// RepoCreate is the builder for creating a Repo entity.
type RepoCreate struct {
	config
	mutation *RepoMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetCreatedAt sets the "created_at" field.
func (rc *RepoCreate) SetCreatedAt(t time.Time) *RepoCreate {
	rc.mutation.SetCreatedAt(t)
	return rc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (rc *RepoCreate) SetNillableCreatedAt(t *time.Time) *RepoCreate {
	if t != nil {
		rc.SetCreatedAt(*t)
	}
	return rc
}

// SetUpdatedAt sets the "updated_at" field.
func (rc *RepoCreate) SetUpdatedAt(t time.Time) *RepoCreate {
	rc.mutation.SetUpdatedAt(t)
	return rc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (rc *RepoCreate) SetNillableUpdatedAt(t *time.Time) *RepoCreate {
	if t != nil {
		rc.SetUpdatedAt(*t)
	}
	return rc
}

// SetDeletedAt sets the "deleted_at" field.
func (rc *RepoCreate) SetDeletedAt(t time.Time) *RepoCreate {
	rc.mutation.SetDeletedAt(t)
	return rc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (rc *RepoCreate) SetNillableDeletedAt(t *time.Time) *RepoCreate {
	if t != nil {
		rc.SetDeletedAt(*t)
	}
	return rc
}

// SetName sets the "name" field.
func (rc *RepoCreate) SetName(s string) *RepoCreate {
	rc.mutation.SetName(s)
	return rc
}

// SetDefaultBranch sets the "default_branch" field.
func (rc *RepoCreate) SetDefaultBranch(s string) *RepoCreate {
	rc.mutation.SetDefaultBranch(s)
	return rc
}

// SetNillableDefaultBranch sets the "default_branch" field if the given value is not nil.
func (rc *RepoCreate) SetNillableDefaultBranch(s *string) *RepoCreate {
	if s != nil {
		rc.SetDefaultBranch(*s)
	}
	return rc
}

// SetGitProjectName sets the "git_project_name" field.
func (rc *RepoCreate) SetGitProjectName(s string) *RepoCreate {
	rc.mutation.SetGitProjectName(s)
	return rc
}

// SetNillableGitProjectName sets the "git_project_name" field if the given value is not nil.
func (rc *RepoCreate) SetNillableGitProjectName(s *string) *RepoCreate {
	if s != nil {
		rc.SetGitProjectName(*s)
	}
	return rc
}

// SetGitProjectID sets the "git_project_id" field.
func (rc *RepoCreate) SetGitProjectID(i int32) *RepoCreate {
	rc.mutation.SetGitProjectID(i)
	return rc
}

// SetNillableGitProjectID sets the "git_project_id" field if the given value is not nil.
func (rc *RepoCreate) SetNillableGitProjectID(i *int32) *RepoCreate {
	if i != nil {
		rc.SetGitProjectID(*i)
	}
	return rc
}

// SetEnabled sets the "enabled" field.
func (rc *RepoCreate) SetEnabled(b bool) *RepoCreate {
	rc.mutation.SetEnabled(b)
	return rc
}

// SetNillableEnabled sets the "enabled" field if the given value is not nil.
func (rc *RepoCreate) SetNillableEnabled(b *bool) *RepoCreate {
	if b != nil {
		rc.SetEnabled(*b)
	}
	return rc
}

// SetNeedGitRepo sets the "need_git_repo" field.
func (rc *RepoCreate) SetNeedGitRepo(b bool) *RepoCreate {
	rc.mutation.SetNeedGitRepo(b)
	return rc
}

// SetNillableNeedGitRepo sets the "need_git_repo" field if the given value is not nil.
func (rc *RepoCreate) SetNillableNeedGitRepo(b *bool) *RepoCreate {
	if b != nil {
		rc.SetNeedGitRepo(*b)
	}
	return rc
}

// SetMarsConfig sets the "mars_config" field.
func (rc *RepoCreate) SetMarsConfig(m *mars.Config) *RepoCreate {
	rc.mutation.SetMarsConfig(m)
	return rc
}

// SetDescription sets the "description" field.
func (rc *RepoCreate) SetDescription(s string) *RepoCreate {
	rc.mutation.SetDescription(s)
	return rc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (rc *RepoCreate) SetNillableDescription(s *string) *RepoCreate {
	if s != nil {
		rc.SetDescription(*s)
	}
	return rc
}

// AddProjectIDs adds the "projects" edge to the Project entity by IDs.
func (rc *RepoCreate) AddProjectIDs(ids ...int) *RepoCreate {
	rc.mutation.AddProjectIDs(ids...)
	return rc
}

// AddProjects adds the "projects" edges to the Project entity.
func (rc *RepoCreate) AddProjects(p ...*Project) *RepoCreate {
	ids := make([]int, len(p))
	for i := range p {
		ids[i] = p[i].ID
	}
	return rc.AddProjectIDs(ids...)
}

// Mutation returns the RepoMutation object of the builder.
func (rc *RepoCreate) Mutation() *RepoMutation {
	return rc.mutation
}

// Save creates the Repo in the database.
func (rc *RepoCreate) Save(ctx context.Context) (*Repo, error) {
	if err := rc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, rc.sqlSave, rc.mutation, rc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (rc *RepoCreate) SaveX(ctx context.Context) *Repo {
	v, err := rc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (rc *RepoCreate) Exec(ctx context.Context) error {
	_, err := rc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rc *RepoCreate) ExecX(ctx context.Context) {
	if err := rc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (rc *RepoCreate) defaults() error {
	if _, ok := rc.mutation.CreatedAt(); !ok {
		if repo.DefaultCreatedAt == nil {
			return fmt.Errorf("ent: uninitialized repo.DefaultCreatedAt (forgotten import ent/runtime?)")
		}
		v := repo.DefaultCreatedAt()
		rc.mutation.SetCreatedAt(v)
	}
	if _, ok := rc.mutation.UpdatedAt(); !ok {
		if repo.DefaultUpdatedAt == nil {
			return fmt.Errorf("ent: uninitialized repo.DefaultUpdatedAt (forgotten import ent/runtime?)")
		}
		v := repo.DefaultUpdatedAt()
		rc.mutation.SetUpdatedAt(v)
	}
	if _, ok := rc.mutation.Enabled(); !ok {
		v := repo.DefaultEnabled
		rc.mutation.SetEnabled(v)
	}
	if _, ok := rc.mutation.NeedGitRepo(); !ok {
		v := repo.DefaultNeedGitRepo
		rc.mutation.SetNeedGitRepo(v)
	}
	if _, ok := rc.mutation.Description(); !ok {
		v := repo.DefaultDescription
		rc.mutation.SetDescription(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (rc *RepoCreate) check() error {
	if _, ok := rc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Repo.created_at"`)}
	}
	if _, ok := rc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "Repo.updated_at"`)}
	}
	if _, ok := rc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "Repo.name"`)}
	}
	if v, ok := rc.mutation.Name(); ok {
		if err := repo.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`ent: validator failed for field "Repo.name": %w`, err)}
		}
	}
	if v, ok := rc.mutation.DefaultBranch(); ok {
		if err := repo.DefaultBranchValidator(v); err != nil {
			return &ValidationError{Name: "default_branch", err: fmt.Errorf(`ent: validator failed for field "Repo.default_branch": %w`, err)}
		}
	}
	if _, ok := rc.mutation.Enabled(); !ok {
		return &ValidationError{Name: "enabled", err: errors.New(`ent: missing required field "Repo.enabled"`)}
	}
	if _, ok := rc.mutation.NeedGitRepo(); !ok {
		return &ValidationError{Name: "need_git_repo", err: errors.New(`ent: missing required field "Repo.need_git_repo"`)}
	}
	if v, ok := rc.mutation.MarsConfig(); ok {
		if err := v.Validate(); err != nil {
			return &ValidationError{Name: "mars_config", err: fmt.Errorf(`ent: validator failed for field "Repo.mars_config": %w`, err)}
		}
	}
	if _, ok := rc.mutation.Description(); !ok {
		return &ValidationError{Name: "description", err: errors.New(`ent: missing required field "Repo.description"`)}
	}
	return nil
}

func (rc *RepoCreate) sqlSave(ctx context.Context) (*Repo, error) {
	if err := rc.check(); err != nil {
		return nil, err
	}
	_node, _spec := rc.createSpec()
	if err := sqlgraph.CreateNode(ctx, rc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	rc.mutation.id = &_node.ID
	rc.mutation.done = true
	return _node, nil
}

func (rc *RepoCreate) createSpec() (*Repo, *sqlgraph.CreateSpec) {
	var (
		_node = &Repo{config: rc.config}
		_spec = sqlgraph.NewCreateSpec(repo.Table, sqlgraph.NewFieldSpec(repo.FieldID, field.TypeInt))
	)
	_spec.OnConflict = rc.conflict
	if value, ok := rc.mutation.CreatedAt(); ok {
		_spec.SetField(repo.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := rc.mutation.UpdatedAt(); ok {
		_spec.SetField(repo.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := rc.mutation.DeletedAt(); ok {
		_spec.SetField(repo.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = &value
	}
	if value, ok := rc.mutation.Name(); ok {
		_spec.SetField(repo.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := rc.mutation.DefaultBranch(); ok {
		_spec.SetField(repo.FieldDefaultBranch, field.TypeString, value)
		_node.DefaultBranch = value
	}
	if value, ok := rc.mutation.GitProjectName(); ok {
		_spec.SetField(repo.FieldGitProjectName, field.TypeString, value)
		_node.GitProjectName = value
	}
	if value, ok := rc.mutation.GitProjectID(); ok {
		_spec.SetField(repo.FieldGitProjectID, field.TypeInt32, value)
		_node.GitProjectID = value
	}
	if value, ok := rc.mutation.Enabled(); ok {
		_spec.SetField(repo.FieldEnabled, field.TypeBool, value)
		_node.Enabled = value
	}
	if value, ok := rc.mutation.NeedGitRepo(); ok {
		_spec.SetField(repo.FieldNeedGitRepo, field.TypeBool, value)
		_node.NeedGitRepo = value
	}
	if value, ok := rc.mutation.MarsConfig(); ok {
		_spec.SetField(repo.FieldMarsConfig, field.TypeJSON, value)
		_node.MarsConfig = value
	}
	if value, ok := rc.mutation.Description(); ok {
		_spec.SetField(repo.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if nodes := rc.mutation.ProjectsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   repo.ProjectsTable,
			Columns: []string{repo.ProjectsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(project.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Repo.Create().
//		SetCreatedAt(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.RepoUpsert) {
//			SetCreatedAt(v+v).
//		}).
//		Exec(ctx)
func (rc *RepoCreate) OnConflict(opts ...sql.ConflictOption) *RepoUpsertOne {
	rc.conflict = opts
	return &RepoUpsertOne{
		create: rc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Repo.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (rc *RepoCreate) OnConflictColumns(columns ...string) *RepoUpsertOne {
	rc.conflict = append(rc.conflict, sql.ConflictColumns(columns...))
	return &RepoUpsertOne{
		create: rc,
	}
}

type (
	// RepoUpsertOne is the builder for "upsert"-ing
	//  one Repo node.
	RepoUpsertOne struct {
		create *RepoCreate
	}

	// RepoUpsert is the "OnConflict" setter.
	RepoUpsert struct {
		*sql.UpdateSet
	}
)

// SetUpdatedAt sets the "updated_at" field.
func (u *RepoUpsert) SetUpdatedAt(v time.Time) *RepoUpsert {
	u.Set(repo.FieldUpdatedAt, v)
	return u
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *RepoUpsert) UpdateUpdatedAt() *RepoUpsert {
	u.SetExcluded(repo.FieldUpdatedAt)
	return u
}

// SetDeletedAt sets the "deleted_at" field.
func (u *RepoUpsert) SetDeletedAt(v time.Time) *RepoUpsert {
	u.Set(repo.FieldDeletedAt, v)
	return u
}

// UpdateDeletedAt sets the "deleted_at" field to the value that was provided on create.
func (u *RepoUpsert) UpdateDeletedAt() *RepoUpsert {
	u.SetExcluded(repo.FieldDeletedAt)
	return u
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (u *RepoUpsert) ClearDeletedAt() *RepoUpsert {
	u.SetNull(repo.FieldDeletedAt)
	return u
}

// SetName sets the "name" field.
func (u *RepoUpsert) SetName(v string) *RepoUpsert {
	u.Set(repo.FieldName, v)
	return u
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *RepoUpsert) UpdateName() *RepoUpsert {
	u.SetExcluded(repo.FieldName)
	return u
}

// SetDefaultBranch sets the "default_branch" field.
func (u *RepoUpsert) SetDefaultBranch(v string) *RepoUpsert {
	u.Set(repo.FieldDefaultBranch, v)
	return u
}

// UpdateDefaultBranch sets the "default_branch" field to the value that was provided on create.
func (u *RepoUpsert) UpdateDefaultBranch() *RepoUpsert {
	u.SetExcluded(repo.FieldDefaultBranch)
	return u
}

// ClearDefaultBranch clears the value of the "default_branch" field.
func (u *RepoUpsert) ClearDefaultBranch() *RepoUpsert {
	u.SetNull(repo.FieldDefaultBranch)
	return u
}

// SetGitProjectName sets the "git_project_name" field.
func (u *RepoUpsert) SetGitProjectName(v string) *RepoUpsert {
	u.Set(repo.FieldGitProjectName, v)
	return u
}

// UpdateGitProjectName sets the "git_project_name" field to the value that was provided on create.
func (u *RepoUpsert) UpdateGitProjectName() *RepoUpsert {
	u.SetExcluded(repo.FieldGitProjectName)
	return u
}

// ClearGitProjectName clears the value of the "git_project_name" field.
func (u *RepoUpsert) ClearGitProjectName() *RepoUpsert {
	u.SetNull(repo.FieldGitProjectName)
	return u
}

// SetGitProjectID sets the "git_project_id" field.
func (u *RepoUpsert) SetGitProjectID(v int32) *RepoUpsert {
	u.Set(repo.FieldGitProjectID, v)
	return u
}

// UpdateGitProjectID sets the "git_project_id" field to the value that was provided on create.
func (u *RepoUpsert) UpdateGitProjectID() *RepoUpsert {
	u.SetExcluded(repo.FieldGitProjectID)
	return u
}

// AddGitProjectID adds v to the "git_project_id" field.
func (u *RepoUpsert) AddGitProjectID(v int32) *RepoUpsert {
	u.Add(repo.FieldGitProjectID, v)
	return u
}

// ClearGitProjectID clears the value of the "git_project_id" field.
func (u *RepoUpsert) ClearGitProjectID() *RepoUpsert {
	u.SetNull(repo.FieldGitProjectID)
	return u
}

// SetEnabled sets the "enabled" field.
func (u *RepoUpsert) SetEnabled(v bool) *RepoUpsert {
	u.Set(repo.FieldEnabled, v)
	return u
}

// UpdateEnabled sets the "enabled" field to the value that was provided on create.
func (u *RepoUpsert) UpdateEnabled() *RepoUpsert {
	u.SetExcluded(repo.FieldEnabled)
	return u
}

// SetNeedGitRepo sets the "need_git_repo" field.
func (u *RepoUpsert) SetNeedGitRepo(v bool) *RepoUpsert {
	u.Set(repo.FieldNeedGitRepo, v)
	return u
}

// UpdateNeedGitRepo sets the "need_git_repo" field to the value that was provided on create.
func (u *RepoUpsert) UpdateNeedGitRepo() *RepoUpsert {
	u.SetExcluded(repo.FieldNeedGitRepo)
	return u
}

// SetMarsConfig sets the "mars_config" field.
func (u *RepoUpsert) SetMarsConfig(v *mars.Config) *RepoUpsert {
	u.Set(repo.FieldMarsConfig, v)
	return u
}

// UpdateMarsConfig sets the "mars_config" field to the value that was provided on create.
func (u *RepoUpsert) UpdateMarsConfig() *RepoUpsert {
	u.SetExcluded(repo.FieldMarsConfig)
	return u
}

// ClearMarsConfig clears the value of the "mars_config" field.
func (u *RepoUpsert) ClearMarsConfig() *RepoUpsert {
	u.SetNull(repo.FieldMarsConfig)
	return u
}

// SetDescription sets the "description" field.
func (u *RepoUpsert) SetDescription(v string) *RepoUpsert {
	u.Set(repo.FieldDescription, v)
	return u
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *RepoUpsert) UpdateDescription() *RepoUpsert {
	u.SetExcluded(repo.FieldDescription)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create.
// Using this option is equivalent to using:
//
//	client.Repo.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//		).
//		Exec(ctx)
func (u *RepoUpsertOne) UpdateNewValues() *RepoUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.CreatedAt(); exists {
			s.SetIgnore(repo.FieldCreatedAt)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Repo.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *RepoUpsertOne) Ignore() *RepoUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *RepoUpsertOne) DoNothing() *RepoUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the RepoCreate.OnConflict
// documentation for more info.
func (u *RepoUpsertOne) Update(set func(*RepoUpsert)) *RepoUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&RepoUpsert{UpdateSet: update})
	}))
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *RepoUpsertOne) SetUpdatedAt(v time.Time) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *RepoUpsertOne) UpdateUpdatedAt() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateUpdatedAt()
	})
}

// SetDeletedAt sets the "deleted_at" field.
func (u *RepoUpsertOne) SetDeletedAt(v time.Time) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.SetDeletedAt(v)
	})
}

// UpdateDeletedAt sets the "deleted_at" field to the value that was provided on create.
func (u *RepoUpsertOne) UpdateDeletedAt() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateDeletedAt()
	})
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (u *RepoUpsertOne) ClearDeletedAt() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.ClearDeletedAt()
	})
}

// SetName sets the "name" field.
func (u *RepoUpsertOne) SetName(v string) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *RepoUpsertOne) UpdateName() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateName()
	})
}

// SetDefaultBranch sets the "default_branch" field.
func (u *RepoUpsertOne) SetDefaultBranch(v string) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.SetDefaultBranch(v)
	})
}

// UpdateDefaultBranch sets the "default_branch" field to the value that was provided on create.
func (u *RepoUpsertOne) UpdateDefaultBranch() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateDefaultBranch()
	})
}

// ClearDefaultBranch clears the value of the "default_branch" field.
func (u *RepoUpsertOne) ClearDefaultBranch() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.ClearDefaultBranch()
	})
}

// SetGitProjectName sets the "git_project_name" field.
func (u *RepoUpsertOne) SetGitProjectName(v string) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.SetGitProjectName(v)
	})
}

// UpdateGitProjectName sets the "git_project_name" field to the value that was provided on create.
func (u *RepoUpsertOne) UpdateGitProjectName() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateGitProjectName()
	})
}

// ClearGitProjectName clears the value of the "git_project_name" field.
func (u *RepoUpsertOne) ClearGitProjectName() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.ClearGitProjectName()
	})
}

// SetGitProjectID sets the "git_project_id" field.
func (u *RepoUpsertOne) SetGitProjectID(v int32) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.SetGitProjectID(v)
	})
}

// AddGitProjectID adds v to the "git_project_id" field.
func (u *RepoUpsertOne) AddGitProjectID(v int32) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.AddGitProjectID(v)
	})
}

// UpdateGitProjectID sets the "git_project_id" field to the value that was provided on create.
func (u *RepoUpsertOne) UpdateGitProjectID() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateGitProjectID()
	})
}

// ClearGitProjectID clears the value of the "git_project_id" field.
func (u *RepoUpsertOne) ClearGitProjectID() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.ClearGitProjectID()
	})
}

// SetEnabled sets the "enabled" field.
func (u *RepoUpsertOne) SetEnabled(v bool) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.SetEnabled(v)
	})
}

// UpdateEnabled sets the "enabled" field to the value that was provided on create.
func (u *RepoUpsertOne) UpdateEnabled() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateEnabled()
	})
}

// SetNeedGitRepo sets the "need_git_repo" field.
func (u *RepoUpsertOne) SetNeedGitRepo(v bool) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.SetNeedGitRepo(v)
	})
}

// UpdateNeedGitRepo sets the "need_git_repo" field to the value that was provided on create.
func (u *RepoUpsertOne) UpdateNeedGitRepo() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateNeedGitRepo()
	})
}

// SetMarsConfig sets the "mars_config" field.
func (u *RepoUpsertOne) SetMarsConfig(v *mars.Config) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.SetMarsConfig(v)
	})
}

// UpdateMarsConfig sets the "mars_config" field to the value that was provided on create.
func (u *RepoUpsertOne) UpdateMarsConfig() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateMarsConfig()
	})
}

// ClearMarsConfig clears the value of the "mars_config" field.
func (u *RepoUpsertOne) ClearMarsConfig() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.ClearMarsConfig()
	})
}

// SetDescription sets the "description" field.
func (u *RepoUpsertOne) SetDescription(v string) *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *RepoUpsertOne) UpdateDescription() *RepoUpsertOne {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateDescription()
	})
}

// Exec executes the query.
func (u *RepoUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for RepoCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *RepoUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *RepoUpsertOne) ID(ctx context.Context) (id int, err error) {
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *RepoUpsertOne) IDX(ctx context.Context) int {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// RepoCreateBulk is the builder for creating many Repo entities in bulk.
type RepoCreateBulk struct {
	config
	err      error
	builders []*RepoCreate
	conflict []sql.ConflictOption
}

// Save creates the Repo entities in the database.
func (rcb *RepoCreateBulk) Save(ctx context.Context) ([]*Repo, error) {
	if rcb.err != nil {
		return nil, rcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(rcb.builders))
	nodes := make([]*Repo, len(rcb.builders))
	mutators := make([]Mutator, len(rcb.builders))
	for i := range rcb.builders {
		func(i int, root context.Context) {
			builder := rcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*RepoMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, rcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = rcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, rcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, rcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (rcb *RepoCreateBulk) SaveX(ctx context.Context) []*Repo {
	v, err := rcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (rcb *RepoCreateBulk) Exec(ctx context.Context) error {
	_, err := rcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rcb *RepoCreateBulk) ExecX(ctx context.Context) {
	if err := rcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Repo.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.RepoUpsert) {
//			SetCreatedAt(v+v).
//		}).
//		Exec(ctx)
func (rcb *RepoCreateBulk) OnConflict(opts ...sql.ConflictOption) *RepoUpsertBulk {
	rcb.conflict = opts
	return &RepoUpsertBulk{
		create: rcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Repo.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (rcb *RepoCreateBulk) OnConflictColumns(columns ...string) *RepoUpsertBulk {
	rcb.conflict = append(rcb.conflict, sql.ConflictColumns(columns...))
	return &RepoUpsertBulk{
		create: rcb,
	}
}

// RepoUpsertBulk is the builder for "upsert"-ing
// a bulk of Repo nodes.
type RepoUpsertBulk struct {
	create *RepoCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Repo.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//		).
//		Exec(ctx)
func (u *RepoUpsertBulk) UpdateNewValues() *RepoUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.CreatedAt(); exists {
				s.SetIgnore(repo.FieldCreatedAt)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Repo.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *RepoUpsertBulk) Ignore() *RepoUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *RepoUpsertBulk) DoNothing() *RepoUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the RepoCreateBulk.OnConflict
// documentation for more info.
func (u *RepoUpsertBulk) Update(set func(*RepoUpsert)) *RepoUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&RepoUpsert{UpdateSet: update})
	}))
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *RepoUpsertBulk) SetUpdatedAt(v time.Time) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *RepoUpsertBulk) UpdateUpdatedAt() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateUpdatedAt()
	})
}

// SetDeletedAt sets the "deleted_at" field.
func (u *RepoUpsertBulk) SetDeletedAt(v time.Time) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.SetDeletedAt(v)
	})
}

// UpdateDeletedAt sets the "deleted_at" field to the value that was provided on create.
func (u *RepoUpsertBulk) UpdateDeletedAt() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateDeletedAt()
	})
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (u *RepoUpsertBulk) ClearDeletedAt() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.ClearDeletedAt()
	})
}

// SetName sets the "name" field.
func (u *RepoUpsertBulk) SetName(v string) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *RepoUpsertBulk) UpdateName() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateName()
	})
}

// SetDefaultBranch sets the "default_branch" field.
func (u *RepoUpsertBulk) SetDefaultBranch(v string) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.SetDefaultBranch(v)
	})
}

// UpdateDefaultBranch sets the "default_branch" field to the value that was provided on create.
func (u *RepoUpsertBulk) UpdateDefaultBranch() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateDefaultBranch()
	})
}

// ClearDefaultBranch clears the value of the "default_branch" field.
func (u *RepoUpsertBulk) ClearDefaultBranch() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.ClearDefaultBranch()
	})
}

// SetGitProjectName sets the "git_project_name" field.
func (u *RepoUpsertBulk) SetGitProjectName(v string) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.SetGitProjectName(v)
	})
}

// UpdateGitProjectName sets the "git_project_name" field to the value that was provided on create.
func (u *RepoUpsertBulk) UpdateGitProjectName() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateGitProjectName()
	})
}

// ClearGitProjectName clears the value of the "git_project_name" field.
func (u *RepoUpsertBulk) ClearGitProjectName() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.ClearGitProjectName()
	})
}

// SetGitProjectID sets the "git_project_id" field.
func (u *RepoUpsertBulk) SetGitProjectID(v int32) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.SetGitProjectID(v)
	})
}

// AddGitProjectID adds v to the "git_project_id" field.
func (u *RepoUpsertBulk) AddGitProjectID(v int32) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.AddGitProjectID(v)
	})
}

// UpdateGitProjectID sets the "git_project_id" field to the value that was provided on create.
func (u *RepoUpsertBulk) UpdateGitProjectID() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateGitProjectID()
	})
}

// ClearGitProjectID clears the value of the "git_project_id" field.
func (u *RepoUpsertBulk) ClearGitProjectID() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.ClearGitProjectID()
	})
}

// SetEnabled sets the "enabled" field.
func (u *RepoUpsertBulk) SetEnabled(v bool) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.SetEnabled(v)
	})
}

// UpdateEnabled sets the "enabled" field to the value that was provided on create.
func (u *RepoUpsertBulk) UpdateEnabled() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateEnabled()
	})
}

// SetNeedGitRepo sets the "need_git_repo" field.
func (u *RepoUpsertBulk) SetNeedGitRepo(v bool) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.SetNeedGitRepo(v)
	})
}

// UpdateNeedGitRepo sets the "need_git_repo" field to the value that was provided on create.
func (u *RepoUpsertBulk) UpdateNeedGitRepo() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateNeedGitRepo()
	})
}

// SetMarsConfig sets the "mars_config" field.
func (u *RepoUpsertBulk) SetMarsConfig(v *mars.Config) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.SetMarsConfig(v)
	})
}

// UpdateMarsConfig sets the "mars_config" field to the value that was provided on create.
func (u *RepoUpsertBulk) UpdateMarsConfig() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateMarsConfig()
	})
}

// ClearMarsConfig clears the value of the "mars_config" field.
func (u *RepoUpsertBulk) ClearMarsConfig() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.ClearMarsConfig()
	})
}

// SetDescription sets the "description" field.
func (u *RepoUpsertBulk) SetDescription(v string) *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *RepoUpsertBulk) UpdateDescription() *RepoUpsertBulk {
	return u.Update(func(s *RepoUpsert) {
		s.UpdateDescription()
	})
}

// Exec executes the query.
func (u *RepoUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the RepoCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for RepoCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *RepoUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

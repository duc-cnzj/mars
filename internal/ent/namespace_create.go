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
	"github.com/duc-cnzj/mars/v5/internal/ent/favorite"
	"github.com/duc-cnzj/mars/v5/internal/ent/member"
	"github.com/duc-cnzj/mars/v5/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v5/internal/ent/project"
)

// NamespaceCreate is the builder for creating a Namespace entity.
type NamespaceCreate struct {
	config
	mutation *NamespaceMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetCreatedAt sets the "created_at" field.
func (nc *NamespaceCreate) SetCreatedAt(t time.Time) *NamespaceCreate {
	nc.mutation.SetCreatedAt(t)
	return nc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (nc *NamespaceCreate) SetNillableCreatedAt(t *time.Time) *NamespaceCreate {
	if t != nil {
		nc.SetCreatedAt(*t)
	}
	return nc
}

// SetUpdatedAt sets the "updated_at" field.
func (nc *NamespaceCreate) SetUpdatedAt(t time.Time) *NamespaceCreate {
	nc.mutation.SetUpdatedAt(t)
	return nc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (nc *NamespaceCreate) SetNillableUpdatedAt(t *time.Time) *NamespaceCreate {
	if t != nil {
		nc.SetUpdatedAt(*t)
	}
	return nc
}

// SetDeletedAt sets the "deleted_at" field.
func (nc *NamespaceCreate) SetDeletedAt(t time.Time) *NamespaceCreate {
	nc.mutation.SetDeletedAt(t)
	return nc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (nc *NamespaceCreate) SetNillableDeletedAt(t *time.Time) *NamespaceCreate {
	if t != nil {
		nc.SetDeletedAt(*t)
	}
	return nc
}

// SetName sets the "name" field.
func (nc *NamespaceCreate) SetName(s string) *NamespaceCreate {
	nc.mutation.SetName(s)
	return nc
}

// SetImagePullSecrets sets the "image_pull_secrets" field.
func (nc *NamespaceCreate) SetImagePullSecrets(s []string) *NamespaceCreate {
	nc.mutation.SetImagePullSecrets(s)
	return nc
}

// SetPrivate sets the "private" field.
func (nc *NamespaceCreate) SetPrivate(b bool) *NamespaceCreate {
	nc.mutation.SetPrivate(b)
	return nc
}

// SetNillablePrivate sets the "private" field if the given value is not nil.
func (nc *NamespaceCreate) SetNillablePrivate(b *bool) *NamespaceCreate {
	if b != nil {
		nc.SetPrivate(*b)
	}
	return nc
}

// SetCreatorEmail sets the "creator_email" field.
func (nc *NamespaceCreate) SetCreatorEmail(s string) *NamespaceCreate {
	nc.mutation.SetCreatorEmail(s)
	return nc
}

// SetDescription sets the "description" field.
func (nc *NamespaceCreate) SetDescription(s string) *NamespaceCreate {
	nc.mutation.SetDescription(s)
	return nc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (nc *NamespaceCreate) SetNillableDescription(s *string) *NamespaceCreate {
	if s != nil {
		nc.SetDescription(*s)
	}
	return nc
}

// AddProjectIDs adds the "projects" edge to the Project entity by IDs.
func (nc *NamespaceCreate) AddProjectIDs(ids ...int) *NamespaceCreate {
	nc.mutation.AddProjectIDs(ids...)
	return nc
}

// AddProjects adds the "projects" edges to the Project entity.
func (nc *NamespaceCreate) AddProjects(p ...*Project) *NamespaceCreate {
	ids := make([]int, len(p))
	for i := range p {
		ids[i] = p[i].ID
	}
	return nc.AddProjectIDs(ids...)
}

// AddFavoriteIDs adds the "favorites" edge to the Favorite entity by IDs.
func (nc *NamespaceCreate) AddFavoriteIDs(ids ...int) *NamespaceCreate {
	nc.mutation.AddFavoriteIDs(ids...)
	return nc
}

// AddFavorites adds the "favorites" edges to the Favorite entity.
func (nc *NamespaceCreate) AddFavorites(f ...*Favorite) *NamespaceCreate {
	ids := make([]int, len(f))
	for i := range f {
		ids[i] = f[i].ID
	}
	return nc.AddFavoriteIDs(ids...)
}

// AddMemberIDs adds the "members" edge to the Member entity by IDs.
func (nc *NamespaceCreate) AddMemberIDs(ids ...int) *NamespaceCreate {
	nc.mutation.AddMemberIDs(ids...)
	return nc
}

// AddMembers adds the "members" edges to the Member entity.
func (nc *NamespaceCreate) AddMembers(m ...*Member) *NamespaceCreate {
	ids := make([]int, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return nc.AddMemberIDs(ids...)
}

// Mutation returns the NamespaceMutation object of the builder.
func (nc *NamespaceCreate) Mutation() *NamespaceMutation {
	return nc.mutation
}

// Save creates the Namespace in the database.
func (nc *NamespaceCreate) Save(ctx context.Context) (*Namespace, error) {
	if err := nc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, nc.sqlSave, nc.mutation, nc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (nc *NamespaceCreate) SaveX(ctx context.Context) *Namespace {
	v, err := nc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (nc *NamespaceCreate) Exec(ctx context.Context) error {
	_, err := nc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (nc *NamespaceCreate) ExecX(ctx context.Context) {
	if err := nc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (nc *NamespaceCreate) defaults() error {
	if _, ok := nc.mutation.CreatedAt(); !ok {
		if namespace.DefaultCreatedAt == nil {
			return fmt.Errorf("ent: uninitialized namespace.DefaultCreatedAt (forgotten import ent/runtime?)")
		}
		v := namespace.DefaultCreatedAt()
		nc.mutation.SetCreatedAt(v)
	}
	if _, ok := nc.mutation.UpdatedAt(); !ok {
		if namespace.DefaultUpdatedAt == nil {
			return fmt.Errorf("ent: uninitialized namespace.DefaultUpdatedAt (forgotten import ent/runtime?)")
		}
		v := namespace.DefaultUpdatedAt()
		nc.mutation.SetUpdatedAt(v)
	}
	if _, ok := nc.mutation.ImagePullSecrets(); !ok {
		v := namespace.DefaultImagePullSecrets
		nc.mutation.SetImagePullSecrets(v)
	}
	if _, ok := nc.mutation.Private(); !ok {
		v := namespace.DefaultPrivate
		nc.mutation.SetPrivate(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (nc *NamespaceCreate) check() error {
	if _, ok := nc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Namespace.created_at"`)}
	}
	if _, ok := nc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "Namespace.updated_at"`)}
	}
	if _, ok := nc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "Namespace.name"`)}
	}
	if v, ok := nc.mutation.Name(); ok {
		if err := namespace.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`ent: validator failed for field "Namespace.name": %w`, err)}
		}
	}
	if _, ok := nc.mutation.ImagePullSecrets(); !ok {
		return &ValidationError{Name: "image_pull_secrets", err: errors.New(`ent: missing required field "Namespace.image_pull_secrets"`)}
	}
	if _, ok := nc.mutation.Private(); !ok {
		return &ValidationError{Name: "private", err: errors.New(`ent: missing required field "Namespace.private"`)}
	}
	if _, ok := nc.mutation.CreatorEmail(); !ok {
		return &ValidationError{Name: "creator_email", err: errors.New(`ent: missing required field "Namespace.creator_email"`)}
	}
	if v, ok := nc.mutation.CreatorEmail(); ok {
		if err := namespace.CreatorEmailValidator(v); err != nil {
			return &ValidationError{Name: "creator_email", err: fmt.Errorf(`ent: validator failed for field "Namespace.creator_email": %w`, err)}
		}
	}
	return nil
}

func (nc *NamespaceCreate) sqlSave(ctx context.Context) (*Namespace, error) {
	if err := nc.check(); err != nil {
		return nil, err
	}
	_node, _spec := nc.createSpec()
	if err := sqlgraph.CreateNode(ctx, nc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	nc.mutation.id = &_node.ID
	nc.mutation.done = true
	return _node, nil
}

func (nc *NamespaceCreate) createSpec() (*Namespace, *sqlgraph.CreateSpec) {
	var (
		_node = &Namespace{config: nc.config}
		_spec = sqlgraph.NewCreateSpec(namespace.Table, sqlgraph.NewFieldSpec(namespace.FieldID, field.TypeInt))
	)
	_spec.OnConflict = nc.conflict
	if value, ok := nc.mutation.CreatedAt(); ok {
		_spec.SetField(namespace.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := nc.mutation.UpdatedAt(); ok {
		_spec.SetField(namespace.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := nc.mutation.DeletedAt(); ok {
		_spec.SetField(namespace.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = &value
	}
	if value, ok := nc.mutation.Name(); ok {
		_spec.SetField(namespace.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := nc.mutation.ImagePullSecrets(); ok {
		_spec.SetField(namespace.FieldImagePullSecrets, field.TypeJSON, value)
		_node.ImagePullSecrets = value
	}
	if value, ok := nc.mutation.Private(); ok {
		_spec.SetField(namespace.FieldPrivate, field.TypeBool, value)
		_node.Private = value
	}
	if value, ok := nc.mutation.CreatorEmail(); ok {
		_spec.SetField(namespace.FieldCreatorEmail, field.TypeString, value)
		_node.CreatorEmail = value
	}
	if value, ok := nc.mutation.Description(); ok {
		_spec.SetField(namespace.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if nodes := nc.mutation.ProjectsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   namespace.ProjectsTable,
			Columns: []string{namespace.ProjectsColumn},
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
	if nodes := nc.mutation.FavoritesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   namespace.FavoritesTable,
			Columns: []string{namespace.FavoritesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(favorite.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := nc.mutation.MembersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   namespace.MembersTable,
			Columns: []string{namespace.MembersColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(member.FieldID, field.TypeInt),
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
//	client.Namespace.Create().
//		SetCreatedAt(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.NamespaceUpsert) {
//			SetCreatedAt(v+v).
//		}).
//		Exec(ctx)
func (nc *NamespaceCreate) OnConflict(opts ...sql.ConflictOption) *NamespaceUpsertOne {
	nc.conflict = opts
	return &NamespaceUpsertOne{
		create: nc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Namespace.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (nc *NamespaceCreate) OnConflictColumns(columns ...string) *NamespaceUpsertOne {
	nc.conflict = append(nc.conflict, sql.ConflictColumns(columns...))
	return &NamespaceUpsertOne{
		create: nc,
	}
}

type (
	// NamespaceUpsertOne is the builder for "upsert"-ing
	//  one Namespace node.
	NamespaceUpsertOne struct {
		create *NamespaceCreate
	}

	// NamespaceUpsert is the "OnConflict" setter.
	NamespaceUpsert struct {
		*sql.UpdateSet
	}
)

// SetUpdatedAt sets the "updated_at" field.
func (u *NamespaceUpsert) SetUpdatedAt(v time.Time) *NamespaceUpsert {
	u.Set(namespace.FieldUpdatedAt, v)
	return u
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *NamespaceUpsert) UpdateUpdatedAt() *NamespaceUpsert {
	u.SetExcluded(namespace.FieldUpdatedAt)
	return u
}

// SetDeletedAt sets the "deleted_at" field.
func (u *NamespaceUpsert) SetDeletedAt(v time.Time) *NamespaceUpsert {
	u.Set(namespace.FieldDeletedAt, v)
	return u
}

// UpdateDeletedAt sets the "deleted_at" field to the value that was provided on create.
func (u *NamespaceUpsert) UpdateDeletedAt() *NamespaceUpsert {
	u.SetExcluded(namespace.FieldDeletedAt)
	return u
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (u *NamespaceUpsert) ClearDeletedAt() *NamespaceUpsert {
	u.SetNull(namespace.FieldDeletedAt)
	return u
}

// SetName sets the "name" field.
func (u *NamespaceUpsert) SetName(v string) *NamespaceUpsert {
	u.Set(namespace.FieldName, v)
	return u
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *NamespaceUpsert) UpdateName() *NamespaceUpsert {
	u.SetExcluded(namespace.FieldName)
	return u
}

// SetImagePullSecrets sets the "image_pull_secrets" field.
func (u *NamespaceUpsert) SetImagePullSecrets(v []string) *NamespaceUpsert {
	u.Set(namespace.FieldImagePullSecrets, v)
	return u
}

// UpdateImagePullSecrets sets the "image_pull_secrets" field to the value that was provided on create.
func (u *NamespaceUpsert) UpdateImagePullSecrets() *NamespaceUpsert {
	u.SetExcluded(namespace.FieldImagePullSecrets)
	return u
}

// SetPrivate sets the "private" field.
func (u *NamespaceUpsert) SetPrivate(v bool) *NamespaceUpsert {
	u.Set(namespace.FieldPrivate, v)
	return u
}

// UpdatePrivate sets the "private" field to the value that was provided on create.
func (u *NamespaceUpsert) UpdatePrivate() *NamespaceUpsert {
	u.SetExcluded(namespace.FieldPrivate)
	return u
}

// SetCreatorEmail sets the "creator_email" field.
func (u *NamespaceUpsert) SetCreatorEmail(v string) *NamespaceUpsert {
	u.Set(namespace.FieldCreatorEmail, v)
	return u
}

// UpdateCreatorEmail sets the "creator_email" field to the value that was provided on create.
func (u *NamespaceUpsert) UpdateCreatorEmail() *NamespaceUpsert {
	u.SetExcluded(namespace.FieldCreatorEmail)
	return u
}

// SetDescription sets the "description" field.
func (u *NamespaceUpsert) SetDescription(v string) *NamespaceUpsert {
	u.Set(namespace.FieldDescription, v)
	return u
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *NamespaceUpsert) UpdateDescription() *NamespaceUpsert {
	u.SetExcluded(namespace.FieldDescription)
	return u
}

// ClearDescription clears the value of the "description" field.
func (u *NamespaceUpsert) ClearDescription() *NamespaceUpsert {
	u.SetNull(namespace.FieldDescription)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create.
// Using this option is equivalent to using:
//
//	client.Namespace.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//		).
//		Exec(ctx)
func (u *NamespaceUpsertOne) UpdateNewValues() *NamespaceUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.CreatedAt(); exists {
			s.SetIgnore(namespace.FieldCreatedAt)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Namespace.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *NamespaceUpsertOne) Ignore() *NamespaceUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *NamespaceUpsertOne) DoNothing() *NamespaceUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the NamespaceCreate.OnConflict
// documentation for more info.
func (u *NamespaceUpsertOne) Update(set func(*NamespaceUpsert)) *NamespaceUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&NamespaceUpsert{UpdateSet: update})
	}))
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *NamespaceUpsertOne) SetUpdatedAt(v time.Time) *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *NamespaceUpsertOne) UpdateUpdatedAt() *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateUpdatedAt()
	})
}

// SetDeletedAt sets the "deleted_at" field.
func (u *NamespaceUpsertOne) SetDeletedAt(v time.Time) *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetDeletedAt(v)
	})
}

// UpdateDeletedAt sets the "deleted_at" field to the value that was provided on create.
func (u *NamespaceUpsertOne) UpdateDeletedAt() *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateDeletedAt()
	})
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (u *NamespaceUpsertOne) ClearDeletedAt() *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.ClearDeletedAt()
	})
}

// SetName sets the "name" field.
func (u *NamespaceUpsertOne) SetName(v string) *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *NamespaceUpsertOne) UpdateName() *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateName()
	})
}

// SetImagePullSecrets sets the "image_pull_secrets" field.
func (u *NamespaceUpsertOne) SetImagePullSecrets(v []string) *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetImagePullSecrets(v)
	})
}

// UpdateImagePullSecrets sets the "image_pull_secrets" field to the value that was provided on create.
func (u *NamespaceUpsertOne) UpdateImagePullSecrets() *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateImagePullSecrets()
	})
}

// SetPrivate sets the "private" field.
func (u *NamespaceUpsertOne) SetPrivate(v bool) *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetPrivate(v)
	})
}

// UpdatePrivate sets the "private" field to the value that was provided on create.
func (u *NamespaceUpsertOne) UpdatePrivate() *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdatePrivate()
	})
}

// SetCreatorEmail sets the "creator_email" field.
func (u *NamespaceUpsertOne) SetCreatorEmail(v string) *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetCreatorEmail(v)
	})
}

// UpdateCreatorEmail sets the "creator_email" field to the value that was provided on create.
func (u *NamespaceUpsertOne) UpdateCreatorEmail() *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateCreatorEmail()
	})
}

// SetDescription sets the "description" field.
func (u *NamespaceUpsertOne) SetDescription(v string) *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *NamespaceUpsertOne) UpdateDescription() *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *NamespaceUpsertOne) ClearDescription() *NamespaceUpsertOne {
	return u.Update(func(s *NamespaceUpsert) {
		s.ClearDescription()
	})
}

// Exec executes the query.
func (u *NamespaceUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for NamespaceCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *NamespaceUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *NamespaceUpsertOne) ID(ctx context.Context) (id int, err error) {
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *NamespaceUpsertOne) IDX(ctx context.Context) int {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// NamespaceCreateBulk is the builder for creating many Namespace entities in bulk.
type NamespaceCreateBulk struct {
	config
	err      error
	builders []*NamespaceCreate
	conflict []sql.ConflictOption
}

// Save creates the Namespace entities in the database.
func (ncb *NamespaceCreateBulk) Save(ctx context.Context) ([]*Namespace, error) {
	if ncb.err != nil {
		return nil, ncb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ncb.builders))
	nodes := make([]*Namespace, len(ncb.builders))
	mutators := make([]Mutator, len(ncb.builders))
	for i := range ncb.builders {
		func(i int, root context.Context) {
			builder := ncb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*NamespaceMutation)
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
					_, err = mutators[i+1].Mutate(root, ncb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = ncb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ncb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ncb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ncb *NamespaceCreateBulk) SaveX(ctx context.Context) []*Namespace {
	v, err := ncb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ncb *NamespaceCreateBulk) Exec(ctx context.Context) error {
	_, err := ncb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ncb *NamespaceCreateBulk) ExecX(ctx context.Context) {
	if err := ncb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Namespace.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.NamespaceUpsert) {
//			SetCreatedAt(v+v).
//		}).
//		Exec(ctx)
func (ncb *NamespaceCreateBulk) OnConflict(opts ...sql.ConflictOption) *NamespaceUpsertBulk {
	ncb.conflict = opts
	return &NamespaceUpsertBulk{
		create: ncb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Namespace.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ncb *NamespaceCreateBulk) OnConflictColumns(columns ...string) *NamespaceUpsertBulk {
	ncb.conflict = append(ncb.conflict, sql.ConflictColumns(columns...))
	return &NamespaceUpsertBulk{
		create: ncb,
	}
}

// NamespaceUpsertBulk is the builder for "upsert"-ing
// a bulk of Namespace nodes.
type NamespaceUpsertBulk struct {
	create *NamespaceCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Namespace.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//		).
//		Exec(ctx)
func (u *NamespaceUpsertBulk) UpdateNewValues() *NamespaceUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.CreatedAt(); exists {
				s.SetIgnore(namespace.FieldCreatedAt)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Namespace.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *NamespaceUpsertBulk) Ignore() *NamespaceUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *NamespaceUpsertBulk) DoNothing() *NamespaceUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the NamespaceCreateBulk.OnConflict
// documentation for more info.
func (u *NamespaceUpsertBulk) Update(set func(*NamespaceUpsert)) *NamespaceUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&NamespaceUpsert{UpdateSet: update})
	}))
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *NamespaceUpsertBulk) SetUpdatedAt(v time.Time) *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *NamespaceUpsertBulk) UpdateUpdatedAt() *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateUpdatedAt()
	})
}

// SetDeletedAt sets the "deleted_at" field.
func (u *NamespaceUpsertBulk) SetDeletedAt(v time.Time) *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetDeletedAt(v)
	})
}

// UpdateDeletedAt sets the "deleted_at" field to the value that was provided on create.
func (u *NamespaceUpsertBulk) UpdateDeletedAt() *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateDeletedAt()
	})
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (u *NamespaceUpsertBulk) ClearDeletedAt() *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.ClearDeletedAt()
	})
}

// SetName sets the "name" field.
func (u *NamespaceUpsertBulk) SetName(v string) *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *NamespaceUpsertBulk) UpdateName() *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateName()
	})
}

// SetImagePullSecrets sets the "image_pull_secrets" field.
func (u *NamespaceUpsertBulk) SetImagePullSecrets(v []string) *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetImagePullSecrets(v)
	})
}

// UpdateImagePullSecrets sets the "image_pull_secrets" field to the value that was provided on create.
func (u *NamespaceUpsertBulk) UpdateImagePullSecrets() *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateImagePullSecrets()
	})
}

// SetPrivate sets the "private" field.
func (u *NamespaceUpsertBulk) SetPrivate(v bool) *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetPrivate(v)
	})
}

// UpdatePrivate sets the "private" field to the value that was provided on create.
func (u *NamespaceUpsertBulk) UpdatePrivate() *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdatePrivate()
	})
}

// SetCreatorEmail sets the "creator_email" field.
func (u *NamespaceUpsertBulk) SetCreatorEmail(v string) *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetCreatorEmail(v)
	})
}

// UpdateCreatorEmail sets the "creator_email" field to the value that was provided on create.
func (u *NamespaceUpsertBulk) UpdateCreatorEmail() *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateCreatorEmail()
	})
}

// SetDescription sets the "description" field.
func (u *NamespaceUpsertBulk) SetDescription(v string) *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.SetDescription(v)
	})
}

// UpdateDescription sets the "description" field to the value that was provided on create.
func (u *NamespaceUpsertBulk) UpdateDescription() *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.UpdateDescription()
	})
}

// ClearDescription clears the value of the "description" field.
func (u *NamespaceUpsertBulk) ClearDescription() *NamespaceUpsertBulk {
	return u.Update(func(s *NamespaceUpsert) {
		s.ClearDescription()
	})
}

// Exec executes the query.
func (u *NamespaceUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the NamespaceCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for NamespaceCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *NamespaceUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

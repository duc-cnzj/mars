// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/duc-cnzj/mars/v5/internal/ent/favorite"
	"github.com/duc-cnzj/mars/v5/internal/ent/namespace"
)

// FavoriteCreate is the builder for creating a Favorite entity.
type FavoriteCreate struct {
	config
	mutation *FavoriteMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetEmail sets the "email" field.
func (fc *FavoriteCreate) SetEmail(s string) *FavoriteCreate {
	fc.mutation.SetEmail(s)
	return fc
}

// SetNamespaceID sets the "namespace_id" field.
func (fc *FavoriteCreate) SetNamespaceID(i int) *FavoriteCreate {
	fc.mutation.SetNamespaceID(i)
	return fc
}

// SetNillableNamespaceID sets the "namespace_id" field if the given value is not nil.
func (fc *FavoriteCreate) SetNillableNamespaceID(i *int) *FavoriteCreate {
	if i != nil {
		fc.SetNamespaceID(*i)
	}
	return fc
}

// SetNamespace sets the "namespace" edge to the Namespace entity.
func (fc *FavoriteCreate) SetNamespace(n *Namespace) *FavoriteCreate {
	return fc.SetNamespaceID(n.ID)
}

// Mutation returns the FavoriteMutation object of the builder.
func (fc *FavoriteCreate) Mutation() *FavoriteMutation {
	return fc.mutation
}

// Save creates the Favorite in the database.
func (fc *FavoriteCreate) Save(ctx context.Context) (*Favorite, error) {
	return withHooks(ctx, fc.sqlSave, fc.mutation, fc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (fc *FavoriteCreate) SaveX(ctx context.Context) *Favorite {
	v, err := fc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (fc *FavoriteCreate) Exec(ctx context.Context) error {
	_, err := fc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fc *FavoriteCreate) ExecX(ctx context.Context) {
	if err := fc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (fc *FavoriteCreate) check() error {
	if _, ok := fc.mutation.Email(); !ok {
		return &ValidationError{Name: "email", err: errors.New(`ent: missing required field "Favorite.email"`)}
	}
	if v, ok := fc.mutation.Email(); ok {
		if err := favorite.EmailValidator(v); err != nil {
			return &ValidationError{Name: "email", err: fmt.Errorf(`ent: validator failed for field "Favorite.email": %w`, err)}
		}
	}
	return nil
}

func (fc *FavoriteCreate) sqlSave(ctx context.Context) (*Favorite, error) {
	if err := fc.check(); err != nil {
		return nil, err
	}
	_node, _spec := fc.createSpec()
	if err := sqlgraph.CreateNode(ctx, fc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	fc.mutation.id = &_node.ID
	fc.mutation.done = true
	return _node, nil
}

func (fc *FavoriteCreate) createSpec() (*Favorite, *sqlgraph.CreateSpec) {
	var (
		_node = &Favorite{config: fc.config}
		_spec = sqlgraph.NewCreateSpec(favorite.Table, sqlgraph.NewFieldSpec(favorite.FieldID, field.TypeInt))
	)
	_spec.OnConflict = fc.conflict
	if value, ok := fc.mutation.Email(); ok {
		_spec.SetField(favorite.FieldEmail, field.TypeString, value)
		_node.Email = value
	}
	if nodes := fc.mutation.NamespaceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   favorite.NamespaceTable,
			Columns: []string{favorite.NamespaceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(namespace.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.NamespaceID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Favorite.Create().
//		SetEmail(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.FavoriteUpsert) {
//			SetEmail(v+v).
//		}).
//		Exec(ctx)
func (fc *FavoriteCreate) OnConflict(opts ...sql.ConflictOption) *FavoriteUpsertOne {
	fc.conflict = opts
	return &FavoriteUpsertOne{
		create: fc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Favorite.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (fc *FavoriteCreate) OnConflictColumns(columns ...string) *FavoriteUpsertOne {
	fc.conflict = append(fc.conflict, sql.ConflictColumns(columns...))
	return &FavoriteUpsertOne{
		create: fc,
	}
}

type (
	// FavoriteUpsertOne is the builder for "upsert"-ing
	//  one Favorite node.
	FavoriteUpsertOne struct {
		create *FavoriteCreate
	}

	// FavoriteUpsert is the "OnConflict" setter.
	FavoriteUpsert struct {
		*sql.UpdateSet
	}
)

// SetEmail sets the "email" field.
func (u *FavoriteUpsert) SetEmail(v string) *FavoriteUpsert {
	u.Set(favorite.FieldEmail, v)
	return u
}

// UpdateEmail sets the "email" field to the value that was provided on create.
func (u *FavoriteUpsert) UpdateEmail() *FavoriteUpsert {
	u.SetExcluded(favorite.FieldEmail)
	return u
}

// SetNamespaceID sets the "namespace_id" field.
func (u *FavoriteUpsert) SetNamespaceID(v int) *FavoriteUpsert {
	u.Set(favorite.FieldNamespaceID, v)
	return u
}

// UpdateNamespaceID sets the "namespace_id" field to the value that was provided on create.
func (u *FavoriteUpsert) UpdateNamespaceID() *FavoriteUpsert {
	u.SetExcluded(favorite.FieldNamespaceID)
	return u
}

// ClearNamespaceID clears the value of the "namespace_id" field.
func (u *FavoriteUpsert) ClearNamespaceID() *FavoriteUpsert {
	u.SetNull(favorite.FieldNamespaceID)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create.
// Using this option is equivalent to using:
//
//	client.Favorite.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//		).
//		Exec(ctx)
func (u *FavoriteUpsertOne) UpdateNewValues() *FavoriteUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Favorite.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *FavoriteUpsertOne) Ignore() *FavoriteUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *FavoriteUpsertOne) DoNothing() *FavoriteUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the FavoriteCreate.OnConflict
// documentation for more info.
func (u *FavoriteUpsertOne) Update(set func(*FavoriteUpsert)) *FavoriteUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&FavoriteUpsert{UpdateSet: update})
	}))
	return u
}

// SetEmail sets the "email" field.
func (u *FavoriteUpsertOne) SetEmail(v string) *FavoriteUpsertOne {
	return u.Update(func(s *FavoriteUpsert) {
		s.SetEmail(v)
	})
}

// UpdateEmail sets the "email" field to the value that was provided on create.
func (u *FavoriteUpsertOne) UpdateEmail() *FavoriteUpsertOne {
	return u.Update(func(s *FavoriteUpsert) {
		s.UpdateEmail()
	})
}

// SetNamespaceID sets the "namespace_id" field.
func (u *FavoriteUpsertOne) SetNamespaceID(v int) *FavoriteUpsertOne {
	return u.Update(func(s *FavoriteUpsert) {
		s.SetNamespaceID(v)
	})
}

// UpdateNamespaceID sets the "namespace_id" field to the value that was provided on create.
func (u *FavoriteUpsertOne) UpdateNamespaceID() *FavoriteUpsertOne {
	return u.Update(func(s *FavoriteUpsert) {
		s.UpdateNamespaceID()
	})
}

// ClearNamespaceID clears the value of the "namespace_id" field.
func (u *FavoriteUpsertOne) ClearNamespaceID() *FavoriteUpsertOne {
	return u.Update(func(s *FavoriteUpsert) {
		s.ClearNamespaceID()
	})
}

// Exec executes the query.
func (u *FavoriteUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for FavoriteCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *FavoriteUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *FavoriteUpsertOne) ID(ctx context.Context) (id int, err error) {
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *FavoriteUpsertOne) IDX(ctx context.Context) int {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// FavoriteCreateBulk is the builder for creating many Favorite entities in bulk.
type FavoriteCreateBulk struct {
	config
	err      error
	builders []*FavoriteCreate
	conflict []sql.ConflictOption
}

// Save creates the Favorite entities in the database.
func (fcb *FavoriteCreateBulk) Save(ctx context.Context) ([]*Favorite, error) {
	if fcb.err != nil {
		return nil, fcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(fcb.builders))
	nodes := make([]*Favorite, len(fcb.builders))
	mutators := make([]Mutator, len(fcb.builders))
	for i := range fcb.builders {
		func(i int, root context.Context) {
			builder := fcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*FavoriteMutation)
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
					_, err = mutators[i+1].Mutate(root, fcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = fcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, fcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, fcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (fcb *FavoriteCreateBulk) SaveX(ctx context.Context) []*Favorite {
	v, err := fcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (fcb *FavoriteCreateBulk) Exec(ctx context.Context) error {
	_, err := fcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fcb *FavoriteCreateBulk) ExecX(ctx context.Context) {
	if err := fcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Favorite.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.FavoriteUpsert) {
//			SetEmail(v+v).
//		}).
//		Exec(ctx)
func (fcb *FavoriteCreateBulk) OnConflict(opts ...sql.ConflictOption) *FavoriteUpsertBulk {
	fcb.conflict = opts
	return &FavoriteUpsertBulk{
		create: fcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Favorite.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (fcb *FavoriteCreateBulk) OnConflictColumns(columns ...string) *FavoriteUpsertBulk {
	fcb.conflict = append(fcb.conflict, sql.ConflictColumns(columns...))
	return &FavoriteUpsertBulk{
		create: fcb,
	}
}

// FavoriteUpsertBulk is the builder for "upsert"-ing
// a bulk of Favorite nodes.
type FavoriteUpsertBulk struct {
	create *FavoriteCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Favorite.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//		).
//		Exec(ctx)
func (u *FavoriteUpsertBulk) UpdateNewValues() *FavoriteUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Favorite.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *FavoriteUpsertBulk) Ignore() *FavoriteUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *FavoriteUpsertBulk) DoNothing() *FavoriteUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the FavoriteCreateBulk.OnConflict
// documentation for more info.
func (u *FavoriteUpsertBulk) Update(set func(*FavoriteUpsert)) *FavoriteUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&FavoriteUpsert{UpdateSet: update})
	}))
	return u
}

// SetEmail sets the "email" field.
func (u *FavoriteUpsertBulk) SetEmail(v string) *FavoriteUpsertBulk {
	return u.Update(func(s *FavoriteUpsert) {
		s.SetEmail(v)
	})
}

// UpdateEmail sets the "email" field to the value that was provided on create.
func (u *FavoriteUpsertBulk) UpdateEmail() *FavoriteUpsertBulk {
	return u.Update(func(s *FavoriteUpsert) {
		s.UpdateEmail()
	})
}

// SetNamespaceID sets the "namespace_id" field.
func (u *FavoriteUpsertBulk) SetNamespaceID(v int) *FavoriteUpsertBulk {
	return u.Update(func(s *FavoriteUpsert) {
		s.SetNamespaceID(v)
	})
}

// UpdateNamespaceID sets the "namespace_id" field to the value that was provided on create.
func (u *FavoriteUpsertBulk) UpdateNamespaceID() *FavoriteUpsertBulk {
	return u.Update(func(s *FavoriteUpsert) {
		s.UpdateNamespaceID()
	})
}

// ClearNamespaceID clears the value of the "namespace_id" field.
func (u *FavoriteUpsertBulk) ClearNamespaceID() *FavoriteUpsertBulk {
	return u.Update(func(s *FavoriteUpsert) {
		s.ClearNamespaceID()
	})
}

// Exec executes the query.
func (u *FavoriteUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the FavoriteCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for FavoriteCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *FavoriteUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

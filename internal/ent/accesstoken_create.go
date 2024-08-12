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
	"github.com/duc-cnzj/mars/v4/internal/ent/accesstoken"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/schematype"
)

// AccessTokenCreate is the builder for creating a AccessToken entity.
type AccessTokenCreate struct {
	config
	mutation *AccessTokenMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetCreatedAt sets the "created_at" field.
func (atc *AccessTokenCreate) SetCreatedAt(t time.Time) *AccessTokenCreate {
	atc.mutation.SetCreatedAt(t)
	return atc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (atc *AccessTokenCreate) SetNillableCreatedAt(t *time.Time) *AccessTokenCreate {
	if t != nil {
		atc.SetCreatedAt(*t)
	}
	return atc
}

// SetUpdatedAt sets the "updated_at" field.
func (atc *AccessTokenCreate) SetUpdatedAt(t time.Time) *AccessTokenCreate {
	atc.mutation.SetUpdatedAt(t)
	return atc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (atc *AccessTokenCreate) SetNillableUpdatedAt(t *time.Time) *AccessTokenCreate {
	if t != nil {
		atc.SetUpdatedAt(*t)
	}
	return atc
}

// SetDeletedAt sets the "deleted_at" field.
func (atc *AccessTokenCreate) SetDeletedAt(t time.Time) *AccessTokenCreate {
	atc.mutation.SetDeletedAt(t)
	return atc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (atc *AccessTokenCreate) SetNillableDeletedAt(t *time.Time) *AccessTokenCreate {
	if t != nil {
		atc.SetDeletedAt(*t)
	}
	return atc
}

// SetToken sets the "token" field.
func (atc *AccessTokenCreate) SetToken(s string) *AccessTokenCreate {
	atc.mutation.SetToken(s)
	return atc
}

// SetUsage sets the "usage" field.
func (atc *AccessTokenCreate) SetUsage(s string) *AccessTokenCreate {
	atc.mutation.SetUsage(s)
	return atc
}

// SetEmail sets the "email" field.
func (atc *AccessTokenCreate) SetEmail(s string) *AccessTokenCreate {
	atc.mutation.SetEmail(s)
	return atc
}

// SetNillableEmail sets the "email" field if the given value is not nil.
func (atc *AccessTokenCreate) SetNillableEmail(s *string) *AccessTokenCreate {
	if s != nil {
		atc.SetEmail(*s)
	}
	return atc
}

// SetExpiredAt sets the "expired_at" field.
func (atc *AccessTokenCreate) SetExpiredAt(t time.Time) *AccessTokenCreate {
	atc.mutation.SetExpiredAt(t)
	return atc
}

// SetNillableExpiredAt sets the "expired_at" field if the given value is not nil.
func (atc *AccessTokenCreate) SetNillableExpiredAt(t *time.Time) *AccessTokenCreate {
	if t != nil {
		atc.SetExpiredAt(*t)
	}
	return atc
}

// SetLastUsedAt sets the "last_used_at" field.
func (atc *AccessTokenCreate) SetLastUsedAt(t time.Time) *AccessTokenCreate {
	atc.mutation.SetLastUsedAt(t)
	return atc
}

// SetNillableLastUsedAt sets the "last_used_at" field if the given value is not nil.
func (atc *AccessTokenCreate) SetNillableLastUsedAt(t *time.Time) *AccessTokenCreate {
	if t != nil {
		atc.SetLastUsedAt(*t)
	}
	return atc
}

// SetUserInfo sets the "user_info" field.
func (atc *AccessTokenCreate) SetUserInfo(si schematype.UserInfo) *AccessTokenCreate {
	atc.mutation.SetUserInfo(si)
	return atc
}

// SetNillableUserInfo sets the "user_info" field if the given value is not nil.
func (atc *AccessTokenCreate) SetNillableUserInfo(si *schematype.UserInfo) *AccessTokenCreate {
	if si != nil {
		atc.SetUserInfo(*si)
	}
	return atc
}

// Mutation returns the AccessTokenMutation object of the builder.
func (atc *AccessTokenCreate) Mutation() *AccessTokenMutation {
	return atc.mutation
}

// Save creates the AccessToken in the database.
func (atc *AccessTokenCreate) Save(ctx context.Context) (*AccessToken, error) {
	if err := atc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, atc.sqlSave, atc.mutation, atc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (atc *AccessTokenCreate) SaveX(ctx context.Context) *AccessToken {
	v, err := atc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (atc *AccessTokenCreate) Exec(ctx context.Context) error {
	_, err := atc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (atc *AccessTokenCreate) ExecX(ctx context.Context) {
	if err := atc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (atc *AccessTokenCreate) defaults() error {
	if _, ok := atc.mutation.CreatedAt(); !ok {
		if accesstoken.DefaultCreatedAt == nil {
			return fmt.Errorf("ent: uninitialized accesstoken.DefaultCreatedAt (forgotten import ent/runtime?)")
		}
		v := accesstoken.DefaultCreatedAt()
		atc.mutation.SetCreatedAt(v)
	}
	if _, ok := atc.mutation.UpdatedAt(); !ok {
		if accesstoken.DefaultUpdatedAt == nil {
			return fmt.Errorf("ent: uninitialized accesstoken.DefaultUpdatedAt (forgotten import ent/runtime?)")
		}
		v := accesstoken.DefaultUpdatedAt()
		atc.mutation.SetUpdatedAt(v)
	}
	if _, ok := atc.mutation.Email(); !ok {
		v := accesstoken.DefaultEmail
		atc.mutation.SetEmail(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (atc *AccessTokenCreate) check() error {
	if _, ok := atc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "AccessToken.created_at"`)}
	}
	if _, ok := atc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "AccessToken.updated_at"`)}
	}
	if _, ok := atc.mutation.Token(); !ok {
		return &ValidationError{Name: "token", err: errors.New(`ent: missing required field "AccessToken.token"`)}
	}
	if v, ok := atc.mutation.Token(); ok {
		if err := accesstoken.TokenValidator(v); err != nil {
			return &ValidationError{Name: "token", err: fmt.Errorf(`ent: validator failed for field "AccessToken.token": %w`, err)}
		}
	}
	if _, ok := atc.mutation.Usage(); !ok {
		return &ValidationError{Name: "usage", err: errors.New(`ent: missing required field "AccessToken.usage"`)}
	}
	if v, ok := atc.mutation.Usage(); ok {
		if err := accesstoken.UsageValidator(v); err != nil {
			return &ValidationError{Name: "usage", err: fmt.Errorf(`ent: validator failed for field "AccessToken.usage": %w`, err)}
		}
	}
	if _, ok := atc.mutation.Email(); !ok {
		return &ValidationError{Name: "email", err: errors.New(`ent: missing required field "AccessToken.email"`)}
	}
	if v, ok := atc.mutation.Email(); ok {
		if err := accesstoken.EmailValidator(v); err != nil {
			return &ValidationError{Name: "email", err: fmt.Errorf(`ent: validator failed for field "AccessToken.email": %w`, err)}
		}
	}
	return nil
}

func (atc *AccessTokenCreate) sqlSave(ctx context.Context) (*AccessToken, error) {
	if err := atc.check(); err != nil {
		return nil, err
	}
	_node, _spec := atc.createSpec()
	if err := sqlgraph.CreateNode(ctx, atc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	atc.mutation.id = &_node.ID
	atc.mutation.done = true
	return _node, nil
}

func (atc *AccessTokenCreate) createSpec() (*AccessToken, *sqlgraph.CreateSpec) {
	var (
		_node = &AccessToken{config: atc.config}
		_spec = sqlgraph.NewCreateSpec(accesstoken.Table, sqlgraph.NewFieldSpec(accesstoken.FieldID, field.TypeInt))
	)
	_spec.OnConflict = atc.conflict
	if value, ok := atc.mutation.CreatedAt(); ok {
		_spec.SetField(accesstoken.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := atc.mutation.UpdatedAt(); ok {
		_spec.SetField(accesstoken.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := atc.mutation.DeletedAt(); ok {
		_spec.SetField(accesstoken.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = &value
	}
	if value, ok := atc.mutation.Token(); ok {
		_spec.SetField(accesstoken.FieldToken, field.TypeString, value)
		_node.Token = value
	}
	if value, ok := atc.mutation.Usage(); ok {
		_spec.SetField(accesstoken.FieldUsage, field.TypeString, value)
		_node.Usage = value
	}
	if value, ok := atc.mutation.Email(); ok {
		_spec.SetField(accesstoken.FieldEmail, field.TypeString, value)
		_node.Email = value
	}
	if value, ok := atc.mutation.ExpiredAt(); ok {
		_spec.SetField(accesstoken.FieldExpiredAt, field.TypeTime, value)
		_node.ExpiredAt = value
	}
	if value, ok := atc.mutation.LastUsedAt(); ok {
		_spec.SetField(accesstoken.FieldLastUsedAt, field.TypeTime, value)
		_node.LastUsedAt = &value
	}
	if value, ok := atc.mutation.UserInfo(); ok {
		_spec.SetField(accesstoken.FieldUserInfo, field.TypeJSON, value)
		_node.UserInfo = value
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.AccessToken.Create().
//		SetCreatedAt(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.AccessTokenUpsert) {
//			SetCreatedAt(v+v).
//		}).
//		Exec(ctx)
func (atc *AccessTokenCreate) OnConflict(opts ...sql.ConflictOption) *AccessTokenUpsertOne {
	atc.conflict = opts
	return &AccessTokenUpsertOne{
		create: atc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.AccessToken.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (atc *AccessTokenCreate) OnConflictColumns(columns ...string) *AccessTokenUpsertOne {
	atc.conflict = append(atc.conflict, sql.ConflictColumns(columns...))
	return &AccessTokenUpsertOne{
		create: atc,
	}
}

type (
	// AccessTokenUpsertOne is the builder for "upsert"-ing
	//  one AccessToken node.
	AccessTokenUpsertOne struct {
		create *AccessTokenCreate
	}

	// AccessTokenUpsert is the "OnConflict" setter.
	AccessTokenUpsert struct {
		*sql.UpdateSet
	}
)

// SetUpdatedAt sets the "updated_at" field.
func (u *AccessTokenUpsert) SetUpdatedAt(v time.Time) *AccessTokenUpsert {
	u.Set(accesstoken.FieldUpdatedAt, v)
	return u
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *AccessTokenUpsert) UpdateUpdatedAt() *AccessTokenUpsert {
	u.SetExcluded(accesstoken.FieldUpdatedAt)
	return u
}

// SetDeletedAt sets the "deleted_at" field.
func (u *AccessTokenUpsert) SetDeletedAt(v time.Time) *AccessTokenUpsert {
	u.Set(accesstoken.FieldDeletedAt, v)
	return u
}

// UpdateDeletedAt sets the "deleted_at" field to the value that was provided on create.
func (u *AccessTokenUpsert) UpdateDeletedAt() *AccessTokenUpsert {
	u.SetExcluded(accesstoken.FieldDeletedAt)
	return u
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (u *AccessTokenUpsert) ClearDeletedAt() *AccessTokenUpsert {
	u.SetNull(accesstoken.FieldDeletedAt)
	return u
}

// SetToken sets the "token" field.
func (u *AccessTokenUpsert) SetToken(v string) *AccessTokenUpsert {
	u.Set(accesstoken.FieldToken, v)
	return u
}

// UpdateToken sets the "token" field to the value that was provided on create.
func (u *AccessTokenUpsert) UpdateToken() *AccessTokenUpsert {
	u.SetExcluded(accesstoken.FieldToken)
	return u
}

// SetUsage sets the "usage" field.
func (u *AccessTokenUpsert) SetUsage(v string) *AccessTokenUpsert {
	u.Set(accesstoken.FieldUsage, v)
	return u
}

// UpdateUsage sets the "usage" field to the value that was provided on create.
func (u *AccessTokenUpsert) UpdateUsage() *AccessTokenUpsert {
	u.SetExcluded(accesstoken.FieldUsage)
	return u
}

// SetEmail sets the "email" field.
func (u *AccessTokenUpsert) SetEmail(v string) *AccessTokenUpsert {
	u.Set(accesstoken.FieldEmail, v)
	return u
}

// UpdateEmail sets the "email" field to the value that was provided on create.
func (u *AccessTokenUpsert) UpdateEmail() *AccessTokenUpsert {
	u.SetExcluded(accesstoken.FieldEmail)
	return u
}

// SetExpiredAt sets the "expired_at" field.
func (u *AccessTokenUpsert) SetExpiredAt(v time.Time) *AccessTokenUpsert {
	u.Set(accesstoken.FieldExpiredAt, v)
	return u
}

// UpdateExpiredAt sets the "expired_at" field to the value that was provided on create.
func (u *AccessTokenUpsert) UpdateExpiredAt() *AccessTokenUpsert {
	u.SetExcluded(accesstoken.FieldExpiredAt)
	return u
}

// ClearExpiredAt clears the value of the "expired_at" field.
func (u *AccessTokenUpsert) ClearExpiredAt() *AccessTokenUpsert {
	u.SetNull(accesstoken.FieldExpiredAt)
	return u
}

// SetLastUsedAt sets the "last_used_at" field.
func (u *AccessTokenUpsert) SetLastUsedAt(v time.Time) *AccessTokenUpsert {
	u.Set(accesstoken.FieldLastUsedAt, v)
	return u
}

// UpdateLastUsedAt sets the "last_used_at" field to the value that was provided on create.
func (u *AccessTokenUpsert) UpdateLastUsedAt() *AccessTokenUpsert {
	u.SetExcluded(accesstoken.FieldLastUsedAt)
	return u
}

// ClearLastUsedAt clears the value of the "last_used_at" field.
func (u *AccessTokenUpsert) ClearLastUsedAt() *AccessTokenUpsert {
	u.SetNull(accesstoken.FieldLastUsedAt)
	return u
}

// SetUserInfo sets the "user_info" field.
func (u *AccessTokenUpsert) SetUserInfo(v schematype.UserInfo) *AccessTokenUpsert {
	u.Set(accesstoken.FieldUserInfo, v)
	return u
}

// UpdateUserInfo sets the "user_info" field to the value that was provided on create.
func (u *AccessTokenUpsert) UpdateUserInfo() *AccessTokenUpsert {
	u.SetExcluded(accesstoken.FieldUserInfo)
	return u
}

// ClearUserInfo clears the value of the "user_info" field.
func (u *AccessTokenUpsert) ClearUserInfo() *AccessTokenUpsert {
	u.SetNull(accesstoken.FieldUserInfo)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create.
// Using this option is equivalent to using:
//
//	client.AccessToken.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//		).
//		Exec(ctx)
func (u *AccessTokenUpsertOne) UpdateNewValues() *AccessTokenUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.CreatedAt(); exists {
			s.SetIgnore(accesstoken.FieldCreatedAt)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.AccessToken.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *AccessTokenUpsertOne) Ignore() *AccessTokenUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *AccessTokenUpsertOne) DoNothing() *AccessTokenUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the AccessTokenCreate.OnConflict
// documentation for more info.
func (u *AccessTokenUpsertOne) Update(set func(*AccessTokenUpsert)) *AccessTokenUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&AccessTokenUpsert{UpdateSet: update})
	}))
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *AccessTokenUpsertOne) SetUpdatedAt(v time.Time) *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *AccessTokenUpsertOne) UpdateUpdatedAt() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateUpdatedAt()
	})
}

// SetDeletedAt sets the "deleted_at" field.
func (u *AccessTokenUpsertOne) SetDeletedAt(v time.Time) *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetDeletedAt(v)
	})
}

// UpdateDeletedAt sets the "deleted_at" field to the value that was provided on create.
func (u *AccessTokenUpsertOne) UpdateDeletedAt() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateDeletedAt()
	})
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (u *AccessTokenUpsertOne) ClearDeletedAt() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.ClearDeletedAt()
	})
}

// SetToken sets the "token" field.
func (u *AccessTokenUpsertOne) SetToken(v string) *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetToken(v)
	})
}

// UpdateToken sets the "token" field to the value that was provided on create.
func (u *AccessTokenUpsertOne) UpdateToken() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateToken()
	})
}

// SetUsage sets the "usage" field.
func (u *AccessTokenUpsertOne) SetUsage(v string) *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetUsage(v)
	})
}

// UpdateUsage sets the "usage" field to the value that was provided on create.
func (u *AccessTokenUpsertOne) UpdateUsage() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateUsage()
	})
}

// SetEmail sets the "email" field.
func (u *AccessTokenUpsertOne) SetEmail(v string) *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetEmail(v)
	})
}

// UpdateEmail sets the "email" field to the value that was provided on create.
func (u *AccessTokenUpsertOne) UpdateEmail() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateEmail()
	})
}

// SetExpiredAt sets the "expired_at" field.
func (u *AccessTokenUpsertOne) SetExpiredAt(v time.Time) *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetExpiredAt(v)
	})
}

// UpdateExpiredAt sets the "expired_at" field to the value that was provided on create.
func (u *AccessTokenUpsertOne) UpdateExpiredAt() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateExpiredAt()
	})
}

// ClearExpiredAt clears the value of the "expired_at" field.
func (u *AccessTokenUpsertOne) ClearExpiredAt() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.ClearExpiredAt()
	})
}

// SetLastUsedAt sets the "last_used_at" field.
func (u *AccessTokenUpsertOne) SetLastUsedAt(v time.Time) *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetLastUsedAt(v)
	})
}

// UpdateLastUsedAt sets the "last_used_at" field to the value that was provided on create.
func (u *AccessTokenUpsertOne) UpdateLastUsedAt() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateLastUsedAt()
	})
}

// ClearLastUsedAt clears the value of the "last_used_at" field.
func (u *AccessTokenUpsertOne) ClearLastUsedAt() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.ClearLastUsedAt()
	})
}

// SetUserInfo sets the "user_info" field.
func (u *AccessTokenUpsertOne) SetUserInfo(v schematype.UserInfo) *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetUserInfo(v)
	})
}

// UpdateUserInfo sets the "user_info" field to the value that was provided on create.
func (u *AccessTokenUpsertOne) UpdateUserInfo() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateUserInfo()
	})
}

// ClearUserInfo clears the value of the "user_info" field.
func (u *AccessTokenUpsertOne) ClearUserInfo() *AccessTokenUpsertOne {
	return u.Update(func(s *AccessTokenUpsert) {
		s.ClearUserInfo()
	})
}

// Exec executes the query.
func (u *AccessTokenUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for AccessTokenCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *AccessTokenUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *AccessTokenUpsertOne) ID(ctx context.Context) (id int, err error) {
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *AccessTokenUpsertOne) IDX(ctx context.Context) int {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// AccessTokenCreateBulk is the builder for creating many AccessToken entities in bulk.
type AccessTokenCreateBulk struct {
	config
	err      error
	builders []*AccessTokenCreate
	conflict []sql.ConflictOption
}

// Save creates the AccessToken entities in the database.
func (atcb *AccessTokenCreateBulk) Save(ctx context.Context) ([]*AccessToken, error) {
	if atcb.err != nil {
		return nil, atcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(atcb.builders))
	nodes := make([]*AccessToken, len(atcb.builders))
	mutators := make([]Mutator, len(atcb.builders))
	for i := range atcb.builders {
		func(i int, root context.Context) {
			builder := atcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*AccessTokenMutation)
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
					_, err = mutators[i+1].Mutate(root, atcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = atcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, atcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, atcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (atcb *AccessTokenCreateBulk) SaveX(ctx context.Context) []*AccessToken {
	v, err := atcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (atcb *AccessTokenCreateBulk) Exec(ctx context.Context) error {
	_, err := atcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (atcb *AccessTokenCreateBulk) ExecX(ctx context.Context) {
	if err := atcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.AccessToken.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.AccessTokenUpsert) {
//			SetCreatedAt(v+v).
//		}).
//		Exec(ctx)
func (atcb *AccessTokenCreateBulk) OnConflict(opts ...sql.ConflictOption) *AccessTokenUpsertBulk {
	atcb.conflict = opts
	return &AccessTokenUpsertBulk{
		create: atcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.AccessToken.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (atcb *AccessTokenCreateBulk) OnConflictColumns(columns ...string) *AccessTokenUpsertBulk {
	atcb.conflict = append(atcb.conflict, sql.ConflictColumns(columns...))
	return &AccessTokenUpsertBulk{
		create: atcb,
	}
}

// AccessTokenUpsertBulk is the builder for "upsert"-ing
// a bulk of AccessToken nodes.
type AccessTokenUpsertBulk struct {
	create *AccessTokenCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.AccessToken.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//		).
//		Exec(ctx)
func (u *AccessTokenUpsertBulk) UpdateNewValues() *AccessTokenUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.CreatedAt(); exists {
				s.SetIgnore(accesstoken.FieldCreatedAt)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.AccessToken.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *AccessTokenUpsertBulk) Ignore() *AccessTokenUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *AccessTokenUpsertBulk) DoNothing() *AccessTokenUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the AccessTokenCreateBulk.OnConflict
// documentation for more info.
func (u *AccessTokenUpsertBulk) Update(set func(*AccessTokenUpsert)) *AccessTokenUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&AccessTokenUpsert{UpdateSet: update})
	}))
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *AccessTokenUpsertBulk) SetUpdatedAt(v time.Time) *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *AccessTokenUpsertBulk) UpdateUpdatedAt() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateUpdatedAt()
	})
}

// SetDeletedAt sets the "deleted_at" field.
func (u *AccessTokenUpsertBulk) SetDeletedAt(v time.Time) *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetDeletedAt(v)
	})
}

// UpdateDeletedAt sets the "deleted_at" field to the value that was provided on create.
func (u *AccessTokenUpsertBulk) UpdateDeletedAt() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateDeletedAt()
	})
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (u *AccessTokenUpsertBulk) ClearDeletedAt() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.ClearDeletedAt()
	})
}

// SetToken sets the "token" field.
func (u *AccessTokenUpsertBulk) SetToken(v string) *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetToken(v)
	})
}

// UpdateToken sets the "token" field to the value that was provided on create.
func (u *AccessTokenUpsertBulk) UpdateToken() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateToken()
	})
}

// SetUsage sets the "usage" field.
func (u *AccessTokenUpsertBulk) SetUsage(v string) *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetUsage(v)
	})
}

// UpdateUsage sets the "usage" field to the value that was provided on create.
func (u *AccessTokenUpsertBulk) UpdateUsage() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateUsage()
	})
}

// SetEmail sets the "email" field.
func (u *AccessTokenUpsertBulk) SetEmail(v string) *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetEmail(v)
	})
}

// UpdateEmail sets the "email" field to the value that was provided on create.
func (u *AccessTokenUpsertBulk) UpdateEmail() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateEmail()
	})
}

// SetExpiredAt sets the "expired_at" field.
func (u *AccessTokenUpsertBulk) SetExpiredAt(v time.Time) *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetExpiredAt(v)
	})
}

// UpdateExpiredAt sets the "expired_at" field to the value that was provided on create.
func (u *AccessTokenUpsertBulk) UpdateExpiredAt() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateExpiredAt()
	})
}

// ClearExpiredAt clears the value of the "expired_at" field.
func (u *AccessTokenUpsertBulk) ClearExpiredAt() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.ClearExpiredAt()
	})
}

// SetLastUsedAt sets the "last_used_at" field.
func (u *AccessTokenUpsertBulk) SetLastUsedAt(v time.Time) *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetLastUsedAt(v)
	})
}

// UpdateLastUsedAt sets the "last_used_at" field to the value that was provided on create.
func (u *AccessTokenUpsertBulk) UpdateLastUsedAt() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateLastUsedAt()
	})
}

// ClearLastUsedAt clears the value of the "last_used_at" field.
func (u *AccessTokenUpsertBulk) ClearLastUsedAt() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.ClearLastUsedAt()
	})
}

// SetUserInfo sets the "user_info" field.
func (u *AccessTokenUpsertBulk) SetUserInfo(v schematype.UserInfo) *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.SetUserInfo(v)
	})
}

// UpdateUserInfo sets the "user_info" field to the value that was provided on create.
func (u *AccessTokenUpsertBulk) UpdateUserInfo() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.UpdateUserInfo()
	})
}

// ClearUserInfo clears the value of the "user_info" field.
func (u *AccessTokenUpsertBulk) ClearUserInfo() *AccessTokenUpsertBulk {
	return u.Update(func(s *AccessTokenUpsert) {
		s.ClearUserInfo()
	})
}

// Exec executes the query.
func (u *AccessTokenUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the AccessTokenCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for AccessTokenCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *AccessTokenUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
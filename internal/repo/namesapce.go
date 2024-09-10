package repo

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/ent/favorite"
	"github.com/duc-cnzj/mars/v5/internal/ent/member"
	"github.com/duc-cnzj/mars/v5/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v5/internal/ent/project"
	"github.com/duc-cnzj/mars/v5/internal/filters"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/mars"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/duc-cnzj/mars/v5/internal/util/serialize"
	"github.com/samber/lo"
)

type Favorite struct {
	ID          int
	NamespaceID int
	Email       string
}

type Member struct {
	ID          int
	NamespaceID int
	Email       string
}

// Namespace is the model entity for the Namespace schema.
type Namespace struct {
	ID               int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	Name             string
	ImagePullSecrets []string
	Description      string
	Private          bool
	CreatorEmail     string

	Projects  []*Project
	Favorites []*Favorite
	Members   []*Member
}

func (ns *Namespace) GetImagePullSecrets() []*types.ImagePullSecret {
	var secrets = make([]*types.ImagePullSecret, 0)
	for _, s := range ns.ImagePullSecrets {
		secrets = append(secrets, &types.ImagePullSecret{Name: s})
	}
	return secrets
}

func ToNamespace(namespace *ent.Namespace) *Namespace {
	if namespace == nil {
		return nil
	}
	cemail := namespace.CreatorEmail
	if cemail == superAdminEmail {
		cemail = "超级管理员"
	}

	return &Namespace{
		ID:               namespace.ID,
		CreatedAt:        namespace.CreatedAt,
		UpdatedAt:        namespace.UpdatedAt,
		DeletedAt:        namespace.DeletedAt,
		Name:             namespace.Name,
		ImagePullSecrets: namespace.ImagePullSecrets,
		Description:      namespace.Description,
		Private:          namespace.Private,
		CreatorEmail:     cemail,
		Projects:         serialize.Serialize(namespace.Edges.Projects, ToProject),
		Favorites:        serialize.Serialize(namespace.Edges.Favorites, ToFavorite),
		Members:          serialize.Serialize(namespace.Edges.Members, ToMember),
	}
}

func ToMember(v *ent.Member) *Member {
	if v == nil {
		return nil
	}
	return &Member{
		ID:          v.ID,
		NamespaceID: v.NamespaceID,
		Email:       v.Email,
	}
}

func ToFavorite(v *ent.Favorite) *Favorite {
	if v == nil {
		return nil
	}
	return &Favorite{
		ID:          v.ID,
		NamespaceID: v.NamespaceID,
		Email:       v.Email,
	}
}

type NamespaceRepo interface {
	List(ctx context.Context, input *ListNamespaceInput) ([]*Namespace, *pagination.Pagination, error)
	Create(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error)
	Show(ctx context.Context, id int) (*Namespace, error)
	Update(ctx context.Context, input *UpdateNamespaceInput) (*Namespace, error)
	Delete(ctx context.Context, id int) error
	GetMarsNamespace(name string) string
	FindByName(ctx context.Context, name string) (*Namespace, error)
	Favorite(ctx context.Context, input *FavoriteNamespaceInput) error
	SyncMembers(ctx context.Context, namespaceID int, memberEmails []string) (*Namespace, error)
	UpdatePrivate(ctx context.Context, namespaceID int, private bool) (*Namespace, error)
	IsOwner(ctx context.Context, namespaceID int, user *auth.UserInfo) (bool, error)
	CanAccess(ctx context.Context, namespaceID int, user *auth.UserInfo) bool
	Transfer(ctx context.Context, id int, email string) (*Namespace, error)
}

var _ NamespaceRepo = (*namespaceRepo)(nil)

type namespaceRepo struct {
	logger   mlog.Logger
	data     data.Data
	NsPrefix string
}

func (repo *namespaceRepo) Transfer(ctx context.Context, id int, email string) (*Namespace, error) {
	ns, err := repo.data.DB().Namespace.Get(ctx, id)
	if err != nil {
		return nil, ToError(404, err)
	}
	if ns.CreatorEmail != email {
		ns, err = ns.Update().SetCreatorEmail(email).Save(ctx)
		if err != nil {
			return nil, err
		}
	}
	return ToNamespace(ns), nil
}

func (repo *namespaceRepo) CanAccess(ctx context.Context, namespaceID int, user *auth.UserInfo) bool {
	if user.IsAdmin() {
		return true
	}
	show, err := repo.data.DB().Namespace.Get(ctx, namespaceID)
	if err != nil {
		repo.logger.ErrorCtx(ctx, err, "namespaceRepo.CanAccess")
		return false
	}
	if !show.Private {
		return true
	}
	if show.CreatorEmail == user.Email {
		return true
	}
	for _, m := range show.QueryMembers().AllX(ctx) {
		if m.Email == user.Email {
			return true
		}
	}
	return false
}

func (repo *namespaceRepo) IsOwner(ctx context.Context, namespaceID int, user *auth.UserInfo) (bool, error) {
	show, err := repo.data.DB().Namespace.Get(ctx, namespaceID)
	if err != nil {
		return false, err
	}
	if user.IsAdmin() || show.CreatorEmail == user.Email {
		return true, nil
	}
	return false, nil
}

func (repo *namespaceRepo) SyncMembers(ctx context.Context, namespaceID int, memberEmails []string) (*Namespace, error) {
	if err := repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		var (
			mc  []*ent.MemberCreate
			err error
		)
		get, err := tx.Namespace.Query().WithMembers().Where(namespace.ID(namespaceID)).First(ctx)
		if err != nil {
			return err
		}
		del, add := lo.Difference(serialize.Serialize(get.Edges.Members, func(v *ent.Member) string { return v.Email }), memberEmails)
		for _, addEmail := range add {
			mc = append(mc, tx.Member.Create().SetEmail(addEmail).SetNamespaceID(namespaceID))
		}
		if _, err = tx.Member.CreateBulk(mc...).Save(ctx); err != nil {
			return err
		}
		ids, err := tx.Member.Query().Where(member.EmailIn(del...)).IDs(ctx)
		if err != nil {
			return err
		}

		if _, err := get.Update().RemoveMemberIDs(ids...).Save(ctx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return repo.Show(ctx, namespaceID)
}

func (repo *namespaceRepo) UpdatePrivate(ctx context.Context, namespaceID int, private bool) (*Namespace, error) {
	if err := repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		get, err := tx.Namespace.Get(ctx, namespaceID)
		if err != nil {
			return err
		}
		up := get.Update().
			SetPrivate(private)
		if !private {
			up.ClearMembers()
			if _, err = tx.Member.Delete().Where(member.NamespaceID(namespaceID)).Exec(ctx); err != nil {
				return err
			}
		}
		_, err = up.Save(ctx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return repo.Show(ctx, namespaceID)
}

type ListNamespaceInput struct {
	Favorite bool
	Email    string
	Page     int32
	PageSize int32
	Name     *string
	IsAdmin  bool
}

func NewNamespaceRepo(logger mlog.Logger, data data.Data) NamespaceRepo {
	return &namespaceRepo{
		logger:   logger.WithModule("repo/namespace"),
		data:     data,
		NsPrefix: data.Config().NsPrefix,
	}
}

func (repo *namespaceRepo) List(ctx context.Context, input *ListNamespaceInput) ([]*Namespace, *pagination.Pagination, error) {
	query := repo.data.DB().Namespace.Query().
		Where(
			filters.IfNameLike(lo.FromPtr(input.Name)),
		)
	if !input.IsAdmin {
		query = query.Where(
			namespace.Or(
				namespace.And(
					namespace.HasMembersWith(member.Email(input.Email)),
					namespace.Private(true),
				),
				namespace.Private(false),
				namespace.CreatorEmail(input.Email),
			),
		)
	}

	if input.Favorite {
		query = query.Where(
			namespace.HasFavoritesWith(favorite.Email(input.Email)),
		)
	}

	all, err := query.Clone().
		Select(
			namespace.FieldID,
			namespace.FieldName,
			namespace.FieldDescription,
			namespace.FieldCreatedAt,
			namespace.FieldUpdatedAt,
			namespace.FieldCreatorEmail,
			namespace.FieldPrivate,
		).
		WithFavorites(func(query *ent.FavoriteQuery) {
			query.Where(favorite.Email(input.Email))
		}).
		WithMembers(func(query *ent.MemberQuery) {
			query.Select(member.FieldID, member.FieldEmail)
		}).
		WithProjects(
			func(query *ent.ProjectQuery) {
				query.Select(
					project.FieldID,
					project.FieldName,
					project.FieldDeployStatus,
					project.FieldNamespaceID,
					project.FieldCreatedAt,
					project.FieldUpdatedAt,
				)
			},
		).
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	count := query.Clone().CountX(ctx)
	return serialize.Serialize(all, ToNamespace), pagination.NewPagination(input.Page, input.PageSize, count), nil
}

type CreateNamespaceInput struct {
	Name             string
	ImagePullSecrets []string
	Description      string
	CreatorEmail     string
}

func (repo *namespaceRepo) Create(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error) {
	save, err := repo.data.DB().Namespace.
		Create().
		SetName(mars.GetMarsNamespace(input.Name, repo.NsPrefix)).
		SetImagePullSecrets(input.ImagePullSecrets).
		SetCreatorEmail(input.CreatorEmail).
		SetDescription(input.Description).
		Save(ctx)
	return ToNamespace(save), err
}

func (repo *namespaceRepo) Show(ctx context.Context, id int) (*Namespace, error) {
	first, err := repo.data.DB().Namespace.Query().
		WithProjects(func(query *ent.ProjectQuery) {
			query.Select(
				project.FieldID,
				project.FieldName,
				project.FieldNamespaceID,
			)
		}).
		WithMembers().
		Where(namespace.ID(id)).
		First(ctx)
	return ToNamespace(first), ToError(404, err)
}

type UpdateNamespaceInput struct {
	ID          int
	Description string
}

func (repo *namespaceRepo) Update(ctx context.Context, input *UpdateNamespaceInput) (*Namespace, error) {
	get, err := repo.data.DB().Namespace.Get(ctx, input.ID)
	if err != nil {
		return nil, ToError(404, err)
	}
	save, err := get.Update().SetDescription(input.Description).Save(ctx)
	return ToNamespace(save), err
}

func (repo *namespaceRepo) GetMarsNamespace(name string) string {
	return mars.GetMarsNamespace(name, repo.NsPrefix)
}

func (repo *namespaceRepo) FindByName(ctx context.Context, name string) (*Namespace, error) {
	first, err := repo.data.DB().Namespace.Query().Where(namespace.Name(mars.GetMarsNamespace(name, repo.NsPrefix))).First(ctx)
	return ToNamespace(first), ToError(404, err)
}

func (repo *namespaceRepo) Delete(ctx context.Context, id int) error {
	first, err := repo.data.DB().Namespace.Query().WithProjects().Where(namespace.ID(id)).First(ctx)
	if err != nil {
		return err
	}
	return repo.data.WithTx(ctx, func(tx *ent.Tx) error {
		if len(first.Edges.Projects) > 0 {
			if _, err := tx.Project.
				Delete().
				Where(project.HasNamespaceWith(namespace.ID(id))).
				Exec(ctx); err != nil {
				return err
			}
		}
		return tx.Namespace.DeleteOneID(id).Exec(ctx)
	})
}

type FavoriteNamespaceInput struct {
	NamespaceID int
	UserEmail   string
	Favorite    bool
}

func (repo *namespaceRepo) Favorite(ctx context.Context, input *FavoriteNamespaceInput) error {
	if !input.Favorite {
		_, err := repo.data.DB().Favorite.Delete().Where(favorite.NamespaceID(input.NamespaceID), favorite.Email(input.UserEmail)).Exec(ctx)
		return err
	}

	if exist, _ := repo.data.DB().Favorite.Query().Where(favorite.NamespaceID(input.NamespaceID), favorite.Email(input.UserEmail)).Exist(ctx); exist {
		return nil
	}
	return repo.data.DB().Favorite.Create().SetNamespaceID(input.NamespaceID).SetEmail(input.UserEmail).Exec(ctx)
}

package repo

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/favorite"
	"github.com/duc-cnzj/mars/v4/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v4/internal/ent/project"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/mars"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
)

type Favorite struct {
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

	Projects  []*Project
	Favorites []*Favorite
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

	return &Namespace{
		ID:               namespace.ID,
		CreatedAt:        namespace.CreatedAt,
		UpdatedAt:        namespace.UpdatedAt,
		DeletedAt:        namespace.DeletedAt,
		Name:             namespace.Name,
		ImagePullSecrets: namespace.ImagePullSecrets,
		Description:      namespace.Description,
		Projects:         serialize.Serialize(namespace.Edges.Projects, ToProject),
		Favorites:        serialize.Serialize(namespace.Edges.Favorites, ToFavorite),
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
	Delete(ctx context.Context, id int) error
	GetMarsNamespace(name string) string
	FindByName(ctx context.Context, name string) (*Namespace, error)
	All(ctx context.Context, input *AllNamespaceInput) ([]*Namespace, error)
	Show(ctx context.Context, id int) (*Namespace, error)
	Create(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error)
	Favorite(ctx context.Context, input *FavoriteNamespaceInput) error
	Update(ctx context.Context, input *UpdateNamespaceInput) (*Namespace, error)
}

var _ NamespaceRepo = (*namespaceRepo)(nil)

type namespaceRepo struct {
	logger   mlog.Logger
	data     data.Data
	NsPrefix string
}

func NewNamespaceRepo(logger mlog.Logger, data data.Data) NamespaceRepo {
	return &namespaceRepo{
		logger:   logger.WithModule("repo/namespace"),
		data:     data,
		NsPrefix: data.Config().NsPrefix,
	}
}

type AllNamespaceInput struct {
	Favorite bool
	Email    string
}

func (repo *namespaceRepo) All(ctx context.Context, input *AllNamespaceInput) ([]*Namespace, error) {
	query := repo.data.DB().Namespace.Query().
		WithFavorites(func(query *ent.FavoriteQuery) {
			query.Where(favorite.Email(input.Email))
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
		)
	if input.Favorite {
		query = query.Where(
			namespace.HasFavoritesWith(favorite.Email(input.Email)),
		)
	}

	all, err := query.Select(
		namespace.FieldID,
		namespace.FieldName,
		namespace.FieldDescription,
		namespace.FieldCreatedAt,
		namespace.FieldUpdatedAt,
	).All(ctx)
	return serialize.Serialize(all, ToNamespace), err
}

type CreateNamespaceInput struct {
	Name             string
	ImagePullSecrets []string
	Description      string
}

func (repo *namespaceRepo) Create(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error) {
	save, err := repo.data.DB().Namespace.
		Create().
		SetName(mars.GetMarsNamespace(input.Name, repo.NsPrefix)).
		SetImagePullSecrets(input.ImagePullSecrets).
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
				Where(
					project.IDIn(
						serialize.Serialize(first.Edges.Projects, func(v *ent.Project) int {
							return v.ID
						})...),
				).Exec(ctx); err != nil {
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

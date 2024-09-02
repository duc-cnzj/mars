package repo

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent/favorite"
	"github.com/duc-cnzj/mars/v5/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v5/internal/ent/project"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNamespaceRepo_List_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ctx := context.TODO()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{},
		DB:  db,
	}))

	// seed data
	ns1 := db.Namespace.Create().SetName("test").SaveX(ctx)
	db.Favorite.Create().SetNamespaceID(ns1.ID).SetEmail("test@example.com").Save(context.TODO())
	db.Namespace.Create().SetName("tes2").SaveX(ctx)
	db.Namespace.Create().SetName("tes3").SaveX(ctx)

	input := &ListNamespaceInput{
		Favorite: true,
		Email:    "test@example.com",
		Page:     1,
		PageSize: 10,
		Name:     nil,
	}

	res, pag, err := repo.List(ctx, input)
	assert.NotNil(t, res)
	assert.NotNil(t, pag)
	assert.Nil(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, int32(1), pag.Count)
	assert.Equal(t, int32(1), pag.Page)
	assert.Equal(t, int32(10), pag.PageSize)
	assert.Equal(t, ns1.ID, res[0].ID)

	input = &ListNamespaceInput{
		Favorite: false,
		Page:     1,
		PageSize: 10,
		Name:     lo.ToPtr("es3"),
	}

	res, pag, _ = repo.List(ctx, input)
	assert.Len(t, res, 1)
	assert.Equal(t, int32(1), pag.Count)
}

func Test_namespaceRepo_Create(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: db,
	}))

	create, err := repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc-aaa", create.Name)
}

func Test_namespaceRepo_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc",
		},
		DB: db,
	}))

	create, err := repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
	})
	createProject(db, create.ID)
	assert.Nil(t, err)
	show, err := repo.Show(context.TODO(), create.ID)
	assert.Nil(t, err)
	assert.Len(t, show.Projects, 1)
}

func Test_namespaceRepo_Update(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc",
		},
		DB: db,
	}))

	create, err := repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
	})
	assert.Nil(t, err)
	update, err := repo.Update(context.TODO(), &UpdateNamespaceInput{
		ID:          create.ID,
		Description: "aaaaaa",
	})
	assert.Nil(t, err)
	assert.Equal(t, "aaaaaa", update.Description)
}

func Test_namespaceRepo_GetMarsNamespace(t *testing.T) {
	repo := NewNamespaceRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
	}))
	namespace := repo.GetMarsNamespace("a")
	assert.Equal(t, "abc-a", namespace)
	marsNamespace := repo.GetMarsNamespace("abc-a")
	assert.Equal(t, "abc-a", marsNamespace)
}

func Test_namespaceRepo_Delete(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc",
		},
		DB: db,
	}))

	create, err := repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
	})
	assert.Nil(t, err)

	create2, err := repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
	})
	assert.Nil(t, err)

	createProject(db, create.ID)
	createProject(db, create.ID)
	createProject(db, create.ID)
	createProject(db, create.ID)
	p2 := createProject(db, create2.ID)

	err = repo.Delete(context.TODO(), create.ID)
	assert.Nil(t, err)
	softDelete := mixin.SkipSoftDelete(context.TODO())
	x := db.Project.Query().Where(project.NamespaceID(create.ID)).AllX(softDelete)

	for _, p := range x {
		assert.NotZero(t, p.DeletedAt)
	}
	first, _ := db.Project.Query().Where(project.ID(p2.ID)).First(context.TODO())
	assert.Zero(t, first.DeletedAt)

	n, err := db.Namespace.Query().Where(namespace.ID(create.ID)).First(softDelete)
	assert.Nil(t, err)
	assert.NotZero(t, n.DeletedAt)
}

func Test_namespaceRepo_Favorite_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{},
		DB:  db,
	}))

	ns := db.Namespace.Create().SetName("test").SaveX(context.TODO())
	input := &FavoriteNamespaceInput{
		NamespaceID: ns.ID,
		UserEmail:   "test@example.com",
		Favorite:    true,
	}

	err := repo.Favorite(context.TODO(), input)
	assert.Nil(t, err)

	fav := db.Favorite.Query().Where(favorite.NamespaceID(ns.ID), favorite.Email("test@example.com")).OnlyX(context.TODO())
	assert.NotNil(t, fav)
}

func Test_namespaceRepo_Favorite_AlreadyExists(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{},
		DB:  db,
	}))

	ns := db.Namespace.Create().SetName("test").SaveX(context.TODO())
	db.Favorite.Create().SetNamespaceID(ns.ID).SetEmail("test@example.com").Save(context.TODO())
	input := &FavoriteNamespaceInput{
		NamespaceID: ns.ID,
		UserEmail:   "test@example.com",
		Favorite:    true,
	}

	err := repo.Favorite(context.TODO(), input)
	assert.Nil(t, err)
}

func Test_namespaceRepo_Favorite_Unfavorite(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{},
		DB:  db,
	}))

	ns := db.Namespace.Create().SetName("test").SaveX(context.TODO())
	db.Favorite.Create().SetNamespaceID(ns.ID).SetEmail("test@example.com").Save(context.TODO())
	input := &FavoriteNamespaceInput{
		NamespaceID: ns.ID,
		UserEmail:   "test@example.com",
		Favorite:    false,
	}

	err := repo.Favorite(context.TODO(), input)
	assert.Nil(t, err)

	_, err = db.Favorite.Query().Where(favorite.NamespaceID(ns.ID), favorite.Email("test@example.com")).Only(context.TODO())
	assert.Error(t, err)
}

func Test_namespaceRepo_FindByName(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: db,
	}))

	repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
	})
	name, _ := repo.FindByName(context.TODO(), "aaa")
	assert.NotNil(t, name)
	name, _ = repo.FindByName(context.TODO(), "abc-aaa")
	assert.NotNil(t, name)
}

func TestNamespace_GetImagePullSecrets(t *testing.T) {
	ns := &Namespace{}
	assert.NotNil(t, ns.GetImagePullSecrets())
	ns = &Namespace{
		ImagePullSecrets: []string{"a", "b"},
	}
	assert.Len(t, ns.GetImagePullSecrets(), 2)
}

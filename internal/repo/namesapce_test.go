package repo

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/ent/favorite"
	"github.com/duc-cnzj/mars/v5/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v5/internal/ent/project"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
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
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{},
		DB:  db,
	}))

	// seed data
	ns1 := db.Namespace.Create().SetCreatorEmail("a").SetName("test").SaveX(ctx)
	db.Favorite.Create().SetNamespaceID(ns1.ID).SetEmail("test@example.com").Save(context.TODO())
	db.Namespace.Create().SetCreatorEmail("a").SetName("tes2").SaveX(ctx)
	db.Namespace.Create().SetCreatorEmail("a").SetName("tes3").SaveX(ctx)
	pri1 := db.Namespace.Create().SetCreatorEmail("a").SetName("pri1").SetPrivate(true).SaveX(ctx)
	db.Member.Create().SetEmail("user@mars.com").SetNamespaceID(pri1.ID).SaveX(ctx)
	db.Namespace.Create().SetCreatorEmail("a").SetName("pri2").SetPrivate(true).SaveX(ctx)

	input := &ListNamespaceInput{
		Favorite: true,
		Email:    "test@example.com",
		Page:     1,
		PageSize: 10,
		Name:     nil,
		IsAdmin:  true,
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

	input = &ListNamespaceInput{
		Page:     1,
		PageSize: 10,
		IsAdmin:  false,
		Email:    "",
	}

	res, pag, _ = repo.List(ctx, input)
	assert.Len(t, res, 3)
	assert.Equal(t, int32(3), pag.Count)

	input = &ListNamespaceInput{
		Page:     1,
		PageSize: 10,
		IsAdmin:  false,
		Email:    "user@mars.com",
	}

	res, pag, _ = repo.List(ctx, input)
	assert.Len(t, res, 4)
	assert.Equal(t, int32(4), pag.Count)

	input = &ListNamespaceInput{
		Page:     1,
		PageSize: 10,
		IsAdmin:  true,
	}

	res, pag, _ = repo.List(ctx, input)
	assert.Len(t, res, 5)
	assert.Equal(t, int32(5), pag.Count)

	input = &ListNamespaceInput{
		Email:    "a",
		Page:     1,
		PageSize: 10,
		IsAdmin:  false,
	}

	res, pag, _ = repo.List(ctx, input)
	assert.Len(t, res, 5)
	assert.Equal(t, int32(5), pag.Count)
}

func Test_namespaceRepo_Create(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: db,
	}))

	create, err := repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc-aaa", create.Name)
	assert.Equal(t, "aa", create.CreatorEmail)
}

func Test_namespaceRepo_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc",
		},
		DB: db,
	}))

	create, err := repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
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
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc",
		},
		DB: db,
	}))

	create, err := repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
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
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
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
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc",
		},
		DB: db,
	}))

	create, err := repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
	})
	assert.Nil(t, err)

	create2, err := repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
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
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{},
		DB:  db,
	}))

	ns := db.Namespace.Create().SetCreatorEmail("a").SetName("test").SaveX(context.TODO())
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
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{},
		DB:  db,
	}))

	ns := db.Namespace.Create().SetCreatorEmail("a").SetName("test").SaveX(context.TODO())
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
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{},
		DB:  db,
	}))

	ns := db.Namespace.Create().SetCreatorEmail("a").SetName("test").SaveX(context.TODO())
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
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: db,
	}))

	repo.Create(context.TODO(), &CreateNamespaceInput{
		Name:             "aaa",
		ImagePullSecrets: []string{"a", "b"},
		Description:      "desc",
		CreatorEmail:     "aa",
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

func Test_namespaceRepo_SyncMembers(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: db,
	}))
	ns := createNamespace(db)
	ns.Update().SetPrivate(true).SaveX(context.TODO())

	res, err := repo.SyncMembers(context.TODO(), ns.ID, []string{"a", "b"})
	assert.Nil(t, err)
	assert.Len(t, res.Members, 2)

	res, err = repo.SyncMembers(context.TODO(), ns.ID, []string{"c"})
	assert.Nil(t, err)
	assert.Len(t, res.Members, 1)
	assert.Equal(t, "c", res.Members[0].Email)

	res, err = repo.SyncMembers(context.TODO(), ns.ID, []string{})
	assert.Nil(t, err)
	assert.Len(t, res.Members, 0)
}

func Test_namespaceRepo_UpdatePrivate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: db,
	}))
	ns := createNamespace(db)
	ns.Update().SetPrivate(true).SaveX(context.TODO())

	res, _ := repo.SyncMembers(context.TODO(), ns.ID, []string{"a", "b"})
	assert.Len(t, res.Members, 2)

	private, err := repo.UpdatePrivate(context.TODO(), ns.ID, false)
	assert.Nil(t, err)
	assert.False(t, private.Private)

	x := db.Member.Query().AllX(context.TODO())
	assert.Len(t, x, 0)

	private, err = repo.UpdatePrivate(context.TODO(), ns.ID, true)
	assert.Nil(t, err)
	assert.True(t, private.Private)
}

func Test_namespaceRepo_IsOwner(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: db,
	}))
	ns := createNamespace(db)
	ns.Update().SetPrivate(true).SetCreatorEmail("bbb").SaveX(context.TODO())

	owner, err := repo.IsOwner(context.TODO(), ns.ID, &auth.UserInfo{Email: "aaa"})
	assert.Nil(t, err)
	assert.False(t, owner)

	owner, err = repo.IsOwner(context.TODO(), ns.ID, &auth.UserInfo{Email: "bbb"})
	assert.Nil(t, err)
	assert.True(t, owner)

	owner, err = repo.IsOwner(context.TODO(), ns.ID, &auth.UserInfo{Email: "cccc", Roles: []string{schematype.MarsAdmin}})
	assert.Nil(t, err)
	assert.True(t, owner)
}

func TestToNamespace(t *testing.T) {
	toNamespace := ToNamespace(&ent.Namespace{
		CreatorEmail: superAdminEmail,
	})

	assert.Equal(t, "超级管理员", toNamespace.CreatorEmail)
	toNamespace = ToNamespace(&ent.Namespace{
		CreatorEmail: "abc@qq.com",
	})

	assert.Equal(t, "abc@qq.com", toNamespace.CreatorEmail)
}

func Test_namespaceRepo_Transfer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewNamespaceRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{
		Cfg: &config.Config{
			NsPrefix: "abc-",
		},
		DB: db,
	}))
	ns := createNamespace(db)
	ns.Update().SetPrivate(true).SetCreatorEmail("bbb").SaveX(context.TODO())

	res, err := repo.Transfer(context.TODO(), ns.ID, "aaa")
	assert.Nil(t, err)
	assert.Equal(t, "aaa", res.CreatorEmail)

	_, err = repo.Transfer(context.TODO(), 9999999, "aaa")
	assert.Error(t, err)
}

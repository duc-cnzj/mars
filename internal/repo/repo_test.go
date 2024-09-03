package repo

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/api/v5/mars"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/status"
)

func TestRepoImpl_All(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))
	res, err := repo.All(context.TODO(), &AllRepoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestRepoImpl_List(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))
	res, pag, err := repo.List(context.TODO(), &ListRepoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, pag)
}

func TestRepoImpl_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))
	_, err := repo.Show(context.TODO(), 1)
	s, _ := status.FromError(err)

	assert.Equal(t, "NotFound", s.Code().String())
}

func TestRepoImpl_Create(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), mockGitRepo)
	mockGitRepo.EXPECT().GetByProjectID(gomock.Any(), 100).Return(&GitProject{}, nil)
	res, err := repo.Create(context.TODO(), &CreateRepoInput{
		Name:         "app",
		Enabled:      true,
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(100)),
		MarsConfig:   &mars.Config{ConfigField: "config"},
		Description:  "desc",
	})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.GitProjectID)

	_, err = repo.Create(context.TODO(), &CreateRepoInput{
		Name: "app",
	})

	s, _ := status.FromError(err)
	assert.Equal(t, "repo 名称已经存在", s.Message())
}

func TestRepoImpl_Update(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), mockGitRepo)

	create, err := repo.Create(context.TODO(), &CreateRepoInput{
		Name: "app",
	})
	assert.Nil(t, err)

	mockGitRepo.EXPECT().GetByProjectID(gomock.Any(), 100).Return(&GitProject{
		DefaultBranch: "dev",
		Name:          "a",
	}, nil)
	res, err := repo.Update(context.TODO(), &UpdateRepoInput{
		ID:           int32(create.ID),
		Name:         "abc",
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(100)),
		MarsConfig:   &mars.Config{ConfigField: "config"},
		Description:  "dex",
	})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "abc", res.Name)
	assert.Equal(t, "dex", res.Description)
	assert.Equal(t, true, res.NeedGitRepo)
	assert.Equal(t, "config", res.MarsConfig.ConfigField)
	assert.Equal(t, "dev", res.DefaultBranch)
	assert.Equal(t, "a", res.GitProjectName)
	assert.Equal(t, int32(100), res.GitProjectID)

	db.Repo.Create().SetName("uuu").SaveX(context.TODO())
	_, err = repo.Update(context.TODO(), &UpdateRepoInput{
		ID:           int32(create.ID),
		Name:         "uuu",
		NeedGitRepo:  true,
		GitProjectID: lo.ToPtr(int32(100)),
		MarsConfig:   &mars.Config{ConfigField: "config"},
		Description:  "dex",
	})

	s, _ := status.FromError(err)
	assert.Equal(t, "repo 名称已经存在", s.Message())

	project := createProject(db, createNamespace(db).ID)
	project.Update().SetRepoID(create.ID).SaveX(context.Background())
	_, err = repo.Update(context.TODO(), &UpdateRepoInput{
		ID:          int32(create.ID),
		Name:        "abcd",
		NeedGitRepo: false,
	})
	s, _ = status.FromError(err)
	assert.Equal(t, "repo 下面还有 1 个项目，不能修改名称", s.Message())
}

func TestRepoImpl_ToggleEnabled(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

	create, err := repo.Create(context.TODO(), &CreateRepoInput{
		Name: "app",
	})
	assert.Nil(t, err)
	assert.False(t, create.Enabled)

	res, err := repo.ToggleEnabled(context.TODO(), create.ID, true)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.Enabled)
}

func TestRepoImpl_ToggleEnabled_WithProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

	create, err := repo.Create(context.TODO(), &CreateRepoInput{
		Name:    "app",
		Enabled: true,
	})
	assert.Nil(t, err)
	assert.True(t, create.Enabled)

	ns := createNamespace(db)
	project := createProject(db, ns.ID)
	project.Update().SetRepoID(create.ID).SaveX(context.Background())

	// Attempt to disable the repo, should fail because it has projects
	_, err = repo.ToggleEnabled(context.TODO(), create.ID, false)
	assert.NotNil(t, err)
}

func TestRepo_GetMarsConfig_WithExistingConfig(t *testing.T) {
	r := &Repo{
		MarsConfig: &mars.Config{ConfigField: "existing_config"},
	}
	cfg := r.GetMarsConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "existing_config", cfg.ConfigField)
}

func TestRepo_GetMarsConfig_WithoutExistingConfig(t *testing.T) {
	r := &Repo{}
	cfg := r.GetMarsConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "", cfg.ConfigField)
}

func TestRepoImpl_GetProjNameAndBranch_WithExistingProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(mlog.NewForConfig(nil), data.NewMockData(m), mockGitRepo)
	mockGitRepo.EXPECT().GetByProjectID(gomock.Any(), gomock.Any()).Return(&GitProject{
		DefaultBranch: "main",
		Name:          "projName",
	}, nil)
	projName, defaultBranch, err := repo.(*repoImpl).GetProjNameAndBranch(context.TODO(), 1)
	assert.Nil(t, err)
	assert.NotNil(t, projName)
	assert.NotNil(t, defaultBranch)
}

func TestRepoImpl_GetProjNameAndBranch_WithNonExistingProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(mlog.NewForConfig(nil), data.NewMockData(m), mockGitRepo)
	mockGitRepo.EXPECT().GetByProjectID(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)

	projName, defaultBranch, err := repo.(*repoImpl).GetProjNameAndBranch(context.TODO(), 1)

	assert.NotNil(t, err)
	assert.Nil(t, projName)
	assert.Nil(t, defaultBranch)
}

func TestCloneRepoWithValidInput(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	db, _ := data.NewSqliteDB()
	defer db.Close()

	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), mockGitRepo)

	create, _ := repo.Create(context.TODO(), &CreateRepoInput{
		Name: "app",
	})

	_, err := repo.Clone(context.TODO(), &CloneRepoInput{
		ID:   create.ID,
		Name: "clone",
	})

	assert.Nil(t, err)
}

func TestCloneRepoWithExistingName(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	db, _ := data.NewSqliteDB()
	defer db.Close()

	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), mockGitRepo)

	create, _ := repo.Create(context.TODO(), &CreateRepoInput{
		Name: "app",
	})

	_, err := repo.Clone(context.TODO(), &CloneRepoInput{
		ID:   create.ID,
		Name: "app",
	})

	assert.NotNil(t, err)
}

func TestRepoImpl_Delete_WithExistingRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

	create, _ := repo.Create(context.TODO(), &CreateRepoInput{
		Name: "app",
	})

	err := repo.Delete(context.TODO(), create.ID)
	assert.Nil(t, err)
}

func TestRepoImpl_Delete_WithNonExistingRepo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

	err := repo.Delete(context.TODO(), 9999)
	assert.NotNil(t, err)
}

func TestRepoImpl_Delete_WithRepoHavingProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewForConfig(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

	create, _ := repo.Create(context.TODO(), &CreateRepoInput{
		Name: "app",
	})

	ns := createNamespace(db)
	project := createProject(db, ns.ID)
	project.Update().SetRepoID(create.ID).SaveX(context.Background())

	err := repo.Delete(context.TODO(), create.ID)
	assert.NotNil(t, err)
}

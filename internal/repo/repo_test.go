package repo

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRepoImpl_All(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))
	res, err := repo.All(context.TODO(), &AllRepoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestRepoImpl_List(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))
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
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))
	_, err := repo.Show(context.TODO(), 1)
	s, _ := status.FromError(err)

	assert.Equal(t, "NotFound", s.Code().String())
}

func TestRepoImpl_Create(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))
	res, err := repo.Create(context.TODO(), &CreateRepoInput{
		Name:         "app",
		Enabled:      true,
		NeedGitRepo:  false,
		GitProjectID: nil,
		MarsConfig:   &mars.Config{ConfigField: "config"},
		Description:  "desc",
	})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestRepoImpl_Update(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

	create, err := repo.Create(context.TODO(), &CreateRepoInput{
		Name: "app",
	})
	assert.Nil(t, err)
	res, err := repo.Update(context.TODO(), &UpdateRepoInput{
		ID:          int32(create.ID),
		Name:        "abc",
		NeedGitRepo: false,
		MarsConfig:  &mars.Config{ConfigField: "config"},
		Description: "dex",
	})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "abc", res.Name)
	assert.Equal(t, "dex", res.Description)
	assert.Equal(t, false, res.NeedGitRepo)
	assert.Equal(t, "config", res.MarsConfig.ConfigField)
}

func TestRepoImpl_ToggleEnabled(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

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
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

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
	repo := NewRepo(mlog.NewLogger(nil), data.NewMockData(m), mockGitRepo)
	project := application.NewMockProject(m)
	mockGitRepo.EXPECT().GetByProjectID(gomock.Any(), gomock.Any()).Return(project, nil)
	project.EXPECT().GetDefaultBranch().Return("main")
	project.EXPECT().GetName().Return("projName")
	projName, defaultBranch, err := repo.(*repoImpl).GetProjNameAndBranch(context.TODO(), 1)
	assert.Nil(t, err)
	assert.NotNil(t, projName)
	assert.NotNil(t, defaultBranch)
}

func TestRepoImpl_GetProjNameAndBranch_WithNonExistingProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockGitRepo := NewMockGitRepo(m)
	repo := NewRepo(mlog.NewLogger(nil), data.NewMockData(m), mockGitRepo)
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
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), mockGitRepo)

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
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), mockGitRepo)

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
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

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
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

	err := repo.Delete(context.TODO(), 9999)
	assert.NotNil(t, err)
}

func TestRepoImpl_Delete_WithRepoHavingProjects(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	db, _ := data.NewSqliteDB()
	defer db.Close()
	repo := NewRepo(mlog.NewLogger(nil), data.NewDataImpl(&data.NewDataParams{DB: db}), NewMockGitRepo(m))

	create, _ := repo.Create(context.TODO(), &CreateRepoInput{
		Name: "app",
	})

	ns := createNamespace(db)
	project := createProject(db, ns.ID)
	project.Update().SetRepoID(create.ID).SaveX(context.Background())

	err := repo.Delete(context.TODO(), create.ID)
	assert.NotNil(t, err)
}

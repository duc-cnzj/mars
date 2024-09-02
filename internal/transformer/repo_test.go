package transformer_test

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v5/mars"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/transformer"
	"github.com/stretchr/testify/assert"
)

func TestFromRepo_NilInput(t *testing.T) {
	var r *repo.Repo
	result := transformer.FromRepo(r)
	assert.Nil(t, result)
}

func TestFromRepo_ValidInput(t *testing.T) {
	r := &repo.Repo{
		ID:             1,
		Name:           "testRepo",
		GitProjectID:   int32(1),
		GitProjectName: "testGitProjectName",
		Enabled:        true,
		MarsConfig:     &mars.Config{},
		NeedGitRepo:    true,
		Description:    "testDescription",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	result := transformer.FromRepo(r)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.Id)
	assert.Equal(t, "testRepo", result.Name)
	assert.Equal(t, int32(1), result.GitProjectId)
	assert.Equal(t, "testGitProjectName", result.GitProjectName)
	assert.Equal(t, true, result.Enabled)
	assert.Equal(t, &mars.Config{}, result.MarsConfig)
	assert.Equal(t, true, result.NeedGitRepo)
	assert.Equal(t, "testDescription", result.Description)
	assert.NotNil(t, result.CreatedAt)
	assert.NotNil(t, result.UpdatedAt)
	assert.Equal(t, "", result.DeletedAt)
}

func TestFromRepo_DeletedRepo(t *testing.T) {
	now := time.Now()
	r := &repo.Repo{
		ID:             1,
		Name:           "testRepo",
		GitProjectID:   int32(1),
		GitProjectName: "testGitProjectName",
		Enabled:        true,
		MarsConfig:     &mars.Config{},
		NeedGitRepo:    true,
		Description:    "testDescription",
		CreatedAt:      now,
		UpdatedAt:      now,
		DeletedAt:      &now,
	}
	result := transformer.FromRepo(r)
	assert.NotNil(t, result)
	assert.NotNil(t, result.DeletedAt)
}

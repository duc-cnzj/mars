package transformer_test

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
	"github.com/stretchr/testify/assert"
)

func TestFromProject_NilInput(t *testing.T) {
	var p *repo.Project
	result := transformer.FromProject(p)
	assert.Nil(t, result)
}

func TestFromProject_ValidInput(t *testing.T) {
	now := time.Now()
	p := &repo.Project{
		ID:             1,
		Name:           "testProject",
		GitProjectID:   1,
		GitBranch:      "testBranch",
		GitCommit:      "testCommit",
		Config:         "testConfig",
		OverrideValues: "testOverrideValues",
		DockerImage:    []string{"testDockerImage"},
		PodSelectors:   []string{"testPodSelectors"},
		NamespaceID:    1,
		Atomic:         true,
		EnvValues: []*types.KeyValue{
			{
				Key:   "k",
				Value: "v",
			},
		},
		ExtraValues: []*websocket_pb.ExtraValue{
			{
				Path:  "p",
				Value: "v",
			},
		},
		FinalExtraValues: []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue"}},
		DeployStatus:     types.Deploy_StatusDeploying,
		ConfigType:       "testConfigType",
		GitCommitWebURL:  "testGitCommitWebURL",
		GitCommitTitle:   "testGitCommitTitle",
		GitCommitAuthor:  "testGitCommitAuthor",
		GitCommitDate:    &now,
		Version:          1,
		RepoID:           1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	result := transformer.FromProject(p)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.Id)
	assert.Equal(t, "testProject", result.Name)
	assert.Equal(t, int32(1), result.GitProjectId)
	assert.Equal(t, "testBranch", result.GitBranch)
	assert.Equal(t, "testCommit", result.GitCommit)
	assert.Equal(t, "testConfig", result.Config)
	assert.Equal(t, "testOverrideValues", result.OverrideValues)
	assert.Equal(t, "testDockerImage", result.DockerImage)
	assert.Equal(t, []string{"testPodSelectors"}, result.PodSelectors)
	assert.Equal(t, int32(1), result.NamespaceId)
	assert.Equal(t, true, result.Atomic)
	assert.Equal(t, []*types.KeyValue{
		{
			Key:   "k",
			Value: "v",
		},
	}, result.EnvValues)
	assert.Equal(t, []*websocket_pb.ExtraValue{
		{
			Path:  "p",
			Value: "v",
		},
	}, result.ExtraValues)
	assert.Equal(t, "testFinalExtraValues", result.FinalExtraValues)
	assert.Equal(t, types.Deploy_StatusDeploying, result.DeployStatus)
	assert.Equal(t, date.ToHumanizeDatetimeString(&now), result.HumanizeCreatedAt)
	assert.Equal(t, date.ToHumanizeDatetimeString(&now), result.HumanizeUpdatedAt)
	assert.Equal(t, "testConfigType", result.ConfigType)
	assert.Equal(t, "testGitCommitWebURL", result.GitCommitWebUrl)
	assert.Equal(t, "testGitCommitTitle", result.GitCommitTitle)
	assert.Equal(t, "testGitCommitAuthor", result.GitCommitAuthor)
	assert.Equal(t, date.ToHumanizeDatetimeString(&now), result.GitCommitDate)
	assert.Equal(t, int32(1), result.Version)
	assert.Equal(t, int32(1), result.RepoId)
	assert.Equal(t, date.ToRFC3339DatetimeString(&now), result.CreatedAt)
	assert.Equal(t, date.ToRFC3339DatetimeString(&now), result.UpdatedAt)
	assert.Empty(t, result.DeletedAt)
}

func TestFromProject_DeletedProject(t *testing.T) {
	now := time.Now()
	p := &repo.Project{
		ID:             1,
		Name:           "testProject",
		GitProjectID:   1,
		GitBranch:      "testBranch",
		GitCommit:      "testCommit",
		Config:         "testConfig",
		OverrideValues: "testOverrideValues",
		DockerImage:    []string{"testDockerImage"},
		PodSelectors:   []string{"testPodSelectors"},
		NamespaceID:    1,
		Atomic:         true,
		EnvValues: []*types.KeyValue{
			{
				Key:   "k",
				Value: "v",
			},
		},
		ExtraValues: []*websocket_pb.ExtraValue{
			{
				Path:  "p",
				Value: "v",
			},
		},
		FinalExtraValues: []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue"}},
		DeployStatus:     types.Deploy_StatusDeploying,
		ConfigType:       "testConfigType",
		GitCommitWebURL:  "testGitCommitWebURL",
		GitCommitTitle:   "testGitCommitTitle",
		GitCommitAuthor:  "testGitCommitAuthor",
		GitCommitDate:    &now,
		Version:          1,
		RepoID:           1,
		CreatedAt:        now,
		UpdatedAt:        now,
		DeletedAt:        &now,
	}
	result := transformer.FromProject(p)
	assert.NotNil(t, result)
	assert.NotNil(t, result.DeletedAt)
}

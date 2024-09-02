package transformer_test

import (
	"strings"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v5/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/transformer"
	"github.com/duc-cnzj/mars/v5/internal/util/date"
	"github.com/stretchr/testify/assert"
)

func TestFromChangeLog_NilInput(t *testing.T) {
	var c *repo.Changelog
	result := transformer.FromChangeLog(c)
	assert.Nil(t, result)
}

func TestFromChangeLog_ValidInput(t *testing.T) {
	now := time.Now()
	c := &repo.Changelog{
		ID:            1,
		Version:       1,
		Username:      "testUser",
		Config:        "testConfig",
		ConfigChanged: true,
		ProjectID:     1,
		Project:       &repo.Project{},
		CreatedAt:     now,
		UpdatedAt:     now,
		GitBranch:     "testBranch",
		GitCommit:     "testCommit",
		DockerImage:   []string{"testImage1", "testImage2"},
		EnvValues: []*types.KeyValue{
			{
				Key:   "testKey1",
				Value: "testValue1",
			},
		},
		ExtraValues: []*websocket_pb.ExtraValue{
			{
				Path:  "a",
				Value: "v",
			},
		},
		FinalExtraValues: []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue"}},
		GitCommitWebURL:  "testUrl",
		GitCommitTitle:   "testTitle",
		GitCommitAuthor:  "testAuthor",
		GitCommitDate:    &now,
		DeletedAt:        nil,
	}
	result := transformer.FromChangeLog(c)
	assert.NotNil(t, result)
	assert.Equal(t, int32(c.ID), result.Id)
	assert.Equal(t, int32(c.Version), result.Version)
	assert.Equal(t, c.Username, result.Username)
	assert.Equal(t, c.Config, result.Config)
	assert.Equal(t, c.ConfigChanged, result.ConfigChanged)
	assert.Equal(t, int64(c.ProjectID), result.ProjectId)
	assert.Equal(t, transformer.FromProject(c.Project), result.Project)
	assert.Equal(t, date.ToHumanizeDatetimeString(&c.CreatedAt), result.Date)
	assert.Equal(t, c.GitBranch, result.GitBranch)
	assert.Equal(t, c.GitCommit, result.GitCommit)
	assert.Equal(t, strings.Join(c.DockerImage, ","), result.DockerImage)
	assert.Equal(t, c.EnvValues, result.EnvValues)
	assert.Equal(t, c.ExtraValues, result.ExtraValues)
	assert.Equal(t, c.FinalExtraValues, result.FinalExtraValues)
	assert.Equal(t, c.GitCommitWebURL, result.GitCommitWebUrl)
	assert.Equal(t, c.GitCommitTitle, result.GitCommitTitle)
	assert.Equal(t, c.GitCommitAuthor, result.GitCommitAuthor)
	assert.Equal(t, date.ToHumanizeDatetimeString(c.GitCommitDate), result.GitCommitDate)
	assert.Equal(t, date.ToRFC3339DatetimeString(&c.CreatedAt), result.CreatedAt)
	assert.Equal(t, date.ToRFC3339DatetimeString(&c.UpdatedAt), result.UpdatedAt)
	assert.Empty(t, result.DeletedAt)
}

func TestFromChangeLog_DeletedChangelog(t *testing.T) {
	now := time.Now()
	c := &repo.Changelog{
		ID:            1,
		Version:       1,
		Username:      "testUser",
		Config:        "testConfig",
		ConfigChanged: true,
		ProjectID:     1,
		Project:       &repo.Project{},
		CreatedAt:     now,
		UpdatedAt:     now,
		GitBranch:     "testBranch",
		GitCommit:     "testCommit",
		DockerImage:   []string{"testImage1", "testImage2"},
		EnvValues: []*types.KeyValue{
			{
				Key:   "testKey1",
				Value: "testValue1",
			},
		},
		ExtraValues: []*websocket_pb.ExtraValue{
			{
				Path:  "a",
				Value: "v",
			},
		},
		FinalExtraValues: []*websocket_pb.ExtraValue{{Path: "testExtraKey", Value: "testExtraValue"}},
		GitCommitWebURL:  "testUrl",
		GitCommitTitle:   "testTitle",
		GitCommitAuthor:  "testAuthor",
		GitCommitDate:    &now,
		DeletedAt:        &now,
	}
	result := transformer.FromChangeLog(c)
	assert.NotNil(t, result)
	assert.Equal(t, date.ToRFC3339DatetimeString(c.DeletedAt), result.DeletedAt)
}

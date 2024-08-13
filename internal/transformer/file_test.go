package transformer_test

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/assert"
)

func TestFromFile_NilInput(t *testing.T) {
	var f *repo.File
	result := transformer.FromFile(f)
	assert.Nil(t, result)
}

func TestFromFile_ValidInput(t *testing.T) {
	now := time.Now()
	f := &repo.File{
		ID:            1,
		Path:          "testPath",
		Size:          1024,
		Username:      "testUsername",
		Namespace:     "testNamespace",
		Pod:           "testPod",
		Container:     "testContainer",
		ContainerPath: "testContainerPath",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	result := transformer.FromFile(f)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.Id)
	assert.Equal(t, "testPath", result.Path)
	assert.Equal(t, int32(1024), result.Size)
	assert.Equal(t, "testUsername", result.Username)
	assert.Equal(t, "testNamespace", result.Namespace)
	assert.Equal(t, "testPod", result.Pod)
	assert.Equal(t, "testContainer", result.Container)
	assert.Equal(t, "testContainerPath", result.Container_Path)
	assert.Equal(t, humanize.Bytes(1024), result.HumanizeSize)
	assert.Equal(t, date.ToRFC3339DatetimeString(&now), result.CreatedAt)
	assert.Equal(t, date.ToRFC3339DatetimeString(&now), result.UpdatedAt)
	assert.Equal(t, "", result.DeletedAt)
}

func TestFromFile_DeletedFile(t *testing.T) {
	now := time.Now()
	f := &repo.File{
		ID:            1,
		Path:          "testPath",
		Size:          1024,
		Username:      "testUsername",
		Namespace:     "testNamespace",
		Pod:           "testPod",
		Container:     "testContainer",
		ContainerPath: "testContainerPath",
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     &now,
	}
	result := transformer.FromFile(f)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.DeletedAt)
}

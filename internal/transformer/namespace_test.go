package transformer_test

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/stretchr/testify/assert"
)

func TestFromNamespace_NilInput(t *testing.T) {
	var ns *repo.Namespace
	result := transformer.FromNamespace(ns)
	assert.Nil(t, result)
}

func TestFromNamespace_ValidInput(t *testing.T) {
	ns := &repo.Namespace{
		ID:          1,
		Name:        "testNamespace",
		Projects:    nil,
		Description: "x",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	result := transformer.FromNamespace(ns)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.Id)
	assert.Equal(t, "testNamespace", result.Name)
	assert.Len(t, result.Projects, 0)
	assert.Equal(t, "x", result.Description)
}

func TestFromNamespace_DeletedNamespace(t *testing.T) {
	now := time.Now()
	ns := &repo.Namespace{
		ID:        1,
		Name:      "testNamespace",
		Projects:  nil,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: &now,
	}
	result := transformer.FromNamespace(ns)
	assert.NotNil(t, result)
	assert.True(t, result.DeletedAt != "")
}

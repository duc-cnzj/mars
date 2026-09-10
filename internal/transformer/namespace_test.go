package transformer_test

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/stretchr/testify/assert"
)

func TestFromNamespace_NilInput(t *testing.T) {
	var ns *biz.Namespace
	result := transformer.FromNamespace(ns)
	assert.Nil(t, result)
}

func TestFromNamespace_ValidInput(t *testing.T) {
	ns := &biz.Namespace{
		ID:          1,
		Name:        "testNamespace",
		Projects:    []*biz.Project{{ID: 1, Name: "testProject"}},
		Members:     []*biz.Member{{ID: 1, Email: "test@example.com"}},
		Description: "x",
		Private:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	result := transformer.FromNamespace(ns)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.Id)
	assert.Equal(t, "testNamespace", result.Name)
	assert.Len(t, result.Projects, 1)
	assert.Equal(t, int32(1), result.Projects[0].Id)
	assert.Equal(t, "testProject", result.Projects[0].Name)
	assert.Len(t, result.Members, 1)
	assert.Equal(t, int32(1), result.Members[0].Id)
	assert.Equal(t, "test@example.com", result.Members[0].Email)
	assert.Equal(t, "x", result.Description)
	assert.True(t, result.Private)
}

func TestFromNamespace_DeletedNamespace(t *testing.T) {
	now := time.Now()
	ns := &biz.Namespace{
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
	assert.Empty(t, result.Projects)
	assert.Empty(t, result.Members)
	assert.False(t, result.Private)
}

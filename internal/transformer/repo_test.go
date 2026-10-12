package transformer_test

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/stretchr/testify/assert"
)

func TestFromRepo_NilInput(t *testing.T) {
	var r *biz.Repo
	result := transformer.FromRepo(r)
	assert.Nil(t, result)
}

func TestFromRepo_ValidInput(t *testing.T) {
	r := &biz.Repo{
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
	r := &biz.Repo{
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

// WithoutHibernate：nil 安全；无 hibernate 时原样返回（不克隆）；有 hibernate 时返回副本且
// 抹掉该字段、不改动原对象（原对象常引用 biz 层共享配置，就地改会污染调用方）。
func TestWithoutHibernate(t *testing.T) {
	t.Run("nil 输入", func(t *testing.T) {
		assert.Nil(t, transformer.WithoutHibernate(nil))
	})
	t.Run("无 mars_config 原样返回", func(t *testing.T) {
		m := &types.RepoModel{Id: 1, Name: "a"}
		assert.Same(t, m, transformer.WithoutHibernate(m))
	})
	t.Run("无 hibernate 原样返回", func(t *testing.T) {
		m := &types.RepoModel{Id: 1, MarsConfig: &mars.Config{ConfigField: "cfg"}}
		assert.Same(t, m, transformer.WithoutHibernate(m))
	})
	t.Run("抹掉 hibernate 且不改原对象", func(t *testing.T) {
		orig := &types.RepoModel{Id: 1, MarsConfig: &mars.Config{
			ConfigField: "cfg",
			Hibernate:   &mars.Hibernate{Enabled: true, WindowName: "夜间"},
		}}
		out := transformer.WithoutHibernate(orig)
		assert.NotSame(t, orig, out)
		assert.Nil(t, out.MarsConfig.Hibernate, "副本应已抹掉 hibernate")
		assert.Equal(t, "cfg", out.MarsConfig.ConfigField, "其余字段应保留")
		assert.NotNil(t, orig.MarsConfig.Hibernate, "原对象不应被就地修改")
	})
}

func TestFromRepo_ZeroValues(t *testing.T) {
	r := &biz.Repo{
		ID:        1,
		Name:      "zeroRepo",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	result := transformer.FromRepo(r)
	assert.NotNil(t, result)
	assert.False(t, result.Enabled)
	assert.False(t, result.NeedGitRepo)
	assert.Equal(t, &mars.Config{}, result.MarsConfig)
}

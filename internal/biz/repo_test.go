package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeRepoRepoForRepoBiz 只覆盖 RepoBiz 测试用到的 RepoRepo 方法，其余由嵌入接口兜底。
type fakeRepoRepoForRepoBiz struct {
	RepoRepo
	get           func(ctx context.Context, id int) (*Repo, error)
	getByName     func(ctx context.Context, name string) (*Repo, error)
	show          func(ctx context.Context, id int) (*Repo, error)
	create        func(ctx context.Context, in *CreateRepoInput) (*Repo, error)
	update        func(ctx context.Context, in *UpdateRepoInput) (*Repo, error)
	clone         func(ctx context.Context, input *CloneRepoInput) (*Repo, error)
	delete        func(ctx context.Context, id int) error
	toggleEnabled func(ctx context.Context, id int, enabled bool) (*Repo, error)
	all           func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error)
	list          func(ctx context.Context, in *ListRepoRequest) ([]*Repo, *pagination.Pagination, error)
	importFn      func(ctx context.Context, items []*ImportRepoItem) (int, int, error)
	previewImport func(ctx context.Context, items []*ImportRepoItem) (int, int, error)
}

func (f *fakeRepoRepoForRepoBiz) All(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
	return f.all(ctx, in)
}

func (f *fakeRepoRepoForRepoBiz) List(ctx context.Context, in *ListRepoRequest) ([]*Repo, *pagination.Pagination, error) {
	return f.list(ctx, in)
}

func (f *fakeRepoRepoForRepoBiz) Get(ctx context.Context, id int) (*Repo, error) {
	return f.get(ctx, id)
}

func (f *fakeRepoRepoForRepoBiz) GetByName(ctx context.Context, name string) (*Repo, error) {
	return f.getByName(ctx, name)
}

func (f *fakeRepoRepoForRepoBiz) Show(ctx context.Context, id int) (*Repo, error) {
	return f.show(ctx, id)
}

func (f *fakeRepoRepoForRepoBiz) Create(ctx context.Context, in *CreateRepoInput) (*Repo, error) {
	return f.create(ctx, in)
}

func (f *fakeRepoRepoForRepoBiz) Update(ctx context.Context, in *UpdateRepoInput) (*Repo, error) {
	return f.update(ctx, in)
}

func (f *fakeRepoRepoForRepoBiz) Clone(ctx context.Context, input *CloneRepoInput) (*Repo, error) {
	return f.clone(ctx, input)
}

func (f *fakeRepoRepoForRepoBiz) Delete(ctx context.Context, id int) error {
	return f.delete(ctx, id)
}

func (f *fakeRepoRepoForRepoBiz) ToggleEnabled(ctx context.Context, id int, enabled bool) (*Repo, error) {
	return f.toggleEnabled(ctx, id, enabled)
}

func (f *fakeRepoRepoForRepoBiz) Import(ctx context.Context, items []*ImportRepoItem) (int, int, error) {
	return f.importFn(ctx, items)
}

func (f *fakeRepoRepoForRepoBiz) PreviewImport(ctx context.Context, items []*ImportRepoItem) (int, int, error) {
	return f.previewImport(ctx, items)
}

// ---- Create ----

func TestRepoBiz_Create_NilInput(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		create: func(ctx context.Context, in *CreateRepoInput) (*Repo, error) {
			t.Fatal("不应走到数据层创建")
			return nil, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Create(context.TODO(), nil)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo 不能为空", status.Convert(err).Message())
}

func TestRepoBiz_Create_NameTaken(t *testing.T) {
	var created bool
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			assert.Equal(t, "app", name)
			return &Repo{ID: 1, Name: "app"}, nil
		},
		create: func(ctx context.Context, in *CreateRepoInput) (*Repo, error) {
			created = true
			return &Repo{ID: 2, Name: in.Name}, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Create(context.TODO(), &CreateRepoInput{Name: "app"})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo 名称已经存在", status.Convert(err).Message())
	assert.False(t, created)
}

func TestRepoBiz_Create_NameFree(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			return nil, notFoundErr()
		},
		create: func(ctx context.Context, in *CreateRepoInput) (*Repo, error) {
			assert.Equal(t, "app", in.Name)
			return &Repo{ID: 1, Name: in.Name}, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Create(context.TODO(), &CreateRepoInput{Name: "app"})
	assert.Nil(t, err)
	assert.Equal(t, 1, got.ID)
}

func TestRepoBiz_Create_GetByNameError(t *testing.T) {
	var created bool
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			return nil, errors.New("db down")
		},
		create: func(ctx context.Context, in *CreateRepoInput) (*Repo, error) {
			created = true
			return nil, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Create(context.TODO(), &CreateRepoInput{Name: "app"})
	assert.Nil(t, got)
	assert.Equal(t, "db down", err.Error())
	assert.False(t, created)
}

// ---- Update ----

func TestRepoBiz_Update_NameTakenByOther(t *testing.T) {
	var updated bool
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			// 另一仓库（ID=2）已占用该名称，与目标 ID=1 冲突。
			return &Repo{ID: 2, Name: "app"}, nil
		},
		update: func(ctx context.Context, in *UpdateRepoInput) (*Repo, error) {
			updated = true
			return nil, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Update(context.TODO(), &UpdateRepoInput{ID: 1, Name: "app"})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo 名称已经存在", status.Convert(err).Message())
	assert.False(t, updated)
}

func TestRepoBiz_Update_SelfNameNoConflict(t *testing.T) {
	// GetByName 命中自身（同 ID）不视为名称冲突。
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			return &Repo{ID: 1, Name: "app"}, nil
		},
		show: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: 1, Name: "app"}, nil
		},
		update: func(ctx context.Context, in *UpdateRepoInput) (*Repo, error) {
			return &Repo{ID: int(in.ID), Name: in.Name}, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Update(context.TODO(), &UpdateRepoInput{ID: 1, Name: "app"})
	assert.Nil(t, err)
	assert.Equal(t, 1, got.ID)
}

func TestRepoBiz_Update_RenameWithoutProjectsAllowed(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			return nil, notFoundErr()
		},
		show: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: 1, Name: "old"}, nil
		},
		update: func(ctx context.Context, in *UpdateRepoInput) (*Repo, error) {
			assert.Equal(t, "new", in.Name)
			return &Repo{ID: 1, Name: "new"}, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Update(context.TODO(), &UpdateRepoInput{ID: 1, Name: "new"})
	assert.Nil(t, err)
	assert.Equal(t, "new", got.Name)
}

func TestRepoBiz_Update_RenameWithProjectsBlocked(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			return nil, notFoundErr()
		},
		show: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: 1, Name: "old", Projects: []*Project{{Name: "p1"}}}, nil
		},
		update: func(ctx context.Context, in *UpdateRepoInput) (*Repo, error) {
			t.Fatal("不应走到数据层更新")
			return nil, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Update(context.TODO(), &UpdateRepoInput{ID: 1, Name: "new"})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo 下面还有 1 个项目，不能修改名称", status.Convert(err).Message())
}

func TestRepoBiz_Update_GetByNameError(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			return nil, errors.New("db down")
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Update(context.TODO(), &UpdateRepoInput{ID: 1, Name: "new"})
	assert.Nil(t, got)
	assert.Equal(t, "db down", err.Error())
}

func TestRepoBiz_Update_ShowError(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			return nil, notFoundErr()
		},
		show: func(ctx context.Context, id int) (*Repo, error) {
			return nil, errors.New("db down")
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Update(context.TODO(), &UpdateRepoInput{ID: 1, Name: "new"})
	assert.Nil(t, got)
	assert.Equal(t, "db down", err.Error())
}

// ---- Clone ----

func TestRepoBiz_Clone_NameTaken(t *testing.T) {
	var cloned bool
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			assert.Equal(t, "app", name)
			return &Repo{ID: 3, Name: "app"}, nil
		},
		clone: func(ctx context.Context, input *CloneRepoInput) (*Repo, error) {
			cloned = true
			return nil, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Clone(context.TODO(), &CloneRepoInput{ID: 1, Name: "app"})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo 名称已经存在", status.Convert(err).Message())
	assert.False(t, cloned)
}

func TestRepoBiz_Clone_NameFree(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			return nil, notFoundErr()
		},
		clone: func(ctx context.Context, input *CloneRepoInput) (*Repo, error) {
			assert.Equal(t, 1, input.ID)
			assert.Equal(t, "app", input.Name)
			return &Repo{ID: 2, Name: "app"}, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Clone(context.TODO(), &CloneRepoInput{ID: 1, Name: "app"})
	assert.Nil(t, err)
	assert.Equal(t, 2, got.ID)
}

func TestRepoBiz_Clone_GetByNameError(t *testing.T) {
	var cloned bool
	r := &fakeRepoRepoForRepoBiz{
		getByName: func(ctx context.Context, name string) (*Repo, error) {
			return nil, errors.New("db down")
		},
		clone: func(ctx context.Context, input *CloneRepoInput) (*Repo, error) {
			cloned = true
			return nil, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.Clone(context.TODO(), &CloneRepoInput{ID: 1, Name: "app"})
	assert.Nil(t, got)
	assert.Equal(t, "db down", err.Error())
	assert.False(t, cloned)
}

// ---- Delete ----

func TestRepoBiz_Delete_HasProjectsBlocked(t *testing.T) {
	var deleted bool
	r := &fakeRepoRepoForRepoBiz{
		show: func(ctx context.Context, id int) (*Repo, error) {
			assert.Equal(t, 1, id)
			return &Repo{ID: 1, Name: "app", Projects: []*Project{{Name: "p1"}, {Name: "p2"}}}, nil
		},
		delete: func(ctx context.Context, id int) error {
			deleted = true
			return nil
		},
	}
	b := NewRepoBiz(r)
	err := b.Delete(context.TODO(), 1)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo 下面还有 2 个项目，不能删除", status.Convert(err).Message())
	assert.False(t, deleted)
}

func TestRepoBiz_Delete_Happy(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		show: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: 1, Name: "app"}, nil
		},
		delete: func(ctx context.Context, id int) error {
			assert.Equal(t, 1, id)
			return nil
		},
	}
	b := NewRepoBiz(r)
	assert.Nil(t, b.Delete(context.TODO(), 1))
}

func TestRepoBiz_Delete_ShowError(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		show: func(ctx context.Context, id int) (*Repo, error) {
			return nil, errors.New("db down")
		},
	}
	b := NewRepoBiz(r)
	assert.Equal(t, "db down", b.Delete(context.TODO(), 1).Error())
}

// ---- ToggleEnabled ----

func TestRepoBiz_ToggleEnabled_DisableWithProjectsBlocked(t *testing.T) {
	var toggled bool
	r := &fakeRepoRepoForRepoBiz{
		get: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: 1, Name: "app", Enabled: true}, nil
		},
		show: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: 1, Name: "app", Projects: []*Project{{Name: "p1"}}}, nil
		},
		toggleEnabled: func(ctx context.Context, id int, enabled bool) (*Repo, error) {
			toggled = true
			return nil, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.ToggleEnabled(context.TODO(), 1, false)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo 下面还有 1 个项目，不能禁用", status.Convert(err).Message())
	assert.False(t, toggled)
}

func TestRepoBiz_ToggleEnabled_DisableWithoutProjectsAllowed(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		get: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: 1, Name: "app", Enabled: true}, nil
		},
		show: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: 1, Name: "app"}, nil
		},
		toggleEnabled: func(ctx context.Context, id int, enabled bool) (*Repo, error) {
			assert.False(t, enabled)
			return &Repo{ID: 1, Name: "app", Enabled: false}, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.ToggleEnabled(context.TODO(), 1, false)
	assert.Nil(t, err)
	assert.False(t, got.Enabled)
}

func TestRepoBiz_ToggleEnabled_EnableNoProjectsCheck(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		get: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: 1, Name: "app", Enabled: false}, nil
		},
		toggleEnabled: func(ctx context.Context, id int, enabled bool) (*Repo, error) {
			assert.True(t, enabled)
			return &Repo{ID: 1, Name: "app", Enabled: true}, nil
		},
	}
	b := NewRepoBiz(r)
	got, err := b.ToggleEnabled(context.TODO(), 1, true)
	assert.Nil(t, err)
	assert.True(t, got.Enabled)
}

func TestRepoBiz_ToggleEnabled_GetError(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		get: func(ctx context.Context, id int) (*Repo, error) {
			return nil, errors.New("db down")
		},
	}
	b := NewRepoBiz(r)
	got, err := b.ToggleEnabled(context.TODO(), 1, true)
	assert.Nil(t, got)
	assert.Equal(t, "db down", err.Error())
}

func TestRepoBiz_ToggleEnabled_ShowError(t *testing.T) {
	r := &fakeRepoRepoForRepoBiz{
		get: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: 1, Name: "app", Enabled: true}, nil
		},
		show: func(ctx context.Context, id int) (*Repo, error) {
			return nil, errors.New("db down")
		},
	}
	b := NewRepoBiz(r)
	got, err := b.ToggleEnabled(context.TODO(), 1, false)
	assert.Nil(t, got)
	assert.Equal(t, "db down", err.Error())
}

// ---- 输入合法性校验（空名/非法 id 在业务规则前拦截，repo 不被调用）----

func TestRepoBiz_Create_EmptyName(t *testing.T) {
	var created bool
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{create: func(ctx context.Context, in *CreateRepoInput) (*Repo, error) {
		created = true
		return &Repo{}, nil
	}})
	got, err := b.Create(context.TODO(), &CreateRepoInput{Name: ""})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo 名称不能为空", status.Convert(err).Message())
	assert.False(t, created)
}

func TestRepoBiz_Update_InvalidID(t *testing.T) {
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{})
	got, err := b.Update(context.TODO(), &UpdateRepoInput{ID: 0})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo id 不能小于等于 0", status.Convert(err).Message())
}

func TestRepoBiz_Delete_InvalidID(t *testing.T) {
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{})
	err := b.Delete(context.TODO(), 0)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo id 不能小于等于 0", status.Convert(err).Message())
}

func TestRepoBiz_Clone_InvalidID(t *testing.T) {
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{})
	got, err := b.Clone(context.TODO(), &CloneRepoInput{ID: 0, Name: "app"})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo id 不能小于等于 0", status.Convert(err).Message())
}

func TestRepoBiz_Clone_EmptyName(t *testing.T) {
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{})
	got, err := b.Clone(context.TODO(), &CloneRepoInput{ID: 1, Name: ""})
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo 名称不能为空", status.Convert(err).Message())
}

func TestRepoBiz_ToggleEnabled_InvalidID(t *testing.T) {
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{})
	got, err := b.ToggleEnabled(context.TODO(), 0, false)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "repo id 不能小于等于 0", status.Convert(err).Message())
}

// ---- 纯透传查询（All/List/Get/Show）----

func TestRepoBiz_All_Passthrough(t *testing.T) {
	enabled := true
	var allCalled bool
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		all: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
			allCalled = true
			assert.True(t, *in.Enabled)
			return []*Repo{{ID: 1, Name: "app"}}, nil
		},
	})
	got, err := b.All(context.TODO(), &AllRepoRequest{Enabled: &enabled})
	assert.NoError(t, err)
	assert.True(t, allCalled)
	assert.Len(t, got, 1)
}

func TestRepoBiz_List_Passthrough(t *testing.T) {
	var listCalled bool
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		list: func(ctx context.Context, in *ListRepoRequest) ([]*Repo, *pagination.Pagination, error) {
			listCalled = true
			assert.Equal(t, int32(1), in.Page)
			return []*Repo{{ID: 1, Name: "app"}}, &pagination.Pagination{Page: 1}, nil
		},
	})
	got, pag, err := b.List(context.TODO(), &ListRepoRequest{Page: 1})
	assert.NoError(t, err)
	assert.True(t, listCalled)
	assert.Len(t, got, 1)
	assert.Equal(t, int32(1), pag.Page)
}

func TestRepoBiz_Get_Passthrough(t *testing.T) {
	var getCalled bool
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		get: func(ctx context.Context, id int) (*Repo, error) {
			getCalled = true
			assert.Equal(t, 1, id)
			return &Repo{ID: 1, Name: "app"}, nil
		},
	})
	got, err := b.Get(context.TODO(), 1)
	assert.NoError(t, err)
	assert.True(t, getCalled)
	assert.Equal(t, "app", got.Name)
}

func TestRepoBiz_Show_Passthrough(t *testing.T) {
	var showCalled bool
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		show: func(ctx context.Context, id int) (*Repo, error) {
			showCalled = true
			return &Repo{ID: id, Name: "app"}, nil
		},
	})
	got, err := b.Show(context.TODO(), 2)
	assert.NoError(t, err)
	assert.True(t, showCalled)
	assert.Equal(t, 2, got.ID)
}

// ---- Import ----

func TestRepoBiz_Import_EmptyItems(t *testing.T) {
	var called bool
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		importFn: func(ctx context.Context, items []*ImportRepoItem) (int, int, error) {
			called = true
			return 0, 0, nil
		},
	})
	created, updated, err := b.Import(context.TODO(), nil)
	assert.NoError(t, err)
	assert.Zero(t, created)
	assert.Zero(t, updated)
	assert.False(t, called)
}

func TestRepoBiz_Import_NilItem(t *testing.T) {
	var called bool
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		importFn: func(ctx context.Context, items []*ImportRepoItem) (int, int, error) {
			called = true
			return 0, 0, nil
		},
	})
	created, updated, err := b.Import(context.TODO(), []*ImportRepoItem{nil})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "导入数据不能为空", status.Convert(err).Message())
	assert.Zero(t, created)
	assert.Zero(t, updated)
	assert.False(t, called)
}

func TestRepoBiz_Import_InvalidName(t *testing.T) {
	for _, name := range []string{"", "bad name", "bad/name", "中文"} {
		t.Run(name, func(t *testing.T) {
			var called bool
			b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
				importFn: func(ctx context.Context, items []*ImportRepoItem) (int, int, error) {
					called = true
					return 0, 0, nil
				},
			})
			created, updated, err := b.Import(context.TODO(), []*ImportRepoItem{{Name: name}})
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Contains(t, status.Convert(err).Message(), "不合法")
			assert.Zero(t, created)
			assert.Zero(t, updated)
			assert.False(t, called)
		})
	}
}

func TestRepoBiz_Import_DuplicateName(t *testing.T) {
	// 同一导入文件内同名：若放行，pre-phase 两条都判 create，无唯一约束下会落重复行。
	// 必须在委托前整体拒绝（零部分变更）。
	var called bool
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		importFn: func(ctx context.Context, items []*ImportRepoItem) (int, int, error) {
			called = true
			return 0, 0, nil
		},
	})
	created, updated, err := b.Import(context.TODO(), []*ImportRepoItem{
		{Name: "app", Enabled: true},
		{Name: "app", Enabled: false},
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "重复")
	assert.Zero(t, created)
	assert.Zero(t, updated)
	assert.False(t, called)
}

func TestRepoBiz_Import_ValidationBeforeAnyMutation(t *testing.T) {
	// 校验先行：首个 name 合法、第二个非法时，import 不被调用（零部分变更）。
	var called bool
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		importFn: func(ctx context.Context, items []*ImportRepoItem) (int, int, error) {
			called = true
			return 0, 0, nil
		},
	})
	created, updated, err := b.Import(context.TODO(), []*ImportRepoItem{
		{Name: "valid"},
		{Name: "invalid name"},
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Zero(t, created)
	assert.Zero(t, updated)
	assert.False(t, called)
}

func TestRepoBiz_Import_Delegates(t *testing.T) {
	// 校验通过后整体委托 data 层事务导入，created/updated 计数原样透传。
	items := []*ImportRepoItem{
		{Name: "app", Enabled: true, Description: "desc"},
		{Name: "new", Enabled: false},
	}
	var got []*ImportRepoItem
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		importFn: func(ctx context.Context, in []*ImportRepoItem) (int, int, error) {
			got = in
			return 1, 1, nil
		},
	})
	created, updated, err := b.Import(context.TODO(), items)
	assert.NoError(t, err)
	assert.Equal(t, 1, created)
	assert.Equal(t, 1, updated)
	assert.Equal(t, items, got)
}

func TestRepoBiz_Import_ErrorPassthrough(t *testing.T) {
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		importFn: func(ctx context.Context, in []*ImportRepoItem) (int, int, error) {
			return 0, 0, errors.New("db down")
		},
	})
	created, updated, err := b.Import(context.TODO(), []*ImportRepoItem{{Name: "app"}})
	assert.Equal(t, "db down", err.Error())
	assert.Zero(t, created)
	assert.Zero(t, updated)
}

// ---- PreviewImport ----

func TestRepoBiz_PreviewImport_EmptyItems(t *testing.T) {
	var called bool
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		previewImport: func(ctx context.Context, items []*ImportRepoItem) (int, int, error) {
			called = true
			return 0, 0, nil
		},
	})
	created, updated, err := b.PreviewImport(context.TODO(), nil)
	assert.NoError(t, err)
	assert.Zero(t, created)
	assert.Zero(t, updated)
	assert.False(t, called)
}

func TestRepoBiz_PreviewImport_InvalidItems(t *testing.T) {
	// 干跑与真实导入共用 validateImportItems：任一 nil/name 非法/重复即整体拒绝，不调 repo。
	cases := []struct {
		name  string
		items []*ImportRepoItem
		msg   string
	}{
		{"nil item", []*ImportRepoItem{nil}, "导入数据不能为空"},
		{"invalid name", []*ImportRepoItem{{Name: "bad name"}}, "不合法"},
		{"duplicate name", []*ImportRepoItem{{Name: "app"}, {Name: "app"}}, "重复"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var called bool
			b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
				previewImport: func(ctx context.Context, items []*ImportRepoItem) (int, int, error) {
					called = true
					return 0, 0, nil
				},
			})
			created, updated, err := b.PreviewImport(context.TODO(), tc.items)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Contains(t, status.Convert(err).Message(), tc.msg)
			assert.Zero(t, created)
			assert.Zero(t, updated)
			assert.False(t, called)
		})
	}
}

func TestRepoBiz_PreviewImport_Delegates(t *testing.T) {
	items := []*ImportRepoItem{{Name: "app"}, {Name: "new"}}
	var got []*ImportRepoItem
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		previewImport: func(ctx context.Context, in []*ImportRepoItem) (int, int, error) {
			got = in
			return 1, 1, nil
		},
	})
	created, updated, err := b.PreviewImport(context.TODO(), items)
	assert.NoError(t, err)
	assert.Equal(t, 1, created)
	assert.Equal(t, 1, updated)
	assert.Equal(t, items, got)
}

func TestRepoBiz_PreviewImport_ErrorPassthrough(t *testing.T) {
	b := NewRepoBiz(&fakeRepoRepoForRepoBiz{
		previewImport: func(ctx context.Context, in []*ImportRepoItem) (int, int, error) {
			return 0, 0, errors.New("db down")
		},
	})
	created, updated, err := b.PreviewImport(context.TODO(), []*ImportRepoItem{{Name: "app"}})
	assert.Equal(t, "db down", err.Error())
	assert.Zero(t, created)
	assert.Zero(t, updated)
}

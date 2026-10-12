package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeHibernateWindowRepo 以函数字段实现端口，未设置的写方法被调用即 panic（证明未被触碰）。
type fakeHibernateWindowRepo struct {
	HibernateWindowRepo
	listFn    func(ctx context.Context) ([]*HibernateWindow, error)
	getFn     func(ctx context.Context, name string) (*HibernateWindow, error)
	getByIDFn func(ctx context.Context, id int) (*HibernateWindow, error)
	refsFn    func(ctx context.Context, name string) ([]string, error)
	createFn  func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error)
	updateFn  func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error)
	cascadeFn func(ctx context.Context, w *HibernateWindow, oldName string) (*HibernateWindow, int, error)
	deleteFn  func(ctx context.Context, id int) error
}

func (f *fakeHibernateWindowRepo) List(ctx context.Context) ([]*HibernateWindow, error) {
	return f.listFn(ctx)
}

func (f *fakeHibernateWindowRepo) GetByName(ctx context.Context, name string) (*HibernateWindow, error) {
	return f.getFn(ctx, name)
}

func (f *fakeHibernateWindowRepo) GetByID(ctx context.Context, id int) (*HibernateWindow, error) {
	return f.getByIDFn(ctx, id)
}

func (f *fakeHibernateWindowRepo) ReposReferencing(ctx context.Context, name string) ([]string, error) {
	return f.refsFn(ctx, name)
}

func (f *fakeHibernateWindowRepo) Create(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
	return f.createFn(ctx, w)
}

func (f *fakeHibernateWindowRepo) Update(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
	return f.updateFn(ctx, w)
}

func (f *fakeHibernateWindowRepo) UpdateCascade(ctx context.Context, w *HibernateWindow, oldName string) (*HibernateWindow, int, error) {
	return f.cascadeFn(ctx, w, oldName)
}

func (f *fakeHibernateWindowRepo) Delete(ctx context.Context, id int) error {
	return f.deleteFn(ctx, id)
}

// TestValidateHibernateWindow 覆盖名称空/纯空白、启止时间非法、启止相同与合法各分支。
func TestValidateHibernateWindow(t *testing.T) {
	// 合法。
	assert.NoError(t, ValidateHibernateWindow("夜间", "23:00", "06:00"))
	assert.NoError(t, ValidateHibernateWindow("午休", "12:00", "13:00"))

	// 名称空 / 纯空白。
	assert.Equal(t, codes.InvalidArgument, status.Code(ValidateHibernateWindow("", "23:00", "06:00")))
	assert.Equal(t, codes.InvalidArgument, status.Code(ValidateHibernateWindow("   ", "23:00", "06:00")))

	// 开始时间非法（含 24:00、非数字）。
	assert.Equal(t, codes.InvalidArgument, status.Code(ValidateHibernateWindow("夜间", "bad", "06:00")))
	assert.Equal(t, codes.InvalidArgument, status.Code(ValidateHibernateWindow("夜间", "24:00", "06:00")))

	// 结束时间非法。
	assert.Equal(t, codes.InvalidArgument, status.Code(ValidateHibernateWindow("夜间", "23:00", "bad")))
	assert.Equal(t, codes.InvalidArgument, status.Code(ValidateHibernateWindow("夜间", "23:00", "12:60")))

	// 启止相同。
	assert.Equal(t, codes.InvalidArgument, status.Code(ValidateHibernateWindow("夜间", "23:00", "23:00")))
}

// TestHibernateWindowBiz_List 覆盖透传查询。
func TestHibernateWindowBiz_List(t *testing.T) {
	b := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		listFn: func(ctx context.Context) ([]*HibernateWindow, error) {
			return []*HibernateWindow{{ID: 1, Name: "夜间"}}, nil
		},
	})
	got, err := b.List(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "夜间", got[0].Name)
}

// TestHibernateWindowBiz_Create 覆盖校验失败短路（不触达 repo）与合法落库。
func TestHibernateWindowBiz_Create(t *testing.T) {
	// 校验失败：repo 的 Create 未设置，触发即 panic，证明被短路。
	b := NewHibernateWindowBiz(&fakeHibernateWindowRepo{})
	if _, err := b.Create(context.TODO(), &HibernateWindow{Name: "", Start: "23:00", End: "06:00"}); err == nil {
		t.Fatal("空名称应校验失败")
	}

	var created *HibernateWindow
	b = NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		createFn: func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
			created = w
			w.ID = 9
			return w, nil
		},
	})
	got, err := b.Create(context.TODO(), &HibernateWindow{Name: "夜间", Start: "23:00", End: "06:00", Description: "过夜"})
	assert.NoError(t, err)
	assert.Equal(t, 9, got.ID)
	assert.Equal(t, "夜间", created.Name)

	// repo 错误上抛。
	errB := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		createFn: func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
			return nil, errors.New("db down")
		},
	})
	if _, err := errB.Create(context.TODO(), &HibernateWindow{Name: "夜间", Start: "23:00", End: "06:00"}); err == nil {
		t.Fatal("repo 错误应上抛")
	}
}

// TestHibernateWindowBiz_Update 覆盖更新路径：校验失败短路、未改名放行（不查引用）、
// 非级联改名（无引用放行 / 有引用 400 / ReposReferencing 错误上抛）、级联改名（迁移引用返回数量 /
// UpdateCascade 错误上抛）、窗口不存在透传 404、GetByID 与 repo.Update 错误上抛。
func TestHibernateWindowBiz_Update(t *testing.T) {
	// 校验失败：repo 的 GetByID/Update 未设置，触发即 panic，证明被短路。
	b := NewHibernateWindowBiz(&fakeHibernateWindowRepo{})
	if _, _, err := b.Update(context.TODO(), &HibernateWindow{ID: 1, Name: "夜间", Start: "23:00", End: "23:00"}, false); err == nil {
		t.Fatal("启止相同应校验失败")
	}

	// 未改名：不触达 ReposReferencing/UpdateCascade（未设置即 panic），直接落库，迁移数为 0。
	var updated *HibernateWindow
	sameName := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "午休"}, nil
		},
		updateFn: func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
			updated = w
			return w, nil
		},
	})
	got, n, err := sameName.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "午休", Start: "12:00", End: "13:00"}, true)
	assert.NoError(t, err)
	assert.Equal(t, 3, got.ID)
	assert.Equal(t, "午休", updated.Name)
	assert.Zero(t, n)

	// 非级联改名且旧名无引用：放行，ReposReferencing 以旧名查询，迁移数为 0。
	var renamed *HibernateWindow
	okRename := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) {
			assert.Equal(t, "夜间", name)
			return nil, nil
		},
		updateFn: func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
			renamed = w
			return w, nil
		},
	})
	got, n, err = okRename.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, false)
	assert.NoError(t, err)
	assert.Equal(t, 3, got.ID)
	assert.Equal(t, "深夜", renamed.Name)
	assert.Zero(t, n)

	// 非级联改名但旧名仍被引用：拒改（400），Update/UpdateCascade 均未被调用（未设置即 panic）。
	blockedRename := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) {
			return []string{"repo-a", "repo-b"}, nil
		},
	})
	_, _, err = blockedRename.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, false)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "夜间")
	assert.Contains(t, status.Convert(err).Message(), "repo-a, repo-b")

	// 非级联改名：ReposReferencing 错误上抛（Update 不被调用）。
	refsErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) { return nil, errors.New("boom") },
	})
	if _, _, err := refsErr.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, false); err == nil {
		t.Fatal("ReposReferencing 错误应上抛")
	}

	// 级联改名：交由 repo.UpdateCascade（以旧名为参），迁移数透传；不触达 ReposReferencing/Update。
	cascaded := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		cascadeFn: func(ctx context.Context, w *HibernateWindow, oldName string) (*HibernateWindow, int, error) {
			assert.Equal(t, "夜间", oldName)
			assert.Equal(t, "深夜", w.Name)
			return w, 3, nil
		},
	})
	got, migrated, err := cascaded.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, true)
	assert.NoError(t, err)
	assert.Equal(t, 3, got.ID)
	assert.Equal(t, 3, migrated)

	// 级联改名：UpdateCascade 错误上抛。
	cascadeErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		cascadeFn: func(ctx context.Context, w *HibernateWindow, oldName string) (*HibernateWindow, int, error) {
			return nil, 0, errors.New("duplicate name")
		},
	})
	if _, _, err := cascadeErr.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, true); err == nil {
		t.Fatal("UpdateCascade 错误应上抛")
	}

	// 窗口不存在（GetByID 返回 nil）：跳过引用校验，交由 Update 落 404。
	var notFoundTarget *HibernateWindow
	missing := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return nil, nil },
		updateFn: func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
			notFoundTarget = w
			return nil, errors.New("not found")
		},
	})
	if _, _, err := missing.Update(context.TODO(), &HibernateWindow{ID: 99, Name: "夜间", Start: "23:00", End: "06:00"}, true); err == nil {
		t.Fatal("窗口不存在应经 Update 返回错误")
	}
	assert.Equal(t, 99, notFoundTarget.ID)

	// GetByID 错误上抛。
	getErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return nil, errors.New("db down") },
	})
	if _, _, err := getErr.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, true); err == nil {
		t.Fatal("GetByID 错误应上抛")
	}

	// 非级联改名：repo.Update 错误（如名称重复）上抛。
	updErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) { return nil, nil },
		updateFn: func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
			return nil, errors.New("duplicate name")
		},
	})
	if _, _, err := updErr.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, false); err == nil {
		t.Fatal("repo.Update 错误应上抛")
	}
}

// TestHibernateWindowBiz_Delete 覆盖删除前关联校验的全部路径：
// 窗口不存在幂等成功、无引用可删、有引用拒删（400 且不触达删除）、GetByID/ReposReferencing 错误上抛。
func TestHibernateWindowBiz_Delete(t *testing.T) {
	// 窗口不存在：幂等成功，不触达 ReposReferencing/Delete（未设置即 panic）。
	missing := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return nil, nil },
	})
	assert.NoError(t, missing.Delete(context.TODO(), 5))

	// 无引用：正常删除，ID 透传。
	var deleted int
	ok := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) {
			assert.Equal(t, "夜间", name)
			return nil, nil
		},
		deleteFn: func(ctx context.Context, id int) error {
			deleted = id
			return nil
		},
	})
	assert.NoError(t, ok.Delete(context.TODO(), 5))
	assert.Equal(t, 5, deleted)

	// 有引用：拒删（400），Delete 未被调用（未设置即 panic）。
	blocked := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) {
			return []string{"repo-a", "repo-b"}, nil
		},
	})
	err := blocked.Delete(context.TODO(), 5)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "夜间")
	assert.Contains(t, status.Convert(err).Message(), "repo-a, repo-b")

	// GetByID 错误上抛。
	getErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return nil, errors.New("db down") },
	})
	if err := getErr.Delete(context.TODO(), 5); err == nil {
		t.Fatal("GetByID 错误应上抛")
	}

	// ReposReferencing 错误上抛（Delete 不被调用）。
	refsErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) { return nil, errors.New("boom") },
	})
	if err := refsErr.Delete(context.TODO(), 5); err == nil {
		t.Fatal("ReposReferencing 错误应上抛")
	}
}

// TestSummarizeHibernateWindowRefs 覆盖仓库名短串压缩：≤5 全列、>5 截断并标注总数。
func TestSummarizeHibernateWindowRefs(t *testing.T) {
	assert.Equal(t, "a, b", summarizeHibernateWindowRefs([]string{"a", "b"}))
	assert.Equal(t, "a, b, c, d, e", summarizeHibernateWindowRefs([]string{"a", "b", "c", "d", "e"}))
	assert.Equal(t, "a, b, c, d, e 等 7 个", summarizeHibernateWindowRefs([]string{"a", "b", "c", "d", "e", "f", "g"}))
}

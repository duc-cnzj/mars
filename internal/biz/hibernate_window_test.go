package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// newWindowBiz 构造只关心 HibernateWindowRepo 的 biz（repoRepo 传 nil）：仅 List/Create/Update/Delete
// 这类不触达仓库聚合的用例用它；需要候选仓库/批量改写的用例直接调 NewHibernateWindowBiz 传 fakeRepoRepo。
func newWindowBiz(repo HibernateWindowRepo) HibernateWindowBiz {
	return NewHibernateWindowBiz(repo, nil)
}

// fakeHibernateWindowRepo 以函数字段实现端口，未设置的写方法被调用即 panic（证明未被触碰）。
type fakeHibernateWindowRepo struct {
	HibernateWindowRepo
	listFn             func(ctx context.Context) ([]*HibernateWindow, error)
	getFn              func(ctx context.Context, name string) (*HibernateWindow, error)
	getByIDFn          func(ctx context.Context, id int) (*HibernateWindow, error)
	windowRefsFn       func(ctx context.Context) (map[string][]string, error)
	refsFn             func(ctx context.Context, name string) ([]string, error)
	projectsRefsFn     func(ctx context.Context, name string) ([]int, error)
	projectRefsRepoIDs func(ctx context.Context, name string) (map[int]int, error)
	createFn           func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error)
	updateFn           func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error)
	cascadeFn          func(ctx context.Context, w *HibernateWindow, oldName string) (*HibernateWindow, int, error)
	deleteFn           func(ctx context.Context, id int) error
}

// fakeRepoRepo 以函数字段实现仓库端口，未设置的写方法被调用即 panic（证明未被触碰）。
type fakeRepoRepo struct {
	RepoRepo
	allFn   func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error)
	batchFn func(ctx context.Context, configs map[int]*mars.Config) error
}

func (f *fakeRepoRepo) All(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
	return f.allFn(ctx, in)
}

func (f *fakeRepoRepo) BatchUpdateMarsConfig(ctx context.Context, configs map[int]*mars.Config) error {
	return f.batchFn(ctx, configs)
}

// hbCfg 造一个含休眠段的仓库配置：副本字段用可承载 "0" 的 InputNumber（enabled=true 时过校验）。
func hbCfg(enabled bool, windows []string) *mars.Config {
	return &mars.Config{
		Hibernate: &mars.Hibernate{Enabled: enabled, ReplicaPaths: []string{"replicas"}, WindowNames: windows},
		Elements:  []*mars.Element{{Path: "replicas", Type: mars.ElementType_ElementTypeInputNumber}},
	}
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

func (f *fakeHibernateWindowRepo) WindowRepoRefs(ctx context.Context) (map[string][]string, error) {
	return f.windowRefsFn(ctx)
}

func (f *fakeHibernateWindowRepo) ReposReferencing(ctx context.Context, name string) ([]string, error) {
	return f.refsFn(ctx, name)
}

func (f *fakeHibernateWindowRepo) ProjectsReferencing(ctx context.Context, name string) ([]int, error) {
	return f.projectsRefsFn(ctx, name)
}

func (f *fakeHibernateWindowRepo) ProjectReferencingRepoIDs(ctx context.Context, name string) (map[int]int, error) {
	return f.projectRefsRepoIDs(ctx, name)
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

// TestHibernateWindowBiz_List 覆盖：查询失败上抛、空列表短路（不查引用）、正常填充每窗 RepoNames、
// 引用查询失败上抛。
func TestHibernateWindowBiz_List(t *testing.T) {
	// 查询失败上抛。
	listErr := newWindowBiz(&fakeHibernateWindowRepo{
		listFn: func(ctx context.Context) ([]*HibernateWindow, error) { return nil, errors.New("db down") },
	})
	if _, err := listErr.List(context.TODO()); err == nil {
		t.Fatal("List 错误应上抛")
	}

	// 空列表：短路返回，不触达 WindowRepoRefs（未设置即 panic）。
	empty := newWindowBiz(&fakeHibernateWindowRepo{
		listFn: func(ctx context.Context) ([]*HibernateWindow, error) { return nil, nil },
	})
	got, err := empty.List(context.TODO())
	assert.NoError(t, err)
	assert.Empty(t, got)

	// 正常：单趟扫描引用映射，按窗名填 RepoNames（无引用的窗为 nil）。
	b := newWindowBiz(&fakeHibernateWindowRepo{
		listFn: func(ctx context.Context) ([]*HibernateWindow, error) {
			return []*HibernateWindow{{ID: 1, Name: "夜间"}, {ID: 2, Name: "午休"}}, nil
		},
		windowRefsFn: func(ctx context.Context) (map[string][]string, error) {
			return map[string][]string{"夜间": {"repo-a", "repo-b"}}, nil
		},
	})
	got, err = b.List(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "夜间", got[0].Name)
	assert.Equal(t, []string{"repo-a", "repo-b"}, got[0].RepoNames)
	assert.Equal(t, "午休", got[1].Name)
	assert.Nil(t, got[1].RepoNames)

	// 引用查询失败上抛。
	refsErr := newWindowBiz(&fakeHibernateWindowRepo{
		listFn: func(ctx context.Context) ([]*HibernateWindow, error) {
			return []*HibernateWindow{{ID: 1, Name: "夜间"}}, nil
		},
		windowRefsFn: func(ctx context.Context) (map[string][]string, error) { return nil, errors.New("boom") },
	})
	if _, err := refsErr.List(context.TODO()); err == nil {
		t.Fatal("WindowRepoRefs 错误应上抛")
	}
}

// TestHibernateWindowBiz_ListCandidateRepos 覆盖：查询失败上抛、过滤未配休眠段的仓库并按名升序。
func TestHibernateWindowBiz_ListCandidateRepos(t *testing.T) {
	allErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) { return nil, errors.New("db down") },
	})
	if _, err := allErr.ListCandidateRepos(context.TODO()); err == nil {
		t.Fatal("All 错误应上抛")
	}

	b := NewHibernateWindowBiz(&fakeHibernateWindowRepo{}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
			return []*Repo{
				{ID: 2, Name: "beta"}, // 未配休眠段 → 排除
				{ID: 1, Name: "alpha", MarsConfig: hbCfg(false, nil)},
				{ID: 3, Name: "gamma", MarsConfig: hbCfg(false, nil)},
			}, nil
		},
	})
	got, err := b.ListCandidateRepos(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, 1, got[0].ID)
	assert.Equal(t, "alpha", got[0].Name)
	assert.Equal(t, 3, got[1].ID)
	assert.Equal(t, "gamma", got[1].Name)
}

// TestHibernateWindowBiz_SetWindowRepos 覆盖集合语义重设引用的全部路径：窗不存在 404、GetByID/All
// 错误上抛、repo_ids 含非候选 400、无变更不触达批量写、加入与移除并存、移除启用仓库最后一个窗 400、
// 批量写失败上抛。
func TestHibernateWindowBiz_SetWindowRepos(t *testing.T) {
	window := func() *HibernateWindow { return &HibernateWindow{ID: 1, Name: "夜间"} }

	// 窗不存在 → 404（不触达 All）。
	missing := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return nil, nil },
	}, &fakeRepoRepo{})
	_, err := missing.SetWindowRepos(context.TODO(), 9, nil)
	assert.Equal(t, codes.NotFound, status.Code(err))

	// GetByID 错误上抛。
	getErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return nil, errors.New("db down") },
	}, &fakeRepoRepo{})
	if _, err := getErr.SetWindowRepos(context.TODO(), 9, nil); err == nil {
		t.Fatal("GetByID 错误应上抛")
	}

	// All 错误上抛。
	allErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return window(), nil },
	}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) { return nil, errors.New("db down") },
	})
	if _, err := allErr.SetWindowRepos(context.TODO(), 1, []int{1}); err == nil {
		t.Fatal("All 错误应上抛")
	}

	// repo_ids 含非候选仓库 → 400（消息点出仓库 ID）。
	nonCandidate := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return window(), nil },
	}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
			return []*Repo{{ID: 1, Name: "alpha", MarsConfig: hbCfg(false, nil)}}, nil
		},
	})
	_, err = nonCandidate.SetWindowRepos(context.TODO(), 1, []int{99})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "99")

	// 无变更：候选已含本窗、目标也含 → 不触达批量写（batchFn 未设置即 panic），返回现有引用。
	nothing := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return window(), nil },
		refsFn: func(ctx context.Context, name string) ([]string, error) {
			assert.Equal(t, "夜间", name)
			return []string{"alpha"}, nil
		},
	}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
			return []*Repo{{ID: 1, Name: "alpha", MarsConfig: hbCfg(false, []string{"夜间"})}}, nil
		},
	})
	names, err := nothing.SetWindowRepos(context.TODO(), 1, []int{1})
	assert.NoError(t, err)
	assert.Equal(t, []string{"alpha"}, names)

	// 加入与移除并存：alpha 原本无本窗→加入；beta 原本含本窗→移除。移除 beta 会触发子集守卫，
	// 令 ProjectReferencingRepoIDs 返回「无项目引用」放行（否则触发未设置的函数字段即 panic）。
	var applied map[int]*mars.Config
	both := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return window(), nil },
		refsFn:    func(ctx context.Context, name string) ([]string, error) { return []string{"alpha"}, nil },
		projectRefsRepoIDs: func(ctx context.Context, name string) (map[int]int, error) {
			return map[int]int{}, nil
		},
	}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
			return []*Repo{
				{ID: 1, Name: "alpha", MarsConfig: hbCfg(false, []string{"午休"})},
				{ID: 2, Name: "beta", MarsConfig: hbCfg(false, []string{"夜间", "午休"})},
			}, nil
		},
		batchFn: func(ctx context.Context, configs map[int]*mars.Config) error {
			applied = configs
			return nil
		},
	})
	names, err = both.SetWindowRepos(context.TODO(), 1, []int{1})
	assert.NoError(t, err)
	assert.Equal(t, []string{"alpha"}, names)
	assert.Len(t, applied, 2)
	assert.Equal(t, []string{"午休", "夜间"}, applied[1].Hibernate.WindowNames)
	assert.Equal(t, []string{"午休"}, applied[2].Hibernate.WindowNames)

	// 移除某启用仓库的最后一个窗 → 400（守住「启用必须有生效窗」），不触达批量写。
	// 移除 alpha 的窗会先触发子集守卫，令 ProjectReferencingRepoIDs 返回「无项目引用」放行，
	// 以继续走到 ValidateHibernateConfig 的 400。
	lastWindow := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return window(), nil },
		projectRefsRepoIDs: func(ctx context.Context, name string) (map[int]int, error) {
			return map[int]int{}, nil
		},
	}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
			return []*Repo{{ID: 1, Name: "alpha", MarsConfig: hbCfg(true, []string{"夜间"})}}, nil
		},
	})
	_, err = lastWindow.SetWindowRepos(context.TODO(), 1, nil)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "alpha")

	// 子集守卫命中：从 alpha（ID 1）移除窗 W，其下项目 42 仍选定 W → 400，点名仓库与项目，
	// 且不触达批量写（batchFn 未设置即 panic）。
	guardBlocked := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return window(), nil },
		projectRefsRepoIDs: func(ctx context.Context, name string) (map[int]int, error) {
			assert.Equal(t, "夜间", name)
			return map[int]int{42: 1}, nil
		},
	}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
			return []*Repo{{ID: 1, Name: "alpha", MarsConfig: hbCfg(false, []string{"夜间"})}}, nil
		},
	})
	_, err = guardBlocked.SetWindowRepos(context.TODO(), 1, nil)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "alpha")
	assert.Contains(t, status.Convert(err).Message(), "42")

	// 收窄正确：beta（ID 2）失去 W，但引用 W 的项目 42 归属仓库 1（不在受影响集合内）→ 放行，
	// 照常改写 beta 的配置（集合语义移除 W）。
	var scoped map[int]*mars.Config
	scopedOK := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return window(), nil },
		projectRefsRepoIDs: func(ctx context.Context, name string) (map[int]int, error) {
			return map[int]int{42: 1}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) { return []string{"alpha"}, nil },
	}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
			return []*Repo{
				{ID: 1, Name: "alpha", MarsConfig: hbCfg(false, []string{"夜间"})},
				{ID: 2, Name: "beta", MarsConfig: hbCfg(false, []string{"夜间", "午休"})},
			}, nil
		},
		batchFn: func(ctx context.Context, configs map[int]*mars.Config) error {
			scoped = configs
			return nil
		},
	})
	names, err = scopedOK.SetWindowRepos(context.TODO(), 1, []int{1})
	assert.NoError(t, err)
	assert.Equal(t, []string{"alpha"}, names)
	assert.Len(t, scoped, 1)
	assert.Equal(t, []string{"午休"}, scoped[2].Hibernate.WindowNames)

	// 守卫查询失败上抛（不触达批量写）。
	guardErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return window(), nil },
		projectRefsRepoIDs: func(ctx context.Context, name string) (map[int]int, error) {
			return nil, errors.New("db down")
		},
	}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
			return []*Repo{{ID: 1, Name: "alpha", MarsConfig: hbCfg(false, []string{"夜间"})}}, nil
		},
	})
	if _, err := guardErr.SetWindowRepos(context.TODO(), 1, nil); err == nil {
		t.Fatal("ProjectReferencingRepoIDs 错误应上抛")
	}

	// 批量写失败上抛。
	batchErr := NewHibernateWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return window(), nil },
	}, &fakeRepoRepo{
		allFn: func(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
			return []*Repo{{ID: 1, Name: "alpha", MarsConfig: hbCfg(false, nil)}}, nil
		},
		batchFn: func(ctx context.Context, configs map[int]*mars.Config) error { return errors.New("tx fail") },
	})
	if _, err := batchErr.SetWindowRepos(context.TODO(), 1, []int{1}); err == nil {
		t.Fatal("BatchUpdateMarsConfig 错误应上抛")
	}
}

// TestHibernateWindowBiz_Create 覆盖校验失败短路（不触达 repo）与合法落库。
func TestHibernateWindowBiz_Create(t *testing.T) {
	// 校验失败：repo 的 Create 未设置，触发即 panic，证明被短路。
	b := newWindowBiz(&fakeHibernateWindowRepo{})
	if _, err := b.Create(context.TODO(), &HibernateWindow{Name: "", Start: "23:00", End: "06:00"}); err == nil {
		t.Fatal("空名称应校验失败")
	}

	var created *HibernateWindow
	b = newWindowBiz(&fakeHibernateWindowRepo{
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
	errB := newWindowBiz(&fakeHibernateWindowRepo{
		createFn: func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
			return nil, errors.New("db down")
		},
	})
	if _, err := errB.Create(context.TODO(), &HibernateWindow{Name: "夜间", Start: "23:00", End: "06:00"}); err == nil {
		t.Fatal("repo 错误应上抛")
	}
}

// TestHibernateWindowBiz_Create_TrimsName 覆盖名称归一化：落库前 trim，令「 夜间 」与「夜间」对齐为
// 同一自然键（引用侧一律取 normalizeWindowNames 的规范名，否则按名引用会未命中、休眠静默失效）。
func TestHibernateWindowBiz_Create_TrimsName(t *testing.T) {
	var created *HibernateWindow
	b := newWindowBiz(&fakeHibernateWindowRepo{
		createFn: func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
			created = w
			w.ID = 9
			return w, nil
		},
	})
	got, err := b.Create(context.TODO(), &HibernateWindow{Name: "  夜间  ", Start: "23:00", End: "06:00"})
	assert.NoError(t, err)
	assert.Equal(t, "夜间", created.Name)
	assert.Equal(t, "夜间", got.Name)
}

// TestHibernateWindowBiz_Update_TrimsName 覆盖名称归一化：更新前先 trim 再比对是否改名。既有记录
// 名「午休」、本次提交「 午休 」trim 后同名 → 视为未改名（不触发引用校验：refsFn 未设置即 panic），
// 直接落库。若 trim 晚于比对，此处会误判为改名并进入非级联守卫。
func TestHibernateWindowBiz_Update_TrimsName(t *testing.T) {
	var updated *HibernateWindow
	b := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "午休"}, nil
		},
		updateFn: func(ctx context.Context, w *HibernateWindow) (*HibernateWindow, error) {
			updated = w
			return w, nil
		},
	})
	got, n, err := b.Update(context.TODO(), &HibernateWindow{ID: 3, Name: " 午休 ", Start: "12:00", End: "13:00"}, false)
	assert.NoError(t, err)
	assert.Equal(t, 3, got.ID)
	assert.Equal(t, "午休", updated.Name)
	assert.Zero(t, n)
}

// TestHibernateWindowBiz_Update 覆盖更新路径：校验失败短路、未改名放行（不查引用）、
// 非级联改名（无引用放行 / 有引用 400 / ReposReferencing 错误上抛）、级联改名（迁移引用返回数量 /
// UpdateCascade 错误上抛）、窗口不存在透传 404、GetByID 与 repo.Update 错误上抛。
func TestHibernateWindowBiz_Update(t *testing.T) {
	// 校验失败：repo 的 GetByID/Update 未设置，触发即 panic，证明被短路。
	b := newWindowBiz(&fakeHibernateWindowRepo{})
	if _, _, err := b.Update(context.TODO(), &HibernateWindow{ID: 1, Name: "夜间", Start: "23:00", End: "23:00"}, false); err == nil {
		t.Fatal("启止相同应校验失败")
	}

	// 未改名：不触达 ReposReferencing/UpdateCascade（未设置即 panic），直接落库，迁移数为 0。
	var updated *HibernateWindow
	sameName := newWindowBiz(&fakeHibernateWindowRepo{
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
	okRename := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) {
			assert.Equal(t, "夜间", name)
			return nil, nil
		},
		projectsRefsFn: func(ctx context.Context, name string) ([]int, error) {
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
	blockedRename := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) {
			return []string{"repo-a", "repo-b"}, nil
		},
		projectsRefsFn: func(ctx context.Context, name string) ([]int, error) {
			return nil, nil
		},
	})
	_, _, err = blockedRename.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, false)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "夜间")
	assert.Contains(t, status.Convert(err).Message(), "repo-a, repo-b")

	// 非级联改名但「仅项目」引用（无仓库引用）：同样拒改（400），错误消息点出项目 ID。
	projectOnlyRename := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn:         func(ctx context.Context, name string) ([]string, error) { return nil, nil },
		projectsRefsFn: func(ctx context.Context, name string) ([]int, error) { return []int{7, 9}, nil },
	})
	_, _, err = projectOnlyRename.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, false)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "项目 7, 9")

	// 非级联改名：ReposReferencing 错误上抛（Update 不被调用）。
	refsErr := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) { return nil, errors.New("boom") },
	})
	if _, _, err := refsErr.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, false); err == nil {
		t.Fatal("ReposReferencing 错误应上抛")
	}

	// 非级联改名：仓库引用查询成功但项目引用查询失败，同样上抛（Update 不被调用）。
	projectsRefsErr := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn:         func(ctx context.Context, name string) ([]string, error) { return nil, nil },
		projectsRefsFn: func(ctx context.Context, name string) ([]int, error) { return nil, errors.New("boom") },
	})
	if _, _, err := projectsRefsErr.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, false); err == nil {
		t.Fatal("ProjectsReferencing 错误应上抛")
	}

	// 级联改名：交由 repo.UpdateCascade（以旧名为参），迁移数透传；不触达 ReposReferencing/Update。
	cascaded := newWindowBiz(&fakeHibernateWindowRepo{
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
	cascadeErr := newWindowBiz(&fakeHibernateWindowRepo{
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
	missing := newWindowBiz(&fakeHibernateWindowRepo{
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
	getErr := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return nil, errors.New("db down") },
	})
	if _, _, err := getErr.Update(context.TODO(), &HibernateWindow{ID: 3, Name: "深夜", Start: "23:00", End: "06:00"}, true); err == nil {
		t.Fatal("GetByID 错误应上抛")
	}

	// 非级联改名：repo.Update 错误（如名称重复）上抛。
	updErr := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn:         func(ctx context.Context, name string) ([]string, error) { return nil, nil },
		projectsRefsFn: func(ctx context.Context, name string) ([]int, error) { return nil, nil },
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
	missing := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return nil, nil },
	})
	assert.NoError(t, missing.Delete(context.TODO(), 5))

	// 无引用：正常删除，ID 透传。
	var deleted int
	ok := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) {
			assert.Equal(t, "夜间", name)
			return nil, nil
		},
		projectsRefsFn: func(ctx context.Context, name string) ([]int, error) {
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
	blocked := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) {
			return []string{"repo-a", "repo-b"}, nil
		},
		projectsRefsFn: func(ctx context.Context, name string) ([]int, error) { return nil, nil },
	})
	err := blocked.Delete(context.TODO(), 5)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "夜间")
	assert.Contains(t, status.Convert(err).Message(), "repo-a, repo-b")

	// 「仅项目」引用（无仓库引用）：同样拒删（400），错误消息点出项目 ID。
	projectOnly := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn:         func(ctx context.Context, name string) ([]string, error) { return nil, nil },
		projectsRefsFn: func(ctx context.Context, name string) ([]int, error) { return []int{11}, nil },
	})
	err = projectOnly.Delete(context.TODO(), 5)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "项目 11")

	// GetByID 错误上抛。
	getErr := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) { return nil, errors.New("db down") },
	})
	if err := getErr.Delete(context.TODO(), 5); err == nil {
		t.Fatal("GetByID 错误应上抛")
	}

	// ReposReferencing 错误上抛（Delete 不被调用）。
	refsErr := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn: func(ctx context.Context, name string) ([]string, error) { return nil, errors.New("boom") },
	})
	if err := refsErr.Delete(context.TODO(), 5); err == nil {
		t.Fatal("ReposReferencing 错误应上抛")
	}

	// ProjectsReferencing 错误上抛（Delete 不被调用）。
	projectsRefsErr := newWindowBiz(&fakeHibernateWindowRepo{
		getByIDFn: func(ctx context.Context, id int) (*HibernateWindow, error) {
			return &HibernateWindow{ID: id, Name: "夜间"}, nil
		},
		refsFn:         func(ctx context.Context, name string) ([]string, error) { return nil, nil },
		projectsRefsFn: func(ctx context.Context, name string) ([]int, error) { return nil, errors.New("boom") },
	})
	if err := projectsRefsErr.Delete(context.TODO(), 5); err == nil {
		t.Fatal("ProjectsReferencing 错误应上抛")
	}
}

// TestSummarizeHibernateWindowRefs 覆盖仓库名短串压缩：≤5 全列、>5 截断并标注总数。
func TestSummarizeHibernateWindowRefs(t *testing.T) {
	assert.Equal(t, "a, b", summarizeHibernateWindowRefs([]string{"a", "b"}))
	assert.Equal(t, "a, b, c, d, e", summarizeHibernateWindowRefs([]string{"a", "b", "c", "d", "e"}))
	assert.Equal(t, "a, b, c, d, e 等 7 个", summarizeHibernateWindowRefs([]string{"a", "b", "c", "d", "e", "f", "g"}))
}

// TestSummarizeHibernateWindowIDs 覆盖项目 ID 短串压缩：≤5 全列、>5 截断并标注总数。
func TestSummarizeHibernateWindowIDs(t *testing.T) {
	assert.Equal(t, "1, 2", summarizeHibernateWindowIDs([]int{1, 2}))
	assert.Equal(t, "1, 2, 3, 4, 5", summarizeHibernateWindowIDs([]int{1, 2, 3, 4, 5}))
	assert.Equal(t, "1, 2, 3, 4, 5 等 7 个", summarizeHibernateWindowIDs([]int{1, 2, 3, 4, 5, 6, 7}))
}

// TestDescribeHibernateWindowRefs 覆盖引用短串拼接：仅仓库 / 仅项目 / 两者并存（逗号顿开）。
func TestDescribeHibernateWindowRefs(t *testing.T) {
	assert.Equal(t, "仓库 a, b", describeHibernateWindowRefs([]string{"a", "b"}, nil))
	assert.Equal(t, "项目 1, 2", describeHibernateWindowRefs(nil, []int{1, 2}))
	assert.Equal(t, "仓库 a，项目 1", describeHibernateWindowRefs([]string{"a"}, []int{1}))
}

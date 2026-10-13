package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/stretchr/testify/assert"
)

// fakeHibernateRepo 以函数字段实现端口，未设置的写方法被调用即 panic（证明未被触碰）。
type fakeHibernateRepo struct {
	HibernateRepo
	getFn    func(ctx context.Context, projectID int) (*ProjectHibernate, error)
	listFn   func(ctx context.Context, projectIDs ...int) (map[int]*ProjectHibernate, error)
	upsertFn func(ctx context.Context, in *ProjectHibernate) (*ProjectHibernate, error)
	deleteFn func(ctx context.Context, projectID int) error
}

func (f *fakeHibernateRepo) Get(ctx context.Context, projectID int) (*ProjectHibernate, error) {
	return f.getFn(ctx, projectID)
}

func (f *fakeHibernateRepo) ListByProjectIDs(ctx context.Context, projectIDs ...int) (map[int]*ProjectHibernate, error) {
	return f.listFn(ctx, projectIDs...)
}

func (f *fakeHibernateRepo) Upsert(ctx context.Context, in *ProjectHibernate) (*ProjectHibernate, error) {
	return f.upsertFn(ctx, in)
}

func (f *fakeHibernateRepo) Delete(ctx context.Context, projectID int) error {
	return f.deleteFn(ctx, projectID)
}

// fakeHibernateProjectRepo 只实现 Toggle 用到的 Show。
type fakeHibernateProjectRepo struct {
	ProjectRepo
	showFn func(ctx context.Context, id int) (*Project, error)
}

func (f *fakeHibernateProjectRepo) Show(ctx context.Context, id int) (*Project, error) {
	return f.showFn(ctx, id)
}

// fakeHibernateRepoRepo 只实现 Toggle 用到的 Get 与 Show。
type fakeHibernateRepoRepo struct {
	RepoRepo
	getFn func(ctx context.Context, id int) (*Repo, error)
}

func (f *fakeHibernateRepoRepo) Get(ctx context.Context, id int) (*Repo, error) {
	return f.getFn(ctx, id)
}

// TestInHibernateWindow 覆盖同天窗口、跨零点窗口与边界（起始含、结束不含）。
func TestInHibernateWindow(t *testing.T) {
	at := func(h, m int) time.Time { return time.Date(2026, 10, 10, h, m, 0, 0, time.Local) }

	// 同天窗口 [09:00, 18:00)
	assert.False(t, InHibernateWindow(at(8, 59), "09:00", "18:00"))
	assert.True(t, InHibernateWindow(at(9, 0), "09:00", "18:00"))
	assert.True(t, InHibernateWindow(at(17, 59), "09:00", "18:00"))
	assert.False(t, InHibernateWindow(at(18, 0), "09:00", "18:00"))

	// 跨零点窗口 [23:00, 06:00)
	assert.True(t, InHibernateWindow(at(23, 30), "23:00", "06:00"))
	assert.True(t, InHibernateWindow(at(3, 0), "23:00", "06:00"))
	assert.True(t, InHibernateWindow(at(23, 0), "23:00", "06:00"))
	assert.False(t, InHibernateWindow(at(6, 0), "23:00", "06:00"))
	assert.False(t, InHibernateWindow(at(12, 0), "23:00", "06:00"))
}

// TestInHibernateWindow_Invalid 覆盖非法时刻与起止相同：一律视为无窗口。
func TestInHibernateWindow_Invalid(t *testing.T) {
	now := time.Date(2026, 10, 10, 0, 30, 0, 0, time.Local)
	assert.False(t, InHibernateWindow(now, "bad", "06:00"))
	assert.False(t, InHibernateWindow(now, "23:00", "24:00"))
	assert.False(t, InHibernateWindow(now, "23:00", "23:00"))
	assert.False(t, InHibernateWindow(now, "", ""))
}

// TestParseClockMinutes 覆盖解析合法与各类非法输入。
func TestParseClockMinutes(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"00:00", 0, true},
		{"23:59", 23*60 + 59, true},
		{"bad", 0, false},
		{"12:60", 0, false},
		{"24:00", 0, false},
		{"-1:00", 0, false},
		{"12:ab", 0, false},
		{"12:00:00", 0, false},
	}
	for _, c := range cases {
		got, ok := parseClockMinutes(c.in)
		assert.Equalf(t, c.ok, ok, "input %q", c.in)
		if c.ok {
			assert.Equalf(t, c.want, got, "input %q", c.in)
		}
	}
}

// TestSleepExtraValues 覆盖命中字段置 0、缺失字段追加、非副本字段原样保留且不改动入参切片。
func TestSleepExtraValues(t *testing.T) {
	cur := []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "3", Description: "副本数"},
		{Path: "app->image", Value: "nginx:1.0", Description: "镜像"},
	}
	out := SleepExtraValues(cur, []string{"app->replicas", "worker->replicas"})

	assert.Equal(t, "0", out[0].Value)
	assert.Equal(t, "副本数", out[0].Description)
	assert.Equal(t, "nginx:1.0", out[1].Value)
	assert.Equal(t, "worker->replicas", out[2].Path)
	assert.Equal(t, "0", out[2].Value)
	// 入参未被改写，且不是同一元素指针（新切片独立对象）。
	assert.Equal(t, "3", cur[0].Value)
	assert.NotSame(t, cur[0], out[0])
}

// TestSnapshotReplicaValues 覆盖只截取副本字段、缺失字段跳过。
func TestSnapshotReplicaValues(t *testing.T) {
	all := []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "3"},
		{Path: "app->image", Value: "nginx:1.0"},
	}
	out := SnapshotReplicaValues(all, []string{"app->replicas", "missing"})
	assert.Len(t, out, 1)
	assert.Equal(t, "app->replicas", out[0].Path)
	assert.Equal(t, "3", out[0].Value)
}

// TestWakeExtraValues 覆盖用快照回填副本字段、不动快照之外的字段、不新增字段。
func TestWakeExtraValues(t *testing.T) {
	cur := []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "0"},
		{Path: "app->image", Value: "nginx:1.0", Description: "镜像"},
	}
	snapshot := []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "3"},
		{Path: "not->exist", Value: "9"},
	}
	out := WakeExtraValues(cur, snapshot)
	assert.Len(t, out, 2)
	assert.Equal(t, "3", out[0].Value)
	assert.Equal(t, "nginx:1.0", out[1].Value)
	assert.NotSame(t, cur[0], out[0])
}

// TestReplicaValuesAllZero 覆盖副本字段全 0 判据的各分支：全 0 为休眠态；任一非 0 或字段
// 缺失都判为「已被外部改动」；空副本字段列表按无改动处理（真空真）。
func TestReplicaValuesAllZero(t *testing.T) {
	cur := []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "0"},
		{Path: "app->min", Value: "0"},
		{Path: "app->image", Value: "nginx:1.0"},
	}
	paths := []string{"app->replicas", "app->min"}

	// 全部副本字段为 "0"（非副本字段不参与判定）→ 仍在休眠态。
	assert.True(t, ReplicaValuesAllZero(cur, paths))

	// 任一副本字段非 0 → 已被用户改动。
	modified := []*websocket_pb.ExtraValue{
		{Path: "app->replicas", Value: "5"},
		{Path: "app->min", Value: "0"},
	}
	assert.False(t, ReplicaValuesAllZero(modified, paths))

	// 副本字段缺失（不在 cur 中）→ 视为已改动（休眠落态会把缺失字段补齐为 "0"）。
	assert.False(t, ReplicaValuesAllZero(cur, []string{"app->ghost"}))

	// 无副本字段可判 → 真空真，交回正常唤醒流程。
	assert.True(t, ReplicaValuesAllZero(cur, nil))
	assert.True(t, ReplicaValuesAllZero(nil, nil))
}

// TestValidateHibernateConfig 覆盖禁用短路与启用时各校验分支（时间窗按 window_names 列表引用）。
func TestValidateHibernateConfig(t *testing.T) {
	inputNumber := &mars.Element{Path: "app->replicas", Type: mars.ElementType_ElementTypeInputNumber}
	selectWithZero := &mars.Element{Path: "web->replicas", Type: mars.ElementType_ElementTypeSelect, SelectValues: []string{"0", "2"}}
	selectNoZero := &mars.Element{Path: "web->replicas", Type: mars.ElementType_ElementTypeSelect, SelectValues: []string{"2", "3"}}
	text := &mars.Element{Path: "app->note", Type: mars.ElementType_ElementTypeInput}

	assert.NoError(t, ValidateHibernateConfig(nil))
	assert.NoError(t, ValidateHibernateConfig(&mars.Config{}))
	assert.NoError(t, ValidateHibernateConfig(&mars.Config{Hibernate: &mars.Hibernate{Enabled: false}}))

	// 启用但副本字段为空。
	assert.Error(t, ValidateHibernateConfig(&mars.Config{
		Hibernate: &mars.Hibernate{Enabled: true, WindowNames: []string{"夜间"}},
	}))

	// 启用、字段不在已定义元素中。
	assert.Error(t, ValidateHibernateConfig(&mars.Config{
		Elements:  []*mars.Element{inputNumber},
		Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"nope"}, WindowNames: []string{"夜间"}},
	}))

	// 字段类型无法承载 0（字符串）。
	assert.Error(t, ValidateHibernateConfig(&mars.Config{
		Elements:  []*mars.Element{text},
		Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"app->note"}, WindowNames: []string{"夜间"}},
	}))

	// Select 白名单不含 0。
	assert.Error(t, ValidateHibernateConfig(&mars.Config{
		Elements:  []*mars.Element{selectNoZero},
		Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"web->replicas"}, WindowNames: []string{"夜间"}},
	}))

	// 副本字段合法但未选时间窗（window_names 为空/纯空白/全空白）。
	assert.Error(t, ValidateHibernateConfig(&mars.Config{
		Elements:  []*mars.Element{inputNumber},
		Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"app->replicas"}},
	}))
	assert.Error(t, ValidateHibernateConfig(&mars.Config{
		Elements:  []*mars.Element{inputNumber},
		Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"app->replicas"}, WindowNames: []string{"  "}},
	}))
	assert.Error(t, ValidateHibernateConfig(&mars.Config{
		Elements:  []*mars.Element{inputNumber},
		Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"app->replicas"}, WindowNames: []string{"", "  "}},
	}))

	// 合法：InputNumber + 白名单含 0 的 Select + 已选多个时间窗。
	assert.NoError(t, ValidateHibernateConfig(&mars.Config{
		Elements:  []*mars.Element{inputNumber, selectWithZero},
		Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"app->replicas", "web->replicas"}, WindowNames: []string{"夜间", "午休"}},
	}))
}

// TestElementSupportsZero 覆盖各元素类型对 "0" 的承载判定。
func TestElementSupportsZero(t *testing.T) {
	assert.True(t, elementSupportsZero(&mars.Element{Type: mars.ElementType_ElementTypeInputNumber}))
	assert.True(t, elementSupportsZero(&mars.Element{Type: mars.ElementType_ElementTypeRadio, SelectValues: []string{"0"}}))
	assert.True(t, elementSupportsZero(&mars.Element{Type: mars.ElementType_ElementTypeNumberSelect, SelectValues: []string{"0"}}))
	assert.True(t, elementSupportsZero(&mars.Element{Type: mars.ElementType_ElementTypeNumberRadio, SelectValues: []string{"0"}}))
	assert.False(t, elementSupportsZero(&mars.Element{Type: mars.ElementType_ElementTypeSelect, SelectValues: []string{"1"}}))
	assert.False(t, elementSupportsZero(&mars.Element{Type: mars.ElementType_ElementTypeInput}))
}

// TestHibernateBiz_ListByProjectIDs 覆盖透传查询：空入参不发查询（直接返回空表）。
func TestHibernateBiz_ListByProjectIDs(t *testing.T) {
	called := false
	h := NewHibernateBiz(&fakeHibernateRepo{
		listFn: func(ctx context.Context, projectIDs ...int) (map[int]*ProjectHibernate, error) {
			called = true
			return map[int]*ProjectHibernate{1: {ProjectID: 1}}, nil
		},
	}, nil, nil)

	empty, err := h.ListByProjectIDs(context.TODO())
	assert.NoError(t, err)
	assert.Empty(t, empty)
	assert.False(t, called, "空入参不应触达 repo")

	got2, err := h.ListByProjectIDs(context.TODO(), 1)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Len(t, got2, 1)
}

// TestHibernateBiz_Toggle 覆盖开关切换：项目不存在上抛；启用时仓库未启用休眠报非法参数；
// 启用时未选时间窗（空选择）报非法参数；启用成功后保留既有运行态与快照。
func TestHibernateBiz_Toggle(t *testing.T) {
	projRepo := &fakeHibernateProjectRepo{
		showFn: func(ctx context.Context, id int) (*Project, error) {
			return &Project{ID: id, RepoID: 5}, nil
		},
	}
	repoRepo := &fakeHibernateRepoRepo{
		getFn: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: id, Name: "r", MarsConfig: &mars.Config{Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"a"}, WindowNames: []string{"夜间"}}}}, nil
		},
	}

	var upserted *ProjectHibernate
	repo := &fakeHibernateRepo{
		getFn: func(ctx context.Context, projectID int) (*ProjectHibernate, error) {
			return &ProjectHibernate{ProjectID: projectID, State: HibernateStateAwake}, nil
		},
		upsertFn: func(ctx context.Context, in *ProjectHibernate) (*ProjectHibernate, error) {
			upserted = in
			return in, nil
		},
		deleteFn: func(ctx context.Context, projectID int) error { return nil },
	}
	h := NewHibernateBiz(repo, projRepo, repoRepo)

	got, err := h.Toggle(context.TODO(), 1, true, []string{"夜间"})
	assert.NoError(t, err)
	assert.True(t, got.Enabled)
	assert.NotNil(t, upserted)

	// 启用时未选任何时间窗：空选择不再表示「继承仓库全量」，而是「未选」——拒绝（400）且不落库。
	upserted = nil
	if _, err := h.Toggle(context.TODO(), 1, true, nil); err == nil {
		t.Fatal("启用时未选时间窗应报错")
	}
	assert.Nil(t, upserted)

	// 项目不存在。
	badProj := &fakeHibernateProjectRepo{showFn: func(ctx context.Context, id int) (*Project, error) {
		return nil, errors.New("not found")
	}}
	if _, err := NewHibernateBiz(repo, badProj, repoRepo).Toggle(context.TODO(), 1, true, nil); err == nil {
		t.Fatal("项目不存在应报错")
	}

	// 仓库未启用休眠：启用开关应报非法参数。
	offRepoRepo := &fakeHibernateRepoRepo{getFn: func(ctx context.Context, id int) (*Repo, error) {
		return &Repo{ID: id, Name: "r", MarsConfig: &mars.Config{}}, nil
	}}
	if _, err := NewHibernateBiz(repo, projRepo, offRepoRepo).Toggle(context.TODO(), 1, true, nil); err == nil {
		t.Fatal("仓库未启用休眠时启用开关应报错")
	}
	// 关闭开关不校验仓库能力。
	if _, err := NewHibernateBiz(repo, projRepo, offRepoRepo).Toggle(context.TODO(), 1, false, nil); err != nil {
		t.Fatalf("关闭开关不应校验仓库能力: %v", err)
	}

	// 仓库已启用休眠但副本字段为空：仍视为未启用休眠。
	emptyPathsRepoRepo := &fakeHibernateRepoRepo{getFn: func(ctx context.Context, id int) (*Repo, error) {
		return &Repo{ID: id, Name: "r", MarsConfig: &mars.Config{Hibernate: &mars.Hibernate{Enabled: true}}}, nil
	}}
	if _, err := NewHibernateBiz(repo, projRepo, emptyPathsRepoRepo).Toggle(context.TODO(), 1, true, nil); err == nil {
		t.Fatal("副本字段为空时启用开关应报错")
	}

	// 仓库已启用休眠、副本字段非空但未选时间窗：视为未启用休眠。
	noWindowRepoRepo := &fakeHibernateRepoRepo{getFn: func(ctx context.Context, id int) (*Repo, error) {
		return &Repo{ID: id, Name: "r", MarsConfig: &mars.Config{Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"a"}}}}, nil
	}}
	if _, err := NewHibernateBiz(repo, projRepo, noWindowRepoRepo).Toggle(context.TODO(), 1, true, nil); err == nil {
		t.Fatal("未选时间窗时启用开关应报错")
	}

	// 仓库查询失败上抛。
	errRepoRepo := &fakeHibernateRepoRepo{getFn: func(ctx context.Context, id int) (*Repo, error) {
		return nil, errors.New("repo not found")
	}}
	if _, err := NewHibernateBiz(repo, projRepo, errRepoRepo).Toggle(context.TODO(), 1, true, nil); err == nil {
		t.Fatal("仓库查询失败应上抛")
	}

	// 休眠记录查询失败上抛。
	errHibernateRepo := &fakeHibernateRepo{getFn: func(ctx context.Context, projectID int) (*ProjectHibernate, error) {
		return nil, errors.New("db down")
	}}
	if _, err := NewHibernateBiz(errHibernateRepo, projRepo, repoRepo).Toggle(context.TODO(), 1, true, nil); err == nil {
		t.Fatal("休眠记录查询失败应上抛")
	}

	// 既有休眠记录带快照：关闭开关时保留运行态与快照。
	var kept *ProjectHibernate
	withSnapshot := &fakeHibernateRepo{
		getFn: func(ctx context.Context, projectID int) (*ProjectHibernate, error) {
			return &ProjectHibernate{ProjectID: projectID, State: HibernateStateSleeping, Snapshot: []*websocket_pb.ExtraValue{{Path: "a", Value: "3"}}}, nil
		},
		upsertFn: func(ctx context.Context, in *ProjectHibernate) (*ProjectHibernate, error) {
			kept = in
			return in, nil
		},
	}
	if _, err := NewHibernateBiz(withSnapshot, projRepo, repoRepo).Toggle(context.TODO(), 1, false, nil); err != nil {
		t.Fatalf("关闭开关失败: %v", err)
	}
	assert.Equal(t, HibernateStateSleeping, kept.State)
	assert.Len(t, kept.Snapshot, 1)
	// 关闭时携带空窗（清空时间窗 = 关闭休眠）：落库的 WindowNames 须被清空（不含任何窗）。
	assert.Empty(t, kept.WindowNames)

	// 重新开启但副本字段非全 0（关闭期间被手动改过）：陈旧休眠态名不副实，重置为清醒并清空快照。
	var corrected *ProjectHibernate
	staleProj := &fakeHibernateProjectRepo{showFn: func(ctx context.Context, id int) (*Project, error) {
		return &Project{ID: id, RepoID: 5, FinalExtraValues: []*websocket_pb.ExtraValue{{Path: "a", Value: "2"}}}, nil
	}}
	staleRepo := &fakeHibernateRepo{
		getFn: func(ctx context.Context, projectID int) (*ProjectHibernate, error) {
			return &ProjectHibernate{ProjectID: projectID, State: HibernateStateSleeping,
				Snapshot: []*websocket_pb.ExtraValue{{Path: "a", Value: "3"}}}, nil
		},
		upsertFn: func(ctx context.Context, in *ProjectHibernate) (*ProjectHibernate, error) {
			corrected = in
			return in, nil
		},
	}
	if _, err := NewHibernateBiz(staleRepo, staleProj, repoRepo).Toggle(context.TODO(), 1, true, []string{"夜间"}); err != nil {
		t.Fatalf("重新开启失败: %v", err)
	}
	assert.Equal(t, HibernateStateAwake, corrected.State)
	assert.Empty(t, corrected.Snapshot)

	// 重新开启且副本字段确实仍全 0（真在休眠）：沿用 sleeping 与快照，不误清。
	var genuine *ProjectHibernate
	genuineProj := &fakeHibernateProjectRepo{showFn: func(ctx context.Context, id int) (*Project, error) {
		return &Project{ID: id, RepoID: 5, FinalExtraValues: []*websocket_pb.ExtraValue{{Path: "a", Value: "0"}}}, nil
	}}
	genuineRepo := &fakeHibernateRepo{
		getFn: func(ctx context.Context, projectID int) (*ProjectHibernate, error) {
			return &ProjectHibernate{ProjectID: projectID, State: HibernateStateSleeping,
				Snapshot: []*websocket_pb.ExtraValue{{Path: "a", Value: "3"}}}, nil
		},
		upsertFn: func(ctx context.Context, in *ProjectHibernate) (*ProjectHibernate, error) {
			genuine = in
			return in, nil
		},
	}
	if _, err := NewHibernateBiz(genuineRepo, genuineProj, repoRepo).Toggle(context.TODO(), 1, true, []string{"夜间"}); err != nil {
		t.Fatalf("重新开启失败: %v", err)
	}
	assert.Equal(t, HibernateStateSleeping, genuine.State)
	assert.Len(t, genuine.Snapshot, 1)
}

// TestHibernateBiz_Toggle_WindowSubset 覆盖项目窗口子集的校验与落库：合法子集归一化后落库、
// 空选择被拒（400，不再继承仓库全量）、越界名被拒（400）。
func TestHibernateBiz_Toggle_WindowSubset(t *testing.T) {
	projRepo := &fakeHibernateProjectRepo{
		showFn: func(ctx context.Context, id int) (*Project, error) {
			return &Project{ID: id, RepoID: 5}, nil
		},
	}
	repoRepo := &fakeHibernateRepoRepo{
		getFn: func(ctx context.Context, id int) (*Repo, error) {
			return &Repo{ID: id, Name: "r", MarsConfig: &mars.Config{
				Hibernate: &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"a"}, WindowNames: []string{"夜间", "午休"}},
			}}, nil
		},
	}
	var upserted *ProjectHibernate
	repo := &fakeHibernateRepo{
		getFn: func(ctx context.Context, projectID int) (*ProjectHibernate, error) {
			return nil, nil
		},
		upsertFn: func(ctx context.Context, in *ProjectHibernate) (*ProjectHibernate, error) {
			upserted = in
			return in, nil
		},
	}
	h := NewHibernateBiz(repo, projRepo, repoRepo)

	// 合法子集（含空白与重复）：归一化（trim/去空/去重、保序）后落库。
	got, err := h.Toggle(context.TODO(), 1, true, []string{"夜间", " 夜间 ", "午休", ""})
	assert.NoError(t, err)
	assert.Equal(t, []string{"夜间", "午休"}, got.WindowNames)
	assert.Equal(t, []string{"夜间", "午休"}, upserted.WindowNames)

	// 空选择：不再表示继承仓库全量，而是「未选」——启用即拒（400），不落库。
	upserted = nil
	_, err = h.Toggle(context.TODO(), 1, true, nil)
	assert.Error(t, err)
	assert.Nil(t, upserted)

	// 越界名：不在仓库可用窗列表中，拒绝（400）。
	_, err = h.Toggle(context.TODO(), 1, true, []string{"不存在"})
	assert.Error(t, err)
}

package deploy

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
)

// TestInstallProject 验证部署编排顺序：GlobalLock → Validate → LoadConfigs → Run →
// Finish/Error，全链路成功时返回 nil。
func TestInstallProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	job := NewMockJob(m)
	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)
	assert.Nil(t, InstallProject(context.TODO(), job))
}

// TestHasUserTokenTemplate 判定 values 是否引用 <.UserToken>（容忍空白）：
// 启用 AI 沙箱的前置校验依赖它拦住「部署时不签发令牌」的软失败配置。
func TestHasUserTokenTemplate(t *testing.T) {
	tests := []struct {
		name   string
		values string
		want   bool
	}{
		{name: "引用模板", values: "mars:\n  token: \"<.UserToken>\"", want: true},
		{name: "容忍空白", values: "token: < .UserToken >", want: true},
		{name: "未引用", values: "mars:\n  token: \"\"", want: false},
		{name: "大小写不符", values: "token: \"<.usertoken>\"", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HasUserTokenTemplate(tt.values))
		})
	}
}

// fakeSleeper 是进度器测试用的休眠替身：立即返回，避免测试真实等待。
type fakeSleeper struct{}

// Sleep 空实现，满足 Sleeper 接口且不引入等待。
func (f fakeSleeper) Sleep(time.Duration) {}

// TestNewProcessPercent 验证进度器构造：绑定 messager、初始进度为 0。
func TestNewProcessPercent(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	pp := NewProcessPercent(msger, fakeSleeper{}).(*processPercent)
	assert.Equal(t, msger, pp.msger)
	assert.Equal(t, int64(0), pp.Current())
}

// TestProcessPercent_Add 验证进度递增与上报；到达 100 后 Add 为 no-op，不再增长。
func TestProcessPercent_Add(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	msger.EXPECT().SendProcessPercent(gomock.Any()).AnyTimes()
	pp := NewProcessPercent(msger, fakeSleeper{}).(*processPercent)
	pp.Add()
	assert.Equal(t, int64(1), pp.Current())
	// 达到 100 后 Add 是 no-op，不再上报
	pp.percent = 100
	pp.Add()
	assert.Equal(t, int64(100), pp.Current())
}

// TestProcessPercent_To 验证 To 设置进度：目标低于当前值时直接落盘目标值，
// 不会倒退等待（覆盖「目标 <= 当前」分支）。
func TestProcessPercent_To(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	msger.EXPECT().SendProcessPercent(gomock.Any()).AnyTimes()
	pp := NewProcessPercent(msger, fakeSleeper{}).(*processPercent)
	pp.To(10)
	assert.Equal(t, int64(10), pp.Current())
	// 目标低于当前值：走 else 直接落盘当前值
	pp.To(5)
	assert.Equal(t, int64(5), pp.Current())
}

// TestRealSleeper 验证真实休眠器至少等待指定时长。
func TestRealSleeper(t *testing.T) {
	s := NewRealSleeper()
	start := time.Now()
	s.Sleep(5 * time.Millisecond)
	assert.GreaterOrEqual(t, time.Since(start), 5*time.Millisecond)
}

// TestDefaultFileOpener 覆盖默认文件打开器的成功与文件不存在两条路径。
func TestDefaultFileOpener(t *testing.T) {
	f, err := os.CreateTemp("", "opener-*")
	assert.NoError(t, err)
	name := f.Name()
	f.Close()
	defer os.Remove(name)

	opener := &defaultFileOpener{}
	rc, err := opener.Open(name)
	assert.NoError(t, err)
	// 返回的 ReadCloser 由调用方（ChartFileLoader）负责关闭。
	assert.NoError(t, rc.Close())

	// os.Open 失败路径
	_, err = opener.Open(filepath.Join(t.TempDir(), "missing.yaml"))
	assert.Error(t, err)
}

// TestDefaultChartLoader 覆盖默认 chart 加载器的错误分支：空目录缺 Chart.yaml、
// 归档内容非法。
func TestDefaultChartLoader(t *testing.T) {
	cl := &defaultChartLoader{}
	// 空目录缺 Chart.yaml
	_, err := cl.LoadDir(t.TempDir())
	assert.Error(t, err)
	// 非法归档内容
	_, err = cl.LoadArchive(bytes.NewReader([]byte("not a chart")))
	assert.Error(t, err)
}

// TestUserConfigLoader_Load_nonEOFError 验证非 io.EOF 的配置解析错误（非法路径分隔符）
// 会被原样上抛，而不是被当成「配置为空」静默吞掉。
func TestUserConfigLoader_Load_nonEOFError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().To(gomock.Any()).AnyTimes()
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ctx := &LoadContext{
		Input: &JobInput{
			Config: "xxx",
		},
		Messager: em,
		Logger:   mlog.NewForConfig(nil),
		Config: &mars.Config{
			IsSimpleEnv: true,
			// 非法分隔符触发 DeepSetKey 错误（非 io.EOF）
			ConfigField: "->app",
		},
	}
	assert.Error(t, (&UserConfigLoader{}).Load(ctx))
}

// TestElementsLoader_Load_defaultError 验证元素默认值非法（Switch 非 bool）时部署
// 加载直接报错——与仓库配置写入时的 ValidateElementDefault 共用同一套规则。
func TestElementsLoader_Load_defaultError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().To(gomock.Any()).AnyTimes()
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ctx := &LoadContext{
		Input:    &JobInput{},
		Messager: em,
		Config: &mars.Config{
			Elements: []*mars.Element{
				{
					Path:    "app->x",
					Type:    mars.ElementType_ElementTypeSwitch,
					Default: "not-a-bool",
				},
			},
		},
	}
	assert.Error(t, (&ElementsLoader{}).Load(ctx))
}

// TestElementsLoader_deepSetItems_error 验证非法路径的键被丢弃，不产生任何 item。
func TestElementsLoader_deepSetItems_error(t *testing.T) {
	items := (&ElementsLoader{}).deepSetItems(map[string]any{"->a": "x"})
	assert.Empty(t, items)
}

// TestSystemVariableLoader_Load_templateError 验证 values.yaml 模板 action 未闭合时
// 解析失败并返回错误。
func TestSystemVariableLoader_Load_templateError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	pl := app.NewMockPluginManager(m)
	domain := app.NewMockDomainManager(m)
	pl.EXPECT().Domain().Return(domain).AnyTimes()
	domain.EXPECT().GetDomainByIndex(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	domain.EXPECT().GetCertSecretName(gomock.Any(), gomock.Any()).AnyTimes()
	domain.EXPECT().GetClusterIssuer().Return("issuer").AnyTimes()
	em := NewMockDeployMsger(m)
	em.EXPECT().To(gomock.Any()).AnyTimes()
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ctx := &LoadContext{
		PluginMgr:        pl,
		Messager:         em,
		Logger:           mlog.NewForConfig(nil),
		Namespace:        &biz.Namespace{Name: "ns"},
		Project:          &biz.Project{Name: "app"},
		Commit:           &biz.Commit{ShortID: "short"},
		ImagePullSecrets: []string{},
		Config: &mars.Config{
			// 未闭合的模板 action 导致 Parse 报错
			ValuesYaml: "image: <.Branch",
		},
		Input: &JobInput{GitBranch: "dev", GitCommit: "c"},
		Repo:  &biz.Repo{},
	}
	assert.Error(t, (&SystemVariableLoader{}).Load(ctx))
}

// TestChartFileLoader_PackageChartError 验证远程 chart 目录打包失败时 Load 透传
// PackageChart 的原始错误。
func TestChartFileLoader_PackageChartError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any()).AnyTimes()
	h := data.NewMockHelmerRepo(m)
	gits := app.NewMockGitServer(m)
	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return([]string{"file1"}, nil)
	gits.EXPECT().GetFileContentWithSha("9999", "master", "file1").Return("file1", nil).Times(1)
	up := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().AbsolutePath(gomock.Any()).Return("/dir")
	up.EXPECT().MkDir(gomock.Any(), true).Times(1)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Times(1)
	h.EXPECT().PackageChart(gomock.Any(), gomock.Any()).Return("", errors.New("package err")).Times(1)
	pl := app.NewMockPluginManager(m)
	pl.EXPECT().Git().Return(gits).AnyTimes()
	ctx := &LoadContext{
		uploader:  up,
		Logger:    mlog.NewForConfig(nil),
		Helmer:    h,
		Input:     &JobInput{},
		Messager:  em,
		Config:    &mars.Config{LocalChartPath: "9999|master|dir"},
		PluginMgr: pl,
	}
	l := &ChartFileLoader{
		chartLoader: &fakeChartLoader{c: &chart.Chart{Metadata: &chart.Metadata{}}},
		fileOpener:  &fakeOpener{},
	}
	assert.Equal(t, "package err", l.Load(ctx).Error())
}

// failOpener 是恒失败的 chart 文件打开器替身，用于覆盖打开错误分支。
type failOpener struct{}

// Open 恒返回错误，模拟 chart 文件不可读。
func (f failOpener) Open(name string) (io.ReadCloser, error) {
	return nil, errors.New("open err")
}

// Close 空实现，满足 fileOpener 接口。
func (f failOpener) Close() error { return nil }

// TestChartFileLoader_OpenError 验证 chart 文件打开失败时 Load 透传打开错误。
func TestChartFileLoader_OpenError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any()).AnyTimes()
	h := data.NewMockHelmerRepo(m)
	gits := app.NewMockGitServer(m)
	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return([]string{"file1"}, nil)
	gits.EXPECT().GetFileContentWithSha("9999", "master", "file1").Return("file1", nil).Times(1)
	up := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().AbsolutePath(gomock.Any()).Return("/dir")
	up.EXPECT().MkDir(gomock.Any(), true).Times(1)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Times(1)
	h.EXPECT().PackageChart(gomock.Any(), gomock.Any()).Return("/app/chart.tgz", nil).Times(1)
	pl := app.NewMockPluginManager(m)
	pl.EXPECT().Git().Return(gits).AnyTimes()
	ctx := &LoadContext{
		uploader:  up,
		Logger:    mlog.NewForConfig(nil),
		Helmer:    h,
		Input:     &JobInput{},
		Messager:  em,
		Config:    &mars.Config{LocalChartPath: "9999|master|dir"},
		PluginMgr: pl,
	}
	l := &ChartFileLoader{
		chartLoader: &fakeChartLoader{c: &chart.Chart{Metadata: &chart.Metadata{}}},
		fileOpener:  failOpener{},
	}
	assert.Equal(t, "open err", l.Load(ctx).Error())
}

// TestChartFileLoader_NoFilesError 覆盖远程目录无任何文件时的错误分支：
// GetDirectoryFilesWithBranch 返回空列表直接报"charts 文件不存在"，
// 不进入打包流程（chartLoader/fileOpener/Helmer 均无需准备）。
func TestChartFileLoader_NoFilesError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any()).AnyTimes()
	gits := app.NewMockGitServer(m)
	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return([]string{}, nil)
	pl := app.NewMockPluginManager(m)
	pl.EXPECT().Git().Return(gits).AnyTimes()
	ctx := &LoadContext{
		Logger:    mlog.NewForConfig(nil),
		Messager:  em,
		Config:    &mars.Config{LocalChartPath: "9999|master|dir"},
		PluginMgr: pl,
	}
	assert.Equal(t, "charts 文件不存在: 9999|master|dir", (&ChartFileLoader{}).Load(ctx).Error())
}

// TestMergeValuesLoader_NewYAMLError 验证系统 values YAML 本身非法时，合并阶段
// 在写文件前就报错。
func TestMergeValuesLoader_NewYAMLError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	ctx := &LoadContext{
		SystemValuesYaml: "invalid: [unclosed",
		Input:            &JobInput{},
		Messager:         msger,
		Logger:           mlog.NewForConfig(nil),
		ValuesOptions:    &values.Options{},
	}
	assert.Error(t, (&MergeValuesLoader{}).Load(ctx))
}

// TestMergeValuesLoader_WriteConfigError 验证临时配置文件写入失败时 Load 透传
// uploader 的原始错误。
func TestMergeValuesLoader_WriteConfigError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	up := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil, errors.New("put err")).Times(1)
	ctx := &LoadContext{
		SystemValuesYaml: "app:\n  one: 1\n",
		Input:            &JobInput{},
		Messager:         msger,
		Logger:           mlog.NewForConfig(nil),
		timer:            timer.NewReal(),
		ValuesOptions:    &values.Options{},
		uploader:         up,
	}
	assert.Equal(t, "put err", (&MergeValuesLoader{}).Load(ctx).Error())
}

// TestMergeValuesLoader_OnFinallyClose 验证合并成功后登记清理函数：Load 返回 nil，
// 执行清理时会删除临时配置文件。
func TestMergeValuesLoader_OnFinallyClose(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	msger.EXPECT().To(gomock.Any()).AnyTimes()
	msger.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	up := uploader.NewMockUploader(m)
	finfo := uploader.NewMockFileInfo(m)
	finfo.EXPECT().Path().Return("/app/config.yaml")
	up.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(finfo, nil).Times(1)
	up.EXPECT().Delete("/app/config.yaml").Return(nil).Times(1)
	ctx := &LoadContext{
		SystemValuesYaml: "app:\n  one: 1\n",
		Input:            &JobInput{},
		Messager:         msger,
		Logger:           mlog.NewForConfig(nil),
		timer:            timer.NewReal(),
		ValuesOptions:    &values.Options{},
		uploader:         up,
	}
	assert.Nil(t, (&MergeValuesLoader{}).Load(ctx))
	assert.Equal(t, "/app/config.yaml", ctx.ValuesOptions.ValueFiles[0])

	// 触发清理，走到 closer.Close 删除临时文件
	for _, c := range ctx.cleanups {
		c()
	}
}

// TestDownloadFilesToDir 覆盖多文件下载的各类分支：git 拉取失败仍写空内容、
// Put 失败与成功、deleteFn 删除失败与成功、空文件列表。
func TestDownloadFilesToDir(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gits := app.NewMockGitServer(m)
	pl := app.NewMockPluginManager(m)
	up := uploader.NewMockUploader(m)
	pl.EXPECT().Git().Return(gits).AnyTimes()
	up.EXPECT().LocalUploader().Return(up).AnyTimes()

	ctx := &LoadContext{
		PluginMgr: pl,
		uploader:  up,
		Logger:    mlog.NewForConfig(nil),
	}

	// 文件1：Git 拉取失败 + Put 失败（两个错误分支同时命中，raw 为空仍继续 Put）
	gits.EXPECT().GetFileContentWithSha("9", "master", "a.txt").Return("", errors.New("git err")).Times(1)
	up.EXPECT().Put(filepath.Join("/dir", "a.txt"), gomock.Any()).Return(nil, errors.New("put err")).Times(1)
	// 文件2：成功路径
	gits.EXPECT().GetFileContentWithSha("9", "master", "b.txt").Return("content", nil).Times(1)
	up.EXPECT().Put(filepath.Join("/dir", "b.txt"), gomock.Any()).Return(nil, nil).Times(1)
	// deleteFn 失败路径
	up.EXPECT().DeleteDir("/dir").Return(errors.New("del err")).Times(1)

	dir, deleteFn := ctx.DownloadFilesToDir(9, "master", []string{"a.txt", "b.txt"}, "/dir")
	assert.Equal(t, "/dir", dir)
	deleteFn() // 命中 DeleteDir 错误分支

	// 空文件列表 + deleteFn 成功路径
	up.EXPECT().DeleteDir("/dir2").Return(nil).Times(1)
	dir2, deleteFn2 := ctx.DownloadFilesToDir(9, "master", nil, "/dir2")
	assert.Equal(t, "/dir2", dir2)
	deleteFn2()
}

// TestValidateElementDefault 覆盖各元素类型的默认值校验：它与 ElementsLoader.typedValue
// 共用同一套转换规则（Switch→bool、InputNumber→整数、选择类→白名单、Number*→整数），
// 确保仓库配置写入时能提前拦住「保存成功但部署必失败」的默认值。
func TestValidateElementDefault(t *testing.T) {
	tests := []struct {
		name    string
		element *mars.Element
		wantErr bool
	}{
		{name: "Switch 空默认值按 false", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeSwitch}},
		{name: "Switch 合法 bool", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeSwitch, Default: "true"}},
		{name: "Switch 非 bool", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeSwitch, Default: "yes"}, wantErr: true},
		{name: "InputNumber 空默认值按 0", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeInputNumber}},
		{name: "InputNumber 合法整数", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeInputNumber, Default: "3"}},
		{name: "InputNumber 非整数", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeInputNumber, Default: "3.5"}, wantErr: true},
		{name: "Select 默认值在白名单", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeSelect, Default: "a", SelectValues: []string{"a", "b"}}},
		{name: "Select 默认值不在白名单", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeSelect, Default: "c", SelectValues: []string{"a", "b"}}, wantErr: true},
		{name: "NumberSelect 合法整数", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeNumberSelect, Default: "2", SelectValues: []string{"1", "2"}}},
		{name: "NumberSelect 非整数", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeNumberSelect, Default: "x", SelectValues: []string{"x"}}, wantErr: true},
		{name: "Input 原样透传", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeInput, Default: "nginx:1.27"}},
		{name: "TextArea 原样透传", element: &mars.Element{Path: "a", Type: mars.ElementType_ElementTypeTextArea, Default: "k: v"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateElementDefault(tt.element)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

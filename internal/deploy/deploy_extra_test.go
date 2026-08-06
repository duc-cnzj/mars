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
	"github.com/duc-cnzj/mars/v6/internal/application"
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

type fakeSleeper struct{}

func (f fakeSleeper) Sleep(time.Duration) {}

func TestNewProcessPercent(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	msger := NewMockDeployMsger(m)
	pp := NewProcessPercent(msger, fakeSleeper{}).(*processPercent)
	assert.Equal(t, msger, pp.msger)
	assert.Equal(t, int64(0), pp.Current())
}

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

func TestRealSleeper(t *testing.T) {
	s := NewRealSleeper()
	start := time.Now()
	s.Sleep(5 * time.Millisecond)
	assert.GreaterOrEqual(t, time.Since(start), 5*time.Millisecond)
}

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

func TestDefaultChartLoader(t *testing.T) {
	cl := &defaultChartLoader{}
	// 空目录缺 Chart.yaml
	_, err := cl.LoadDir(t.TempDir())
	assert.Error(t, err)
	// 非法归档内容
	_, err = cl.LoadArchive(bytes.NewReader([]byte("not a chart")))
	assert.Error(t, err)
}

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

func TestElementsLoader_deepSetItems_error(t *testing.T) {
	items := (&ElementsLoader{}).deepSetItems(map[string]any{"->a": "x"})
	assert.Empty(t, items)
}

func TestSystemVariableLoader_Load_templateError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	pl := application.NewMockPluginManager(m)
	domain := application.NewMockDomainManager(m)
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

func TestChartFileLoader_PackageChartError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any()).AnyTimes()
	h := data.NewMockHelmerRepo(m)
	gits := application.NewMockGitServer(m)
	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return([]string{"file1"}, nil)
	gits.EXPECT().GetFileContentWithSha("9999", "master", "file1").Return("file1", nil).Times(1)
	up := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().AbsolutePath(gomock.Any()).Return("/dir")
	up.EXPECT().MkDir(gomock.Any(), true).Times(1)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Times(1)
	h.EXPECT().PackageChart(gomock.Any(), gomock.Any()).Return("", errors.New("package err")).Times(1)
	pl := application.NewMockPluginManager(m)
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

type failOpener struct{}

func (f failOpener) Open(name string) (io.ReadCloser, error) {
	return nil, errors.New("open err")
}

func (f failOpener) Close() error { return nil }

func TestChartFileLoader_OpenError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	em := NewMockDeployMsger(m)
	em.EXPECT().SendMsg(gomock.Any()).AnyTimes()
	em.EXPECT().To(gomock.Any()).AnyTimes()
	h := data.NewMockHelmerRepo(m)
	gits := application.NewMockGitServer(m)
	gits.EXPECT().GetDirectoryFilesWithBranch("9999", "master", "dir", true).Return([]string{"file1"}, nil)
	gits.EXPECT().GetFileContentWithSha("9999", "master", "file1").Return("file1", nil).Times(1)
	up := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().AbsolutePath(gomock.Any()).Return("/dir")
	up.EXPECT().MkDir(gomock.Any(), true).Times(1)
	up.EXPECT().Put(gomock.Any(), gomock.Any()).Times(1)
	h.EXPECT().PackageChart(gomock.Any(), gomock.Any()).Return("/app/chart.tgz", nil).Times(1)
	pl := application.NewMockPluginManager(m)
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

func TestDownloadFilesToDir(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	gits := application.NewMockGitServer(m)
	pl := application.NewMockPluginManager(m)
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

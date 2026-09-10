package data

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/values"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	helmtime "helm.sh/helm/v3/pkg/time"
	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// helmMemConfig 构造内存存储 + 假 kube 客户端的 helm action 配置：
// 安装/升级/状态查询的成功分支无需真实集群即可执行（假客户端 Create/Update 为 no-op）。
func helmMemConfig() *action.Configuration {
	return &action.Configuration{
		RESTClientGetter: genericclioptions.NewConfigFlags(true),
		Releases:         storage.Init(driver.NewMemory()),
		KubeClient:       &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: io.Discard}},
		Capabilities:     chartutil.DefaultCapabilities.Copy(),
		Log:              func(string, ...any) {},
	}
}

// withHelmMemConfig 把 newActionConfig 替换为返回共享内存配置的构建函数，测试结束还原。
// 注意：替换全局缝的函数不得 t.Parallel，避免与其它测试并发读写。
func withHelmMemConfig(t *testing.T, cfg *action.Configuration) {
	t.Helper()
	orig := newActionConfig
	newActionConfig = func(namespace, kubeconfig string, log func(format string, v ...any)) *action.Configuration {
		return cfg
	}
	t.Cleanup(func() { newActionConfig = orig })
}

// seedRelease 往内存存储写入一条指定状态的 release，供 History/Status 查询命中。
func seedRelease(t *testing.T, cfg *action.Configuration, name, namespace string, status release.Status) {
	t.Helper()
	require.NoError(t, cfg.Releases.Create(&release.Release{
		Name:      name,
		Namespace: namespace,
		Version:   1,
		Info: &release.Info{
			FirstDeployed: helmtime.Now(),
			LastDeployed:  helmtime.Now(),
			Status:        status,
		},
	}))
}

// Test_runInstall 覆盖 runInstall 三个分支：chart 不可安装、Devel 版本回落 + 成功安装、values 合并失败。
func Test_runInstall(t *testing.T) {
	t.Run("checkIfInstallable error", func(t *testing.T) {
		client := action.NewInstall(helmMemConfig())
		_, err := (&DefaultHelmer{logger: mlog.NewForConfig(nil)}).runInstall(
			context.TODO(), "rel", &chart.Chart{Metadata: &chart.Metadata{Type: "library"}}, client, &values.Options{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not installable")
	})

	t.Run("Devel version and install success", func(t *testing.T) {
		client := action.NewInstall(helmMemConfig())
		client.Devel = true
		rel, err := (&DefaultHelmer{logger: mlog.NewForConfig(nil)}).runInstall(
			context.TODO(), "rel", &chart.Chart{Metadata: &chart.Metadata{}}, client, &values.Options{})
		require.NoError(t, err)
		assert.NotNil(t, rel)
		assert.Equal(t, "rel", rel.Name)
	})

	t.Run("merge values error", func(t *testing.T) {
		client := action.NewInstall(helmMemConfig())
		_, err := (&DefaultHelmer{logger: mlog.NewForConfig(nil)}).runInstall(
			context.TODO(), "rel", &chart.Chart{Metadata: &chart.Metadata{}}, client,
			&values.Options{ValueFiles: []string{"/nonexistent-values.yaml"}})
		assert.Error(t, err)
	})
}

// Test_upgradeOrInstall_installPath 覆盖首次安装路径（helm.go 181-186）：
// 空存储时 History 返回 ErrReleaseNotFound → NewInstall + fillInstall + runInstall。
func Test_upgradeOrInstall_installPath(t *testing.T) {
	cfg := helmMemConfig()
	withHelmMemConfig(t, cfg)
	fn := biz.WrapLogFn(func([]*websocket_pb.Container, string, ...any) {})
	rel, err := (&DefaultHelmer{logger: mlog.NewForConfig(nil)}).upgradeOrInstall(
		context.TODO(), "rel", "ns", &chart.Chart{Metadata: &chart.Metadata{}}, &values.Options{}, fn,
		false, 0, false, nil, "desc", "/nonexistent/kubeconfig", nil)
	require.NoError(t, err)
	assert.Equal(t, "rel", rel.Name)
}

// Test_upgradeOrInstall_upgradePath 覆盖升级路径（helm.go 189-200）：release 已存在时
// 走 values 合并 + 依赖校验 + Upgrade.RunWithContext；valueOpts=nil 触发缺省回落。
func Test_upgradeOrInstall_upgradePath(t *testing.T) {
	cfg := helmMemConfig()
	seedRelease(t, cfg, "rel", "ns", release.StatusDeployed)
	withHelmMemConfig(t, cfg)
	fn := biz.WrapLogFn(func([]*websocket_pb.Container, string, ...any) {})
	rel, err := (&DefaultHelmer{logger: mlog.NewForConfig(nil)}).upgradeOrInstall(
		context.TODO(), "rel", "ns", &chart.Chart{Metadata: &chart.Metadata{}}, nil, fn,
		false, 30, false, nil, "desc", "/nonexistent/kubeconfig", nil)
	require.NoError(t, err)
	assert.NotNil(t, rel)
}

// Test_upgradeOrInstall_mergeError 覆盖 upgradeOrInstall values 合并失败分支（helm.go 190-192）。
func Test_upgradeOrInstall_mergeError(t *testing.T) {
	cfg := helmMemConfig()
	seedRelease(t, cfg, "rel", "ns", release.StatusDeployed)
	withHelmMemConfig(t, cfg)
	fn := biz.WrapLogFn(func([]*websocket_pb.Container, string, ...any) {})
	_, err := (&DefaultHelmer{logger: mlog.NewForConfig(nil)}).upgradeOrInstall(
		context.TODO(), "rel", "ns", &chart.Chart{Metadata: &chart.Metadata{}},
		&values.Options{ValueFiles: []string{"/nonexistent-values.yaml"}}, fn,
		false, 0, false, nil, "desc", "/nonexistent/kubeconfig", nil)
	assert.Error(t, err)
}

// Test_upgradeOrInstall_depsError 覆盖 chart 依赖缺失分支（helm.go 194-197）：
// CheckDependencies 报依赖在 Chart.yaml 声明但 charts/ 目录缺失。
func Test_upgradeOrInstall_depsError(t *testing.T) {
	cfg := helmMemConfig()
	seedRelease(t, cfg, "rel", "ns", release.StatusDeployed)
	withHelmMemConfig(t, cfg)
	fn := biz.WrapLogFn(func([]*websocket_pb.Container, string, ...any) {})
	ch := &chart.Chart{Metadata: &chart.Metadata{
		Dependencies: []*chart.Dependency{{Name: "dep", Version: "1.0.0"}},
	}}
	_, err := (&DefaultHelmer{logger: mlog.NewForConfig(nil)}).upgradeOrInstall(
		context.TODO(), "rel", "ns", ch, &values.Options{}, fn,
		false, 0, false, nil, "desc", "/nonexistent/kubeconfig", nil)
	assert.Error(t, err)
}

// Test_upgradeOrInstall_waitBranch 覆盖 wait && !dryRun 分支（helm.go 134-159）：
// 注册 Pod/Event 扇出监听、按 podSelectors 解析选择器并启动两个 watch goroutine。
func Test_upgradeOrInstall_waitBranch(t *testing.T) {
	cfg := helmMemConfig()
	seedRelease(t, cfg, "rel", "ns", release.StatusDeployed)
	withHelmMemConfig(t, cfg)
	fn := biz.WrapLogFn(func([]*websocket_pb.Container, string, ...any) {})
	k8sClient := &K8sClient{
		podFanOut:   newFanOut(mlog.NewForConfig(nil), "pod", make(chan Obj[*corev1.Pod]), map[string]chan<- Obj[*corev1.Pod]{}),
		eventFanOut: newFanOut(mlog.NewForConfig(nil), "event", make(chan Obj[*eventv1.Event]), map[string]chan<- Obj[*eventv1.Event]{}),
	}
	rel, err := (&DefaultHelmer{logger: mlog.NewForConfig(nil)}).upgradeOrInstall(
		context.TODO(), "rel", "ns", &chart.Chart{Metadata: &chart.Metadata{}}, &values.Options{}, fn,
		true, 30, false, []string{"app=release"}, "desc", "/nonexistent/kubeconfig", k8sClient)
	require.NoError(t, err)
	assert.NotNil(t, rel)
}

// Test_releaseStatus_Deployed 覆盖 releaseStatus 成功路径（helm.go 372-373）：
// 内存存储命中 deployed release 映射为 StatusDeployed。
func Test_releaseStatus_Deployed(t *testing.T) {
	cfg := helmMemConfig()
	seedRelease(t, cfg, "rel", "ns", release.StatusDeployed)
	withHelmMemConfig(t, cfg)
	got := (&DefaultHelmer{logger: mlog.NewForConfig(nil)}).releaseStatus("rel", "ns", "/nonexistent/kubeconfig")
	assert.Equal(t, types.Deploy_StatusDeployed, got)
}

// Test_packageChart_LoadDirError 覆盖 packageChart 加载 chart 目录失败分支（helm.go 437-439）。
func Test_packageChart_LoadDirError(t *testing.T) {
	_, err := packageChart(filepath.Join(t.TempDir(), "nonexistent"), "", config.DockerAuths{}, false)
	assert.Error(t, err)
}

// Test_packageChart_DependencyBranch 覆盖 packageChart 依赖分支（helm.go 440-460）：
// Chart.yaml 声明依赖但 charts/ 缺失 → 触发 dockerCfgOnce 写凭据 → registry 客户端创建 →
// downloadManager.Update 访问不可达仓库失败；篡改凭据文件为非法 JSON 后再次打包，
// 覆盖 registry 客户端创建失败分支（helm.go 446-448）。
func Test_packageChart_DependencyBranch(t *testing.T) {
	chartDir := t.TempDir()
	chartYaml := "apiVersion: v2\nname: test-chart\nversion: 0.1.0\ntype: application\n" +
		"dependencies:\n  - name: dep\n    version: 1.0.0\n    repository: https://example.invalid/repo\n"
	require.NoError(t, os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte(chartYaml), 0o644))

	t.Run("dependency update fails", func(t *testing.T) {
		_, err := packageChart(chartDir, "", config.DockerAuths{}, false)
		assert.Error(t, err)
	})

	t.Run("registry client error after config tamper", func(t *testing.T) {
		// dockerCfgOnce 已写入合法凭据，篡改为非法 JSON 使 registry 客户端创建失败。
		require.NoError(t, os.WriteFile(dockerCfgOncePath, []byte("not-json{{{"), 0o600))
		_, err := packageChart(chartDir, "", config.DockerAuths{}, false)
		assert.Error(t, err)
	})
}

// TestDefaultHelmer_UpgradeOrInstall_WaitProbe 覆盖 UpgradeOrInstall 的 wait 探测路径
// （helm.go 73-80）：dry-run 探测成功 → 按 manifest 提取 Pod 选择器 → 正式执行带 wait。
func TestDefaultHelmer_UpgradeOrInstall_WaitProbe(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	cfg := helmMemConfig()
	withHelmMemConfig(t, cfg)

	k8sClient := &K8sClient{
		podFanOut:   newFanOut(mlog.NewForConfig(nil), "pod", make(chan Obj[*corev1.Pod]), map[string]chan<- Obj[*corev1.Pod]{}),
		eventFanOut: newFanOut(mlog.NewForConfig(nil), "event", make(chan Obj[*eventv1.Event]), map[string]chan<- Obj[*eventv1.Event]{}),
	}
	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(k8sClient).Times(2)
	mockK8s := NewMockK8sRepo(m)
	mockK8s.EXPECT().SplitManifests(gomock.Any()).Return([]string{"m1"})
	mockK8s.EXPECT().GetPodSelectorsByManifest([]string{"m1"}).Return([]string{"app=release"})

	d := &DefaultHelmer{
		logger:     mlog.NewForConfig(nil),
		data:       mockData,
		k8sRepo:    mockK8s,
		KubeConfig: "/nonexistent/kubeconfig",
	}
	fn := biz.WrapLogFn(func([]*websocket_pb.Container, string, ...any) {})
	rel, err := d.UpgradeOrInstall(context.TODO(), "rel", "ns", &chart.Chart{Metadata: &chart.Metadata{}}, &values.Options{}, fn, true, 0, false, "desc")
	require.NoError(t, err)
	assert.NotNil(t, rel)
}

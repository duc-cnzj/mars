package data

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	entnamespace "github.com/duc-cnzj/mars/v6/internal/data/ent/namespace"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
)

func Test_dataImpl(t *testing.T) {
	d := &dataImpl{minioCli: &minio.Client{}, oidc: biz.OidcConfig{}}
	assert.NotNil(t, d.MinioCli())
	assert.NotNil(t, d.OidcConfig())
}

// TestDataImpl_AdminPassword 覆盖 admin 登录密码取数（供 biz.AuthConfigProvider 使用）。
func TestDataImpl_AdminPassword(t *testing.T) {
	d := &dataImpl{cfg: &config.Config{AdminPassword: "secret"}}
	assert.Equal(t, "secret", d.AdminPassword())
}

func Test_filterEvent(t *testing.T) {
	b := filterEvent("aaa")(&eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "aaa-bbb-ccc",
		},
		Reason: "x",
		Regarding: corev1.ObjectReference{
			Kind: "Pod",
		},
	})
	assert.True(t, b)
	b = filterEvent("aaa")(eventsv1.Event{
		Reason: "x",
		Regarding: corev1.ObjectReference{
			Kind: "Pod",
		},
	})
	assert.False(t, b)
}

func Test_filterPod(t *testing.T) {
	b := filterPod("aaa")(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "aaa-bbb-ccc",
		},
	})
	assert.True(t, b)
	b = filterPod("aaa")(corev1.Pod{})
	assert.False(t, b)
}

// 回归防护：InitDB 失败必须透出错误（修复前 once.Do 吞错，恒返 (nil, nil)），
// DB 初始化失败时调用方（DBBootstrapper）才能 fail-fast。
func TestDataImpl_InitDB_ErrorPropagated(t *testing.T) {
	d := &dataImpl{
		cfg:    &config.Config{DBDriver: "oracle"},
		logger: mlog.NewForConfig(nil),
	}
	closeFunc, err := d.InitDB()
	assert.Nil(t, closeFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported database driver oracle")
}

// 成功路径：closeFunc 非 nil 且可关闭，DB 客户端已就位。
func TestDataImpl_InitDB_Success(t *testing.T) {
	d := &dataImpl{
		cfg:    &config.Config{DBDriver: "sqlite", DBDatabase: ":memory:"},
		logger: mlog.NewForConfig(nil),
	}
	closeFunc, err := d.InitDB()
	assert.NoError(t, err)
	assert.NotNil(t, closeFunc)
	assert.NotNil(t, d.DB())
	assert.NoError(t, closeFunc())
}

// DBDebug=true 时经 .Debug() 激活 ent 调试日志——覆盖 data.go InitDB 的
// DBDebug 分支，且后续真实查询会触发 database.go InitDB 的 ent.Log 闭包体。
func TestDataImpl_InitDB_DBDebug(t *testing.T) {
	d := &dataImpl{
		cfg:    &config.Config{DBDriver: "sqlite", DBDatabase: filepath.Join(t.TempDir(), "dbg.db"), DBDebug: true},
		logger: mlog.NewForConfig(nil),
	}
	closeFunc, err := d.InitDB()
	assert.NoError(t, err)
	assert.NotNil(t, closeFunc)
	assert.NotNil(t, d.DB())
	// 执行真实 SQL，激活 ent.Log → logger.Debug 闭包
	assert.NoError(t, d.DB().Schema.Create(context.TODO()))
	assert.NoError(t, closeFunc())
}

// Migrate 收敛到 Data 门面：对已初始化的 ent 客户端执行 schema 迁移且幂等。
// 用独立的 shared-memory DB 而非包级 testDB 单例——Migrate 带 WithDropColumn/
// WithDropIndex，复用共享库会干扰其它测试的表结构（顺序依赖）。
func TestDataImpl_Migrate(t *testing.T) {
	client, err := ent.Open("sqlite3", "file:test_migrate?mode=memory&cache=shared&_fk=1&loc=Local")
	assert.NoError(t, err)
	defer client.Close()
	assert.NoError(t, client.Schema.Create(context.TODO()))

	d := &dataImpl{db: client}
	assert.NoError(t, d.Migrate())
}

// TestDataImpl_InitS3 覆盖 S3 初始化三分支：禁用早退 / 配置缺失报错 / 正常建客户端。
func TestDataImpl_InitS3(t *testing.T) {
	t.Run("disabled skips", func(t *testing.T) {
		d := &dataImpl{cfg: &config.Config{}, logger: mlog.NewForConfig(nil)}
		assert.NoError(t, d.InitS3())
		assert.Nil(t, d.MinioCli())
	})

	t.Run("missing config errors", func(t *testing.T) {
		d := &dataImpl{cfg: &config.Config{S3Enabled: true, S3Endpoint: "", S3AccessKeyID: "", S3SecretAccessKey: ""}, logger: mlog.NewForConfig(nil)}
		err := d.InitS3()
		assert.ErrorContains(t, err, "s3 config error")
	})

	t.Run("happy path", func(t *testing.T) {
		d := &dataImpl{cfg: &config.Config{S3Enabled: true, S3Endpoint: "127.0.0.1:9000", S3AccessKeyID: "ak", S3SecretAccessKey: "sk"}, logger: mlog.NewForConfig(nil)}
		assert.NoError(t, d.InitS3())
		assert.NotNil(t, d.MinioCli())
	})

	t.Run("minio client construction fails", func(t *testing.T) {
		// endpoint 非空但无法解析（scheme 后接非法路径）→ minio.New 构造失败。
		d := &dataImpl{cfg: &config.Config{S3Enabled: true, S3Endpoint: "://bad", S3AccessKeyID: "ak", S3SecretAccessKey: "sk"}, logger: mlog.NewForConfig(nil)}
		err := d.InitS3()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "init s3")
		assert.Nil(t, d.MinioCli())
	})
}

// oidcDiscovery 返回一个最小可用的 OIDC 发现文档（含 extraValues 字段），
// 供 oidc.NewProvider 与 provider.Claims 使用。withScopes 控制 addOidcCfg 的 scopes 回退分支。
func oidcDiscovery(issuer string, withScopes bool) string {
	scopes := `"scopes_supported":["openid","profile"],`
	if !withScopes {
		scopes = ""
	}
	return fmt.Sprintf(`{
		"issuer": %q,
		"jwks_uri": %q,
		"authorization_endpoint": %q,
		"token_endpoint": %q,
		"response_types_supported": ["code"],
		"subject_types_supported": ["public"],
		"id_token_signing_alg_values_supported": ["RS256"],
		%[5]s
		"check_session_iframe": %q,
		"end_session_endpoint": %q
	}`, issuer, issuer+"/keys", issuer+"/auth", issuer+"/token", scopes, issuer+"/session", issuer+"/logout")
}

// newOidcServer 构造带发现文档的 OIDC 假服务，返回 server。
func newOidcServer(t *testing.T, withScopes bool) *httptest.Server {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, oidcDiscovery(srv.URL, withScopes))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestDataImpl_InitOidcProvider 覆盖 OIDC provider 装配：全禁用 / 启用成功 / provider 拉取失败。
func TestDataImpl_InitOidcProvider(t *testing.T) {
	t.Run("all disabled leaves empty", func(t *testing.T) {
		d := &dataImpl{
			cfg:    &config.Config{Oidc: []config.OidcSetting{{Name: "a", Enabled: false}}},
			logger: mlog.NewForConfig(nil),
		}
		d.InitOidcProvider()
		assert.Empty(t, d.OidcConfig())
	})

	t.Run("enabled happy path", func(t *testing.T) {
		srv := newOidcServer(t, true)
		d := &dataImpl{
			cfg: &config.Config{Oidc: []config.OidcSetting{{
				Name: "ali", Enabled: true, ProviderUrl: srv.URL, ClientID: "cid", ClientSecret: "cs", RedirectUrl: "http://cb",
			}}},
			logger: mlog.NewForConfig(nil),
		}
		d.InitOidcProvider()
		item, ok := d.OidcConfig()["ali"]
		require.True(t, ok)
		assert.Equal(t, "cid", item.Config.ClientID)
		assert.Equal(t, "http://cb", item.Config.RedirectURL)
		assert.Equal(t, srv.URL+"/logout", item.EndSessionEndpoint)
	})

	t.Run("provider fetch fails leaves empty", func(t *testing.T) {
		d := &dataImpl{
			cfg:    &config.Config{Oidc: []config.OidcSetting{{Name: "bad", Enabled: true, ProviderUrl: "http://127.0.0.1:1/x"}}},
			logger: mlog.NewForConfig(nil),
		}
		d.InitOidcProvider()
		assert.Empty(t, d.OidcConfig())
	})

	t.Run("claims unmarshal type-conflict leaves empty", func(t *testing.T) {
		// oidc.NewProvider 缓存发现文档，Claims 不再二次 HTTP。用类型冲突 doc：
		// scopes_supported 给数字——NewProvider 的 providerJSON 不用该字段而忽略，
		// Claims 反序列化 extraValues（[]string 型）时必失败。
		var srv *httptest.Server
		srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"issuer": %q,
				"jwks_uri": %q,
				"authorization_endpoint": %q,
				"token_endpoint": %q,
				"response_types_supported": ["code"],
				"subject_types_supported": ["public"],
				"id_token_signing_alg_values_supported": ["RS256"],
				"scopes_supported": 123,
				"end_session_endpoint": %q
			}`, srv.URL, srv.URL+"/keys", srv.URL+"/auth", srv.URL+"/token", srv.URL+"/logout")
		}))
		t.Cleanup(srv.Close)

		d := &dataImpl{
			cfg:    &config.Config{Oidc: []config.OidcSetting{{Name: "claims-bad", Enabled: true, ProviderUrl: srv.URL}}},
			logger: mlog.NewForConfig(nil),
		}
		d.InitOidcProvider()
		assert.Empty(t, d.OidcConfig())
	})

	t.Run("one provider fails others still configured", func(t *testing.T) {
		// P1 回归：坏 provider 失败只跳过自身，好的 provider 仍被装配
		// （修复前 return 直接退出整个闭包，后续 provider 全部中断）。
		srv := newOidcServer(t, true)
		d := &dataImpl{
			cfg: &config.Config{Oidc: []config.OidcSetting{
				{Name: "bad", Enabled: true, ProviderUrl: "http://127.0.0.1:1/x"},
				{Name: "good", Enabled: true, ProviderUrl: srv.URL, ClientID: "cid", ClientSecret: "cs", RedirectUrl: "http://cb"},
			}},
			logger: mlog.NewForConfig(nil),
		}
		d.InitOidcProvider()
		item, ok := d.OidcConfig()["good"]
		require.True(t, ok, "坏 provider 不应阻断其余 provider 装配")
		assert.Equal(t, "cid", item.Config.ClientID)
		assert.Equal(t, "http://cb", item.Config.RedirectURL)
		_, badOk := d.OidcConfig()["bad"]
		assert.False(t, badOk, "失败的 provider 不应进入配置")
	})
}

// TestDataImpl_InitK8s_ErrorBranches 覆盖 InitK8s 两个错误分支：
// kubeconfig 路径无效 / 集群外无 InClusterConfig。happy path 需真实集群（informer 同步），
// 属集成边界，不在单测范围。
func TestDataImpl_InitK8s_ErrorBranches(t *testing.T) {
	t.Run("invalid kubeconfig path", func(t *testing.T) {
		d := &dataImpl{
			cfg:    &config.Config{KubeConfig: "/nonexistent/kubeconfig"},
			logger: mlog.NewForConfig(nil),
		}
		err := d.InitK8s(make(chan struct{}))
		assert.Error(t, err)
		assert.Nil(t, d.K8s())
	})

	t.Run("no in-cluster config", func(t *testing.T) {
		d := &dataImpl{
			cfg:    &config.Config{KubeConfig: ""},
			logger: mlog.NewForConfig(nil),
		}
		err := d.InitK8s(make(chan struct{}))
		assert.Error(t, err)
		assert.Nil(t, d.K8s())
	})
}

// TestAddOidcCfg 覆盖 scopes 缺省回退分支（发现文档不含 scopes_supported → ScopeOpenID）。
func TestAddOidcCfg(t *testing.T) {
	srv := newOidcServer(t, false)
	provider, err := oidc.NewProvider(context.TODO(), srv.URL)
	require.NoError(t, err)

	var ev extraValues
	require.NoError(t, provider.Claims(&ev))

	m := make(biz.OidcConfig)
	addOidcCfg(provider, ev, config.OidcSetting{Name: "n", ClientID: "c", RedirectUrl: "http://r"}, m)
	item, ok := m["n"]
	require.True(t, ok)
	assert.Equal(t, []string{oidc.ScopeOpenID}, item.Config.Scopes)
}

// TestDataImpl_WithTx 覆盖事务成功提交 / fn 出错回滚 / fn panic 回滚并重抛 / 事务创建失败。
func TestDataImpl_WithTx(t *testing.T) {
	entdb, err := NewSqliteDB()
	require.NoError(t, err)
	t.Cleanup(func() { entdb.Close() })
	d := &dataImpl{db: entdb, logger: mlog.NewForConfig(nil)}

	t.Run("commit on nil fn error", func(t *testing.T) {
		err := d.WithTx(context.TODO(), func(tx *ent.Tx) error {
			return tx.Namespace.Create().SetName("tx-commit").SetCreatorEmail("e@x.y").Exec(context.TODO())
		})
		assert.NoError(t, err)
		exists, err := entdb.Namespace.Query().Where(entnamespace.Name("tx-commit")).Exist(context.TODO())
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("rollback on fn error", func(t *testing.T) {
		before, _ := entdb.Namespace.Query().Count(context.TODO())
		err := d.WithTx(context.TODO(), func(tx *ent.Tx) error {
			assert.NoError(t, tx.Namespace.Create().SetName("tx-rollback").SetCreatorEmail("e@x.y").Exec(context.TODO()))
			return fmt.Errorf("boom")
		})
		assert.Error(t, err)
		after, _ := entdb.Namespace.Query().Count(context.TODO())
		assert.Equal(t, before, after)
	})

	t.Run("repanic after rollback", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = d.WithTx(context.TODO(), func(tx *ent.Tx) error {
				panic("oops")
			})
		})
	})

	t.Run("rollback error wraps fn error", func(t *testing.T) {
		// fn 内部已提交事务后再返回错误 → 外层 Rollback 必失败（已提交），走回滚错误分支
		err := d.WithTx(context.TODO(), func(tx *ent.Tx) error {
			_ = tx.Commit()
			return fmt.Errorf("boom")
		})
		assert.ErrorContains(t, err, "rolling back transaction")
	})

	t.Run("commit error wrapped", func(t *testing.T) {
		// fn 内部已提交事务后返回 nil → 外层二次 Commit 必失败（已提交），走提交错误分支
		err := d.WithTx(context.TODO(), func(tx *ent.Tx) error {
			return tx.Commit()
		})
		assert.ErrorContains(t, err, "committing transaction")
	})

	t.Run("Tx creation error propagated", func(t *testing.T) {
		closed := &dataImpl{db: mustClosedDB(t), logger: mlog.NewForConfig(nil)}
		err := closed.WithTx(context.TODO(), func(tx *ent.Tx) error { return nil })
		assert.Error(t, err)
	})
}

// writeTestKubeconfig 生成指向 serverURL 的最小 kubeconfig 文件，供 InitK8s 经
// clientcmd.BuildConfigFromFlags 解析出合法 rest 配置。
func writeTestKubeconfig(t *testing.T, serverURL string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kubeconfig")
	content := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: test
  cluster:
    server: %s
contexts:
- name: test
  context:
    cluster: test
    user: test
current-context: test
users:
- name: test
  user: {}
`, serverURL)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// waitPod 等待 pod 扇出监听通道在超时内收到指定类型事件，否则失败。
func waitPod(t *testing.T, ch <-chan Obj[*corev1.Pod], want fanOutType) Obj[*corev1.Pod] {
	t.Helper()
	select {
	case obj := <-ch:
		assert.Equal(t, want, obj.Type())
		return obj
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for pod fanOut event")
		return nil
	}
}

// waitEvent 等待 event 扇出监听通道在超时内收到指定类型事件，否则失败。
func waitEvent(t *testing.T, ch <-chan Obj[*eventsv1.Event], want fanOutType) Obj[*eventsv1.Event] {
	t.Helper()
	select {
	case obj := <-ch:
		assert.Equal(t, want, obj.Type())
		return obj
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for event fanOut event")
		return nil
	}
}

// TestDataImpl_InitK8s_HappyPath 覆盖 InitK8s 成功路径：newK8sClientset 缝注入
// fake clientset（内存 list/watch 供 Pod/Secret informer 同步），httptest 假 API
// 只响应 CRD list（含 httproutes → 触发 gateway 分支）。断言 listers 装配后向 fake
// 注入 Pod/Event 变更，验证 informer 回调→扇出广播链路；gateway informer 对假 API
// 的 404 触发 per-informer WatchErrorHandler 闭包。
func TestDataImpl_InitK8s_HappyPath(t *testing.T) {
	fk := kubefake.NewSimpleClientset()
	origClientset := newK8sClientset
	newK8sClientset = func(_ *restclient.Config) (kubernetes.Interface, error) { return fk, nil }
	t.Cleanup(func() {
		newK8sClientset = origClientset
	})

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "customresourcedefinitions"):
			w.Header().Set("Content-Type", "application/json")
			// 首个 CRD 不匹配 → for 循环 if 假分支；第二个匹配 → gwinstalled=true + break。
			fmt.Fprint(w, `{"apiVersion":"apiextensions.k8s.io/v1","kind":"CustomResourceDefinitionList","metadata":{"resourceVersion":"1"},"items":[{"metadata":{"name":"ingresses.networking.k8s.io"}},{"metadata":{"name":"httproutes.gateway.networking.k8s.io"}}]}`)
		default:
			// gateway informer 的 HTTPRoutes list 落 404 → reflector 报错 → WatchErrorHandler 闭包。
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer api.Close()

	done := make(chan struct{})
	t.Cleanup(func() { close(done) })

	d := &dataImpl{
		cfg:    &config.Config{KubeConfig: writeTestKubeconfig(t, api.URL), NsPrefix: "test"},
		logger: mlog.NewForConfig(nil),
	}
	require.NoError(t, d.InitK8s(done))
	k := d.K8s()
	require.NotNil(t, k)
	assert.True(t, k.GatewayApiInstalled)
	assert.NotNil(t, k.PodLister)
	assert.NotNil(t, k.SecretLister)
	assert.NotNil(t, k.HTTPRouteLister)

	// 注册监听者，验证 informer 回调 → 扇出广播链路（先注册再注入，避免事件漏听）。
	podCh := make(chan Obj[*corev1.Pod], 8)
	k.podFanOut.AddListener("test-pod", podCh)
	evCh := make(chan Obj[*eventsv1.Event], 8)
	k.eventFanOut.AddListener("test-ev", evCh)

	ctx := context.TODO()
	// Pod Add（namespace 前缀匹配 filterPod("test")）
	_, err := k.Client.CoreV1().Pods("test-ns").Create(ctx, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "test-ns", ResourceVersion: "1"}}, metav1.CreateOptions{})
	require.NoError(t, err)
	assert.Equal(t, "p1", waitPod(t, podCh, Add).Current().Name)
	// Pod Update（RV 变更才过 UpdateFunc 的守卫）
	_, err = k.Client.CoreV1().Pods("test-ns").Update(ctx, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "test-ns", ResourceVersion: "2"}}, metav1.UpdateOptions{})
	require.NoError(t, err)
	assert.Equal(t, "p1", waitPod(t, podCh, Update).Current().Name)
	// Pod Delete
	require.NoError(t, k.Client.CoreV1().Pods("test-ns").Delete(ctx, "p1", metav1.DeleteOptions{}))
	assert.Equal(t, "p1", waitPod(t, podCh, Delete).Current().Name)
	// Event Add（Regarding.Kind=Pod、Reason 非 Unhealthy 才过 filterEvent("test")）
	_, err = k.Client.EventsV1().Events("test-ns").Create(ctx, &eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: "e1", Namespace: "test-ns"},
		Regarding:  corev1.ObjectReference{Kind: "Pod"},
		Reason:     "Started",
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	assert.Equal(t, "e1", waitEvent(t, evCh, Add).Current().Name)

	// 留出时间让 gateway informer 对 404 假 API 报错，执行 per-informer WatchErrorHandler 闭包。
	time.Sleep(300 * time.Millisecond)
}

// Test_newK8sClientset 覆盖构造缝默认实现：合法配置下返回非空 kubernetes 客户端
// （构造不做网络请求，仅校验配置）。
func Test_newK8sClientset(t *testing.T) {
	client, err := newK8sClientset(&restclient.Config{Host: "https://127.0.0.1:1", Timeout: time.Second})
	require.NoError(t, err)
	assert.NotNil(t, client)
}

// TestDataImpl_InitK8s_ClientsetError 覆盖 newK8sClientset 构造失败分支（data.go 300-303）：
// 缝返回错误，InitK8s 早退且不装配 K8sClient。
func TestDataImpl_InitK8s_ClientsetError(t *testing.T) {
	origClientset := newK8sClientset
	newK8sClientset = func(_ *restclient.Config) (kubernetes.Interface, error) {
		return nil, errors.New("clientset boom")
	}
	t.Cleanup(func() {
		newK8sClientset = origClientset
	})

	d := &dataImpl{
		cfg:    &config.Config{KubeConfig: writeTestKubeconfig(t, "https://127.0.0.1:1")},
		logger: mlog.NewForConfig(nil),
	}
	err := d.InitK8s(make(chan struct{}))
	assert.ErrorContains(t, err, "new k8s clientset")
	assert.Nil(t, d.K8s())
}

// TestDataImpl_InitK8s_CrdListError 覆盖 CRD list 失败分支（data.go 313-316）：
// 假 API 对 CRD 请求返回 500。
func TestDataImpl_InitK8s_CrdListError(t *testing.T) {
	fk := kubefake.NewSimpleClientset()
	origClientset := newK8sClientset
	newK8sClientset = func(_ *restclient.Config) (kubernetes.Interface, error) { return fk, nil }
	t.Cleanup(func() {
		newK8sClientset = origClientset
	})

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer api.Close()

	d := &dataImpl{
		cfg:    &config.Config{KubeConfig: writeTestKubeconfig(t, api.URL)},
		logger: mlog.NewForConfig(nil),
	}
	err := d.InitK8s(make(chan struct{}))
	assert.ErrorContains(t, err, "list gateway crds")
	assert.Nil(t, d.K8s())
}

// Test_sendOrDrop 覆盖事件投递二分：通道有空位 → 投递成功；通道已满 → 走 default 丢弃。
func Test_sendOrDrop(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	pod := newObj[*corev1.Pod](nil, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p"}}, Add)

	ch := make(chan Obj[*corev1.Pod], 1)
	sendOrDrop(ch, pod, logger, "podFanOutObj")
	assert.Len(t, ch, 1)
	// 缓冲满 → default 分支丢弃，通道仍只有 1 个。
	sendOrDrop(ch, pod, logger, "podFanOutObj")
	assert.Len(t, ch, 1)
	assert.Same(t, pod, <-ch)
}

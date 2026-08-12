// Package wssendertest 提供 wssender 三个后端包（memory/redis/nsq）测试共享的脚手架。
//
// 三个兄弟包此前各自复制了一份测试夹具（fakeApp stub、内存 ent DB、ProjectRepo、
// 项目种子、pod 事件消息与 Pod 构造），且副本已出现漂移（memory 的 seedProject 只返回
// pid、testMsg 缺 Result 字段）。此包收敛为单一事实来源，统一签名与字段，避免副本再演化。
package wssendertest

import (
	"context"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/go-redis/redis/v8"
	gonsq "github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// redisAddrOnce 保证整包只探测一次真实 redis，结果缓存复用。
var (
	redisAddrOnce sync.Once
	redisAddrVal  string
	redisAddrOK   bool
)

// RedisAddr 返回可用的真实 Redis 地址；不可用则 t.Skip。
// 地址取自 REDIS_HOST/REDIS_PORT 环境变量（缺省 127.0.0.1:6379），与 CI test.yaml 的
// redis service 及 dev/docker-compose.yml 对齐。首次调用探测一次，连接失败整包跳过集成测试。
func RedisAddr(t testing.TB) string {
	t.Helper()
	redisAddrOnce.Do(func() {
		host := os.Getenv("REDIS_HOST")
		if host == "" {
			host = "127.0.0.1"
		}
		port := os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379"
		}
		redisAddrVal = net.JoinHostPort(host, port)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		rdb := redis.NewClient(&redis.Options{Addr: redisAddrVal})
		defer func() { _ = rdb.Close() }()
		redisAddrOK = rdb.Ping(ctx).Err() == nil
	})
	if !redisAddrOK {
		t.Skipf("redis 不可用 (%s)，跳过集成测试；docker compose up -d redis 后重跑", redisAddrVal)
	}
	return redisAddrVal
}

// nsqdAddrOnce 保证整包只探测一次真实 nsqd，结果缓存复用。
var (
	nsqdAddrOnce sync.Once
	nsqdAddrVal  string
	nsqdAddrOK   bool
)

// NSQDAddr 返回可用的真实 nsqd 地址；不可用则 t.Skip。
// 地址取自 NSQ_ADDR（缺省 127.0.0.1:4150），与 CI test.yaml 的 nsqd 及 dev/docker-compose.yml 对齐。
// 探测用 go-nsq producer Ping。集成测试只直连 nsqd、不依赖 nsqlookupd（裸 lookupd 无 producer 注册，
// 见 nsq.go Initialize 注释），lookupd 分支由死端口失败测试覆盖。
func NSQDAddr(t testing.TB) string {
	t.Helper()
	nsqdAddrOnce.Do(func() {
		nsqdAddrVal = os.Getenv("NSQ_ADDR")
		if nsqdAddrVal == "" {
			nsqdAddrVal = "127.0.0.1:4150"
		}
		if p, err := gonsq.NewProducer(nsqdAddrVal, gonsq.NewConfig()); err == nil {
			defer p.Stop()
			nsqdAddrOK = p.Ping() == nil
		}
	})
	if !nsqdAddrOK {
		t.Skipf("nsqd 不可用 (%s)，跳过集成测试；docker compose up -d nsqd nsqlookupd 后重跑", nsqdAddrVal)
	}
	return nsqdAddrVal
}

// FakeApp 是 app.PluginApp 的最小 stub，供各后端 Initialize 测试使用。
// 字段命名 Repo/Log 避免与方法名（ProjectRepo/Logger）冲突。
type FakeApp struct {
	Repo biz.ProjectRepo
	Log  mlog.Logger
}

var _ app.PluginApp = FakeApp{}

// Logger 返回日志器。
func (f FakeApp) Logger() mlog.Logger { return f.Log }

// ProjectRepo 返回项目仓库（wssender 插件按项目路由命名空间用）。
func (f FakeApp) ProjectRepo() biz.ProjectRepo { return f.Repo }

// NewDB 打开一个内存 sqlite ent 客户端，并在测试结束时关闭。
func NewDB(t *testing.T) *ent.Client {
	t.Helper()
	db, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&loc=Local")
	require.NoError(t, err)
	require.NoError(t, db.Schema.Create(context.TODO()))
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// NewTestRepo 构造绑定指定 ent DB 的真实 ProjectRepo。
func NewTestRepo(t *testing.T, db *ent.Client) biz.ProjectRepo {
	t.Helper()
	impl := data.NewDataImpl(&data.NewDataParams{Cfg: &config.Config{}, DB: db})
	return data.NewProjectRepo(mlog.NewForConfig(nil), impl)
}

// SeedProject 创建 namespace + project 并返回 nsID、projectID。
func SeedProject(t *testing.T, db *ent.Client, selectors []string) (nsID, pid int) {
	t.Helper()
	ns := db.Namespace.Create().SetName("devops-test").SetCreatorEmail("a@b.c").SaveX(context.TODO())
	proj := db.Project.Create().SetName("my-app").SetCreator("tester").
		SetNamespaceID(ns.ID).SetPodSelectors(selectors).SaveX(context.TODO())
	return ns.ID, proj.ID
}

// TestMsg 构造一个最小可序列化的 pod 事件消息（To_ToAll，ProjectId 42）。
func TestMsg() *websocket_pb.WsProjectPodEventResponse {
	return &websocket_pb.WsProjectPodEventResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     "",
			Type:   websocket_pb.Type_ProjectPodEvent,
			End:    true,
			Result: websocket_pb.ResultType_Success,
			To:     websocket_pb.To_ToAll,
		},
		ProjectId: 42,
	}
}

// TestPod 构造带选择器匹配标签的测试 Pod。
func TestPod() *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "devops-test",
			Labels:    map[string]string{"app": "test"},
		},
	}
}

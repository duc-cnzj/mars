package metrics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedFamilies 是 NewRegistry 中注册的全部应用指标家族名。
// 与 NewRegistry 的 MustRegister 列表一一对应，漏注册任何一个都会让本测试失败。
var expectedFamilies = []string{
	"mars_websocket_connections",
	"mars_bootstrapper_duration_seconds",
	"mars_grpc_duration_seconds",
	"mars_grpc_errors_total",
	"mars_grpc_request_total",
	"mars_fanout_channel_length",
	"mars_websocket_request_duration_seconds",
	"mars_websocket_request_panic_total",
	"mars_websocket_request_total",
	"mars_cache_bytes",
	"mars_cache_remember_duration_seconds",
	"mars_cron_panic_total",
	"mars_cron_duration_seconds",
	"mars_cron_command_total",
	"mars_cron_error_total",
	"mars_fanout_listener_count",
}

// TestNewRegistry 验证 NewRegistry 返回的 registry 能真实输出全部应用指标：
// 先给每个指标打一个样本值（确保系列存在），再 Gather 抓取，
// 断言 16 个指标家族全部出现且 Help 文案非空。
// 覆盖的是"注册了且可收集"这一完整链路，而不是像 assert.NotNil 那样只验证对象非空。
func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	// 给每个指标写一个样本，保证 Gather 时系列必然存在。
	BootstrapperStartMetrics.WithLabelValues("boot").Set(1)
	WebsocketConnectionsCount.WithLabelValues("user").Set(1)
	GrpcLatency.WithLabelValues("/mars.Mars/Test").Observe(0.1)
	GrpcRequestTotal.WithLabelValues("/mars.Mars/Test", "success").Inc()
	GrpcErrorCount.WithLabelValues("/mars.Mars/Test").Inc()
	WebsocketRequestLatency.WithLabelValues("ws_method").Observe(0.1)
	WebsocketPanicCount.WithLabelValues("ws_method").Inc()
	WebsocketRequestTotal.WithLabelValues("ws_method", "success").Inc()
	CacheBytesGauge.WithLabelValues("cache_key").Set(1)
	CacheRememberDuration.WithLabelValues("cache_key").Observe(0.1)
	CronPanicCount.WithLabelValues("cron_name").Inc()
	CronErrorCount.WithLabelValues("cron_name").Inc()
	CronCommandCount.WithLabelValues("cron_name").Inc()
	CronDuration.WithLabelValues("cron_name").Observe(1)
	K8sInformerFanOutListenerCount.WithLabelValues("pod").Set(1)
	FanOutChannelLength.WithLabelValues("fanout").Set(0)

	families, err := registry.Gather()
	require.NoError(t, err)

	// 只收 mars_ 前缀的应用指标，与 go_*/process_* 运行时收集器区分开。
	got := make(map[string]string, len(families))
	for _, family := range families {
		if !strings.HasPrefix(family.GetName(), "mars_") {
			continue
		}
		got[family.GetName()] = family.GetHelp()
	}

	// 双向精确匹配：既抓"注册了但漏进期望清单"（反向盲区），
	// 也抓"期望清单写了但没注册"——任何一边漏都会让测试失败。
	gotNames := make([]string, 0, len(got))
	for name := range got {
		gotNames = append(gotNames, name)
	}
	assert.ElementsMatch(t, expectedFamilies, gotNames)

	for _, name := range expectedFamilies {
		// Prometheus 要求指标必须有 Help；空 Help 会让 promtool check metrics 报错，
		// 属于 S 级"注释正确"在指标层的外溢要求。
		assert.NotEmpty(t, got[name], "指标 %s 缺少 Help 文案", name)
	}
}

package mlog

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/config"
	pkgErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

// TestNewForConfigWrapped 断言 NewForConfig 返回的实例总是被 logWrapper 包裹，
// 覆盖 zap/logrus/未知 channel 回落/nil 配置四条分支。
func TestNewForConfigWrapped(t *testing.T) {
	cfgs := []*config.Config{
		{LogChannel: "zap", Debug: true},
		{LogChannel: "logrus", Debug: true},
		{LogChannel: "unknown", Debug: true},
		nil,
	}
	for _, cfg := range cfgs {
		_, ok := NewForConfig(cfg).(*logWrapper)
		assert.True(t, ok, "NewForConfig 返回的实例应被 logWrapper 包裹")
	}
}

// TestNewForConfigZapChannel 断言 zap channel 时内层后端是 zap。
func TestNewForConfigZapChannel(t *testing.T) {
	inner := NewForConfig(&config.Config{LogChannel: "zap", Debug: true}).(*logWrapper).Logger
	_, ok := inner.(*zapLogger)
	assert.True(t, ok)
}

// TestNewForConfigLogrusChannel 断言 logrus channel 时内层后端是 logrus。
func TestNewForConfigLogrusChannel(t *testing.T) {
	inner := NewForConfig(&config.Config{LogChannel: "logrus", Debug: true}).(*logWrapper).Logger
	_, ok := inner.(*logrusLogger)
	assert.True(t, ok)
}

// TestNewForConfigDefaultChannel 断言未知 channel 回落 logrus 默认后端。
func TestNewForConfigDefaultChannel(t *testing.T) {
	inner := NewForConfig(&config.Config{LogChannel: "unknown", Debug: true}).(*logWrapper).Logger
	_, ok := inner.(*logrusLogger)
	assert.True(t, ok)
}

// TestNewForConfigNilConfig 断言 nil 配置走 logrus 默认后端且保持包裹。
func TestNewForConfigNilConfig(t *testing.T) {
	inner := NewForConfig(nil).(*logWrapper).Logger
	_, ok := inner.(*logrusLogger)
	assert.True(t, ok)
}

// TestNewForConfigErrorStack 端到端断言：NewForConfig 返回的包裹实例
// 打 pkg/errors 栈错误时，输出包含完整起源栈（%+v 展开）。
// 断言分两层：消息链（root/outer）是 err.Error() 也能满足的弱断言；
// 调用函数名 TestNewForConfigErrorStack 只出现在 %+v 栈帧（zap 的 file 字段
// 是 文件:行、不含函数名）——去掉 formatError 的 fmt.Formatter 分支后此断言
// 必失败，锁死"错误日志带完整栈"的回归。
func TestNewForConfigErrorStack(t *testing.T) {
	out := captureZapOutput(t, func(logger Logger) {
		NewForConfig(&config.Config{LogChannel: "zap", Debug: true}).
			Error("boom", pkgErrors.Wrap(errors.New("root"), "outer"))
	})
	assert.Contains(t, out, "root")
	assert.Contains(t, out, "outer")
	assert.Contains(t, out, "TestNewForConfigErrorStack", "%+v 完整起源栈应含调用函数名，而非仅 Error() 消息")
}

// captureOutput 在 logger 构造窗口内重定向 stdout/stderr，返回捕获的日志输出。
// mk 在重定向窗口内构造 logger（构造时固化 output fd），说明：zap 生产配置写
// stderr、logrus 写 stdout，故两者同时重定向。
func captureOutput(t *testing.T, mk func() Logger, fn func(logger Logger)) string {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr
	fn(mk())
	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	stdout, _ := io.ReadAll(rOut)
	stderr, _ := io.ReadAll(rErr)
	return string(stdout) + string(stderr)
}

// captureZapOutput 便捷版：用 zap 生产 logger 捕获。
func captureZapOutput(t *testing.T, fn func(logger Logger)) string {
	t.Helper()
	return captureOutput(t, func() Logger { return NewZapLogger(false) }, fn)
}

// 回归防护：wrapper 包住 logger 后，file 字段必须指向真实调用点（本测试文件），
// 而不是 wrapper 内部（mlog.go）。去掉 NewLogWrapper 里的
// CallerSkipAdjuster 补偿，这两个子测试必须失败。
func TestNewLogWrapper_CallerPointsToCallSite(t *testing.T) {
	for name, mk := range map[string]func() Logger{
		"zap":    func() Logger { return NewZapLogger(false) },
		"logrus": func() Logger { return NewLogrusLogger(false) },
	} {
		t.Run(name, func(t *testing.T) {
			out := captureOutput(t, mk, func(logger Logger) {
				wrapped := NewLogWrapper(logger)
				wrapped.Error("boom", pkgErrors.Wrap(errors.New("root"), "outer"))
			})
			assert.Contains(t, out, "mlog_test.go", "file 字段应指向真实调用点文件")
			assert.NotContains(t, out, "mlog/mlog.go", "file 字段不应指向 wrapper 内部")
		})
	}
}

// logWrapper 实现 mlog.CallerSkipAdjuster：WithCallerSkip 透传给内层，供收敛
// helper（如 services.logError）补偿自身引入的调用帧。模拟生产三层链路：
// 真实调用点 → helper（多一层）→ logWrapper。无补偿时 caller 落在 helper 帧，
// 补偿 1 帧后越过 helper 指向真实调用点（两次 caller 落点不同）。
func Test_logWrapper_WithCallerSkip(t *testing.T) {
	z, buf := newTestZap(t)
	wrapped := NewLogWrapper(z)
	logAt := func(skip bool) string {
		buf.Reset()
		helper := func(logger Logger) { // logError 的角色：helper 自身多一层
			if skip {
				if a, ok := logger.(CallerSkipAdjuster); ok {
					logger = a.WithCallerSkip(1)
				}
			}
			logger.Error("x")
		}
		realCaller := func() { helper(wrapped) } // 真实调用点
		realCaller()
		m := parseZap(t, buf)
		f, _ := m["file"].(string)
		assert.Regexp(t, `mlog_test\.go:\d+$`, f, "caller 应指向测试文件真实调用点")
		return f
	}
	// WithCallerSkip(1) 补偿 helper 帧：两次 caller 落点不同。
	assert.NotEqual(t, logAt(false), logAt(true))
}

// logWrapper 内层不实现 CallerSkipAdjuster（如 mock）时：WithCallerSkip 静默
// 跳过补偿返回自身，不破坏原有日志链路（对齐 NewLogWrapper/With 的同类回退）。
func Test_logWrapper_WithCallerSkip_NoInnerAdjuster(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	wrapped := &logWrapper{Logger: NewMockLogger(m)}
	got := wrapped.WithCallerSkip(1)
	assert.Same(t, wrapped, got, "内层不实现 CallerSkipAdjuster 时应返回自身，不做任何包裹")
}

// formatError 对 pkg/errors（fmt.Formatter）用 %+v 打出完整堆栈。
// 栈帧断言（函数名 TestErrorLogWrapper_Error_PkgErrorsStack）验证 %+v 真实生效：
// 只查 root/outer 消息子串时，即使 formatError 退化为 err.Error() 测试也通过
// （两者都含这俩子串）；函数名只出现在 %+v 栈帧，不出现于 zap 的 file 字段
// （file 字段是调用点 文件:行，无函数名），故是栈存在的唯一可断言签名。
func TestErrorLogWrapper_Error_PkgErrorsStack(t *testing.T) {
	out := captureZapOutput(t, func(logger Logger) {
		NewLogWrapper(logger).Error("boom", pkgErrors.Wrap(errors.New("root"), "outer"))
	})
	assert.Contains(t, out, "outer")
	assert.Contains(t, out, "root")
	assert.Contains(t, out, "TestErrorLogWrapper_Error_PkgErrorsStack", "完整起源栈应包含调用函数栈帧（%+v），而非仅错误消息")
}

// formatError 对标准库 error（非 Formatter）退回 Error()，不丢信息也不崩溃。
func TestErrorLogWrapper_Error_PlainError(t *testing.T) {
	out := captureZapOutput(t, func(logger Logger) {
		NewLogWrapper(logger).Error("boom", errors.New("plain"))
	})
	assert.Contains(t, out, "plain")
}

// Errorf/wrapArgs 路径：格式化参数 + 标准库 error。
func TestErrorLogWrapper_Errorf_Formats(t *testing.T) {
	out := captureZapOutput(t, func(logger Logger) {
		NewLogWrapper(logger).Errorf("code=%d err=%v", 500, errors.New("plain"))
	})
	assert.Contains(t, out, "code=500")
	assert.Contains(t, out, "plain")
}

func TestErrorLogWrapper_ErrorCtx_Plain(t *testing.T) {
	out := captureZapOutput(t, func(logger Logger) {
		NewLogWrapper(logger).ErrorCtx(context.Background(), "boom", errors.New("plain"))
	})
	assert.Contains(t, out, "boom")
	assert.Contains(t, out, "plain")
}

func TestErrorLogWrapper_ErrorCtxf_Formats(t *testing.T) {
	out := captureZapOutput(t, func(logger Logger) {
		NewLogWrapper(logger).ErrorCtxf(context.Background(), "code=%d err=%v", 500, errors.New("plain"))
	})
	assert.Contains(t, out, "code=500")
	assert.Contains(t, out, "plain")
}

// formatError 对 nil 错误返回空串。
func TestFormatError_Nil(t *testing.T) {
	assert.Equal(t, "", formatError(nil))
}

// 回归防护：Error 级别不得双重堆栈。wrapper 已用 %+v 展开 pkg/errors 完整
// 起源栈；若生产 zap 配置再开 AddStacktrace(ErrorLevel)，会在 JSON 输出附加
// "日志打印位置"的运行时 stacktrace 字段，与 %+v 内容高度重叠（探针实证）。
func TestErrorLogWrapper_NoZapDuplicateStacktrace(t *testing.T) {
	out := captureZapOutput(t, func(logger Logger) {
		NewLogWrapper(logger).Error("boom", pkgErrors.Wrap(errors.New("root"), "outer"))
	})
	assert.Contains(t, out, "root", "pkg/errors 完整起源栈应打印")
	assert.Contains(t, out, "outer")
	assert.Contains(t, out, "TestErrorLogWrapper_NoZapDuplicateStacktrace", "pkg/errors 完整起源栈应含调用函数栈帧（%+v 生效）")
	assert.NotContains(t, out, `"stacktrace"`, "不得附加 zap 自动运行时栈（与 %+v 重复）")
}

// wrapper 全 level 覆盖：非 Fatal 全 level（含 Ctx 变体）在 zap/logrus 双后端打点，
// 并断言 With 附加的 trace.id 字段生效（经 FieldsInjector 落后端）。
// Fatal 系列 os.Exit 不可测（S 级排除项）。
func TestErrorLogWrapper_AllLevels(t *testing.T) {
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01},
		SpanID:     trace.SpanID{0x02},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)
	cases := []struct {
		name      string
		wantTrace bool
		log       func(l Logger)
	}{
		// 非 Ctx 方法用 context.Background() 求值：无效 span → 字段被跳过（P0-2）。
		{"Debug", false, func(l Logger) { l.Debug("m") }},
		{"Debugf", false, func(l Logger) { l.Debugf("m%d", 1) }},
		{"Warning", false, func(l Logger) { l.Warning("m") }},
		{"Warningf", false, func(l Logger) { l.Warningf("m%d", 1) }},
		{"Info", false, func(l Logger) { l.Info("m") }},
		{"Infof", false, func(l Logger) { l.Infof("m%d", 1) }},
		{"Error", false, func(l Logger) { l.Error("m") }},
		{"Errorf", false, func(l Logger) { l.Errorf("m%d", 1) }},
		// Ctx 方法用调用时 ctx 求值：有效 span → 落 trace.id 字段。
		{"DebugCtx", true, func(l Logger) { l.DebugCtx(ctx, "m") }},
		{"DebugCtxf", true, func(l Logger) { l.DebugCtxf(ctx, "m%d", 1) }},
		{"WarningCtx", true, func(l Logger) { l.WarningCtx(ctx, "m") }},
		{"WarningCtxf", true, func(l Logger) { l.WarningCtxf(ctx, "m%d", 1) }},
		{"InfoCtx", true, func(l Logger) { l.InfoCtx(ctx, "m") }},
		{"InfoCtxf", true, func(l Logger) { l.InfoCtxf(ctx, "m%d", 1) }},
		{"ErrorCtx", true, func(l Logger) { l.ErrorCtx(ctx, "m") }},
		{"ErrorCtxf", true, func(l Logger) { l.ErrorCtxf(ctx, "m%d", 1) }},
	}
	run := func(name string, mk func(t *testing.T) (Logger, func() map[string]any)) {
		t.Run(name, func(t *testing.T) {
			for _, c := range cases {
				t.Run(c.name, func(t *testing.T) {
					logger, extract := mk(t)
					c.log(With(logger, "trace.id", TraceID()))
					if c.wantTrace {
						assert.Equal(t, sc.TraceID().String(), extract()["trace.id"], c.name+" 应落 trace.id 字段")
					} else {
						assert.Nil(t, extract()["trace.id"], c.name+" 非 Ctx 方法不应落 trace.id 字段")
					}
				})
			}
		})
	}
	run("zap", func(t *testing.T) (Logger, func() map[string]any) {
		z, buf := newTestZap(t)
		return z, func() map[string]any { return parseZap(t, buf) }
	})
	run("logrus", func(t *testing.T) (Logger, func() map[string]any) {
		l, hook := newTestLogrus(t, true)
		return l, func() map[string]any {
			e := hook.LastEntry()
			return map[string]any{"trace.id": e.Data["trace.id"]}
		}
	})
}

// WithModule 保持包装（CallerSkip 补偿 + 模块字段）。
func TestErrorLogWrapper_WithModule(t *testing.T) {
	out := captureZapOutput(t, func(logger Logger) {
		NewLogWrapper(logger).WithModule("grpc").Error("boom")
	})
	assert.Contains(t, out, `"module":"grpc"`)
}

// logCtxKey 供 Valuer 测试从 ctx 取"每调用可变"的值，验证惰性求值。
type logCtxKey struct{}

// With + 有效 span：trace.id/span.id 字段按当前 ctx 落日志（Kratos Valuer 机制）。
func TestWith_TraceID_SpanID_Injected(t *testing.T) {
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01},
		SpanID:     trace.SpanID{0x02},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)
	out := captureZapOutput(t, func(logger Logger) {
		logger = With(logger, "trace.id", TraceID(), "span.id", SpanID())
		logger.InfoCtx(ctx, "x")
	})
	assert.Contains(t, out, sc.TraceID().String())
	assert.Contains(t, out, sc.SpanID().String())
}

// With + 无效 span：不落空 trace.id/span.id 字段（P0-2：evalValuers 跳过空串值）。
func TestWith_NoEmptyTraceFields_WhenInvalidSpan(t *testing.T) {
	out := captureZapOutput(t, func(logger Logger) {
		logger = With(logger, "trace.id", TraceID(), "span.id", SpanID())
		logger.InfoCtx(context.Background(), "x")
	})
	assert.NotContains(t, out, "trace.id")
	assert.NotContains(t, out, "span.id")
}

// With + 静态字段：非 Ctx 方法（用 context.Background()）也能附加常量字段。
func TestWith_StaticField_OnNonCtx(t *testing.T) {
	out := captureZapOutput(t, func(logger Logger) {
		With(logger, "service", "mars").Info("x")
	})
	assert.Contains(t, out, `"service":"mars"`)
}

// With + Valuer 惰性求值：命名 Valuer 与未命名闭包（func(ctx)any）两条断言路径，
// 每次打日志都用当前 ctx 重新求值，两次调用落不同值。
func TestWith_Valuer_EvaluatedPerCall(t *testing.T) {
	var calls int
	named := Valuer(func(ctx context.Context) any {
		calls++
		if v, ok := ctx.Value(logCtxKey{}).(string); ok {
			return "named:" + v
		}
		return ""
	})
	inline := func(ctx context.Context) any {
		calls++
		if v, ok := ctx.Value(logCtxKey{}).(string); ok {
			return "inline:" + v
		}
		return ""
	}
	out := captureZapOutput(t, func(logger Logger) {
		logger = With(logger, "n", named, "i", inline)
		logger.InfoCtx(context.WithValue(context.Background(), logCtxKey{}, "a"), "first")
		logger.InfoCtx(context.WithValue(context.Background(), logCtxKey{}, "b"), "second")
	})
	assert.Equal(t, 4, calls, "每个 Valuer 应在每次打日志时重新求值，而非构造时捕获")
	assert.Contains(t, out, `"n":"named:a"`)
	assert.Contains(t, out, `"n":"named:b"`)
	assert.Contains(t, out, `"i":"inline:a"`)
	assert.Contains(t, out, `"i":"inline:b"`)
}

// With + 用户字段（模拟 main 注入 biz.GetUser 闭包）：从 ctx 提取，无用户时落空串被跳过。
func TestWith_UserField_OnCtx(t *testing.T) {
	type user struct{ Name, Email string }
	type ukey struct{}
	ctx := context.WithValue(context.Background(), ukey{}, &user{Name: "duc", Email: "duc@example.com"})
	out := captureZapOutput(t, func(logger Logger) {
		logger = With(logger,
			"user_name", func(ctx context.Context) any {
				if u, ok := ctx.Value(ukey{}).(*user); ok {
					return u.Name
				}
				return ""
			},
			"email", func(ctx context.Context) any {
				if u, ok := ctx.Value(ukey{}).(*user); ok {
					return u.Email
				}
				return ""
			},
		)
		logger.InfoCtx(ctx, "x")
	})
	assert.Contains(t, out, `"user_name":"duc"`)
	assert.Contains(t, out, `"email":"duc@example.com"`)
}

// With 二次调用：字段合并进同一 wrapper（不嵌套），内层 backend 保持同一实例。
func TestWith_MergeIntoExistingWrapper(t *testing.T) {
	first := With(NewZapLogger(false), "a", 1)
	second := With(first, "b", 2)
	w, ok := second.(*logWrapper)
	assert.True(t, ok, "二次 With 应仍是 logWrapper")
	assert.Same(t, w.Logger, first.(*logWrapper).Logger, "不应嵌套 wrapper，内层应为同一 backend")
	assert.Equal(t, []any{"a", 1, "b", 2}, w.kvs)
	// 行为验证：合并后一次调用同时落 a/b 字段。
	out := captureOutput(t, func() Logger { return NewZapLogger(false) }, func(logger Logger) {
		With(With(logger, "a", 1), "b", 2).Info("x")
	})
	assert.Contains(t, out, `"a":1`)
	assert.Contains(t, out, `"b":2`)
}

// With 包住 logger 后 caller 补偿不漂移：file 字段仍指向真实调用点（本文件）。
func TestWith_CallerPointsToCallSite(t *testing.T) {
	for name, mk := range map[string]func() Logger{
		"zap":    func() Logger { return NewZapLogger(false) },
		"logrus": func() Logger { return NewLogrusLogger(false) },
	} {
		t.Run(name, func(t *testing.T) {
			out := captureOutput(t, mk, func(logger Logger) {
				With(logger, "trace.id", TraceID()).Info("boom")
			})
			assert.Contains(t, out, "mlog_test.go", "file 字段应指向真实调用点文件")
			assert.NotContains(t, out, "mlog/mlog.go", "file 字段不应指向 wrapper 内部")
		})
	}
}

// evalValuers 纯函数：跳过 nil 键/nil 值/空串，命名 Valuer 与未命名闭包都求值。
func TestEvalValuers(t *testing.T) {
	fields := evalValuers([]any{
		"nil_val", nil,
		nil, "no_key",
		"empty", "",
		"static", "v",
		"named", Valuer(func(ctx context.Context) any { return "n" }),
		"inline", func(ctx context.Context) any { return "i" },
		"nil_valuer", Valuer(func(ctx context.Context) any { return nil }),
	}, context.Background())
	assert.Equal(t, map[string]any{"static": "v", "named": "n", "inline": "i"}, fields)
}

// TestEvalValuers_OddLengthDropsDanglingKey 锁定奇数长度 kvs 的边界：最后的悬空
// key 无配对 value 时静默丢弃，不 panic 不落字段（约定调用方传偶数 kvs）。
func TestEvalValuers_OddLengthDropsDanglingKey(t *testing.T) {
	fields := evalValuers([]any{"a", 1, "b", 2, "dangling"}, context.Background())
	assert.Equal(t, map[string]any{"a": 1, "b": 2}, fields, "悬空 key 应静默丢弃")
}

// With 包住不实现 CallerSkipAdjuster 的 logger（如 mock）：静默跳过补偿，返回 wrapper。
func TestWith_BareLoggerNoCallerSkipAdjuster(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mock := NewMockLogger(m)
	mock.EXPECT().Info(gomock.Any())
	// mock 不实现 FieldsInjector：字段为空时透传内层 mock；调用不 panic。
	With(mock, "trace.id", TraceID()).Info("x")
}

// withFields 对不实现 FieldsInjector 的内层：字段求值非空但无法附加，静默丢弃不 panic。
func TestWith_InnerNoFieldsInjector_DropsFields(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mock := NewMockLogger(m)
	mock.EXPECT().Info(gomock.Any())
	With(mock, "trace.id", "static").Info("x")
}

// NewLogWrapper 对不实现 CallerSkipAdjuster 的 logger：静默跳过补偿，仍正常打日志。
func TestNewLogWrapper_NoCallerSkipAdjuster(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mock := NewMockLogger(m)
	mock.EXPECT().Error(gomock.Any(), gomock.Any())
	NewLogWrapper(mock).Error("boom", errors.New("plain"))
}

// 回归防护（P0）：recover 必须直接出现在被 defer 的 HandlePanic/HandlePanicWithCallback
// 方法体里。Go 的 recover 帧匹配（runtime 按 argp 判断）要求 recover 在被 defer 的
// 函数内，抽到共享 helper 后只剩"被 defer 函数恰好内联 helper"时才生效，接口类型
// defer 必然穿透——消费方全走 mlog.Logger 接口（cron.wrap、server recovery 中间件
// 等 30+ 处），一旦抽回 helper，生产 goroutine 的 panic 直接漏出。实锤见 cron
// Test_cronManager_wrap 红绿。captureOutput 传接口参数，模拟消费方调用形态。
func TestHandlePanic_InterfaceDeferRecovers(t *testing.T) {
	for name, mk := range map[string]func() Logger{
		"zap":    func() Logger { return NewZapLogger(false) },
		"logrus": func() Logger { return NewLogrusLogger(false) },
	} {
		t.Run(name, func(t *testing.T) {
			called := false
			captureOutput(t, mk, func(l Logger) {
				assert.NotPanics(t, func() {
					func() {
						defer l.HandlePanicWithCallback("boom", func(error) { called = true })
						panic("err")
					}()
				})
			})
			assert.True(t, called, "接口 defer 下 callback 应触发（recover 必须放方法体）")
		})
	}
}

// TestPanicStackGrowth 制造深栈触发 panicStack 的"截断→倍增"分支：起始缓冲仅
// 5KB，深递归 goroutine 栈远超此量，runtime.Stack 首次必然截断，panicStack 必须
// 翻倍缓冲直至抓全（P0：深层 goroutine 的 panic 日志不能因缓冲不足丢栈）。
// 承重断言在栈底帧 testing.tRunner：runtime.Stack 从栈顶（最内层帧）写起，5KB
// 截断只切栈底，panicStack/TestPanicStackGrowth 两帧在栈顶 5KB 内必在——只查它俩
// 时去掉倍增逻辑测试仍通过（实锤变异可存活）；tRunner 是最外层帧，仅完整缓冲才有。
func TestPanicStackGrowth(t *testing.T) {
	var deep func(n int) []byte
	deep = func(n int) []byte {
		if n <= 0 {
			return panicStack()
		}
		return deep(n - 1)
	}
	out := string(deep(20000)) // 约数 MB 栈，远超 5KB 起始缓冲
	assert.Contains(t, out, "panicStack")
	assert.Contains(t, out, "TestPanicStackGrowth")
	assert.Contains(t, out, "testing.tRunner", "倍增必须抓全深栈到最外层 tRunner 帧，而非只留栈顶 5KB")
}

package mlog

//go:generate go tool mockgen -destination ./mock_logger.go -package mlog github.com/duc-cnzj/mars/v6/internal/mlog Logger
import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"go.opentelemetry.io/otel/trace"
)

// Logger 是 mlog 的日志抽象：zap/logrus 两后端与 logWrapper 包装都实现它，
// 全仓依赖方经此接口注入日志（NewForConfig 按配置构建具体后端）。
// 方法族：WithModule 派生带 module 字段的 logger；Debug/Warning/Info/Error/Fatal
// 五级 × (Print/Printf/Ctx/Ctxf) 四种变体；Flush 冲刷后端缓冲；HandlePanic 系列
// 兜底 recover panic（debug 模式重抛）。Ctx 变体用调用时 ctx 求值附加字段。
type Logger interface {
	WithModule(module string) Logger

	Debug(v ...any)
	Debugf(format string, v ...any)

	DebugCtx(ctx context.Context, v ...any)
	DebugCtxf(ctx context.Context, format string, v ...any)

	Warning(v ...any)
	Warningf(format string, v ...any)

	WarningCtx(ctx context.Context, v ...any)
	WarningCtxf(ctx context.Context, format string, v ...any)

	Info(v ...any)
	Infof(format string, v ...any)

	InfoCtx(ctx context.Context, v ...any)
	InfoCtxf(ctx context.Context, format string, v ...any)

	Error(v ...any)
	Errorf(format string, v ...any)

	ErrorCtx(ctx context.Context, v ...any)
	ErrorCtxf(ctx context.Context, format string, v ...any)

	Fatal(v ...any)
	Fatalf(format string, v ...any)

	FatalCtx(ctx context.Context, v ...any)
	FatalCtxf(ctx context.Context, format string, v ...any)

	Flush() error

	HandlePanic(title string)
	HandlePanicWithCallback(title string, callback func(error))
}

// CallerSkipAdjuster 是可选的日志能力：包装 Logger 的实现
// （如 logWrapper）用它补偿自身引入的调用帧，
// 避免 caller 定位被包装层污染而指向包装器内部。
type CallerSkipAdjuster interface {
	WithCallerSkip(skip int) Logger
}

// NewForConfig 按配置构建日志后端（zap/logrus），并用 NewLogWrapper
// 包裹返回的实例，使 Error/Fatal 系列方法自动打出错误完整栈。
func NewForConfig(cfg *config.Config) Logger {
	var (
		channel string
		debug   bool
	)
	if cfg != nil {
		channel = cfg.LogChannel
		debug = cfg.Debug
	}
	switch channel {
	case "zap":
		return NewLogWrapper(NewZapLogger(debug))
	case "logrus":
		fallthrough
	default:
		return NewLogWrapper(NewLogrusLogger(debug))
	}
}

// Valuer 是惰性求值的日志字段值：每次打日志时用当前 ctx 求值。
// 配合 With 附加随请求变化的元数据（TraceID/SpanID/用户信息），字段的
// "来源"由注入方（main）定义，mlog 只提供附加与求值机制，不 import auth。
type Valuer func(ctx context.Context) any

// FieldsInjector 是可选的日志能力：支持附加字段的 Logger 实现
// （logrus/zap 原生实现，纯日志原语）。logWrapper 用它在日志上附加
// Valuer 求值结果。
type FieldsInjector interface {
	WithFields(fields map[string]any) Logger
}

// timestampLayout 是 zap/logrus 两后端统一的时间戳格式：NewZapLogger 的
// EncodeTime 与 NewLogrusLogger 的 Text/JSON formatter 共用同一布局，
// 保证 debug(Text)/production(JSON) 两后端输出时间戳一致。
const timestampLayout = "2006-01-02 15:04:05.000"

// TraceID 返回从 ctx 提取 TraceID 的 Valuer，配合 With 附加 trace.id 字段。
// ctx 无有效 span 时返回空串（evalValuers 跳过空值，不落空字段污染日志）。
func TraceID() Valuer {
	return func(ctx context.Context) any {
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			return sc.TraceID().String()
		}
		return ""
	}
}

// SpanID 返回从 ctx 提取 SpanID 的 Valuer，配合 With 附加 span.id 字段。
func SpanID() Valuer {
	return func(ctx context.Context) any {
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			return sc.SpanID().String()
		}
		return ""
	}
}

// evalValuers 把 kvs 求值成字段 map：Valuer 用 ctx 求值，跳过 nil 键、nil 值
// 与空串值（无效 span / 无用户时不落空字段污染日志，对齐 P0-2 设计）。
// 同时接受未命名闭包 func(context.Context) any（main 里内联写的字段来源）：
// 定义类型 Valuer 与未命名函数字面量非同一类型，val.(Valuer) 断言不到内联闭包，
// 必须再兜底一层 func(ctx)any 断言（实测验证：named assert OK / unnamed FAIL）。
func evalValuers(kvs []any, ctx context.Context) map[string]any {
	var fields map[string]any
	for i := 0; i+1 < len(kvs); i += 2 {
		key, val := kvs[i], kvs[i+1]
		switch fn := val.(type) {
		case Valuer:
			val = fn(ctx)
		case func(context.Context) any:
			val = fn(ctx)
		}
		if key == nil || val == nil || val == "" {
			continue
		}
		if fields == nil {
			fields = make(map[string]any, 1)
		}
		fields[fmt.Sprint(key)] = val
	}
	return fields
}

// formatError 优先用 %+v 打出完整错误栈——仅对实现了 fmt.Formatter 的错误
// （如 pkg/errors 的 withStack、biz.grpcStatusError）生效；其余错误（标准库
// %w 链、status.Error）退回 err.Error()，保证不丢信息也不依赖特定包装库。
func formatError(err error) string {
	if err == nil {
		return ""
	}
	if _, ok := err.(fmt.Formatter); ok {
		return fmt.Sprintf("%+v", err)
	}
	return err.Error()
}

// wrapErrors 把参数中的 error 逐个展开为栈字符串，非 error 原样保留。
func wrapErrors(v []any) []any {
	out := make([]any, len(v))
	for i, x := range v {
		if err, ok := x.(error); ok {
			out[i] = formatError(err)
		} else {
			out[i] = x
		}
	}
	return out
}

// panicStack 抓取当前 goroutine 的栈，起始 5KB、满则倍增，避免深栈截断。
func panicStack() []byte {
	bf := make([]byte, 1024*5)
	for {
		n := runtime.Stack(bf, false)
		if n < len(bf) {
			return bf[:n]
		}
		bf = make([]byte, len(bf)*2)
	}
}

// handleRecovered 是 zap/logrus 后端共用的 panic 后续处理：recover 必须由各
// HandlePanic/HandlePanicWithCallback 方法体直接调用（Go 的 recover 帧匹配要求
// 它出现在被 defer 的函数里，抽到 helper 后接口类型 defer 会穿透 panic，见
// zap.go/logrus.go HandlePanic 注释），此函数只处理已捕获的错误：归一 panic 值触发
// callback（可为 nil，即 HandlePanic 路径），打错误日志（含 panicStack），
// debug 模式下重新抛出。log 为后端的 Errorf，panic 日志走后端自身输出。
func handleRecovered(log func(format string, v ...any), debug bool, title string, callback func(error), err any) {
	if err == nil {
		return
	}
	if callback != nil {
		switch e := err.(type) {
		case error:
			callback(e)
		case string:
			callback(errors.New(e))
		default:
			// panic 值可以是任意类型，非 error/string 时也要触发 callback，
			// 否则调用方会误以为"没有 panic"。
			callback(fmt.Errorf("%v", e))
		}
	}
	log("[Panic]: title: %v, err: %v --- [%s]", title, err, string(panicStack()))
	if debug {
		panic(err)
	}
}

// logWrapper 是 NewLogWrapper/With 返回的包装类型：内嵌 Logger，
// 使 Error/Fatal 系列方法自动把 error 参数展开为完整栈（见 formatError），
// 并按调用时 ctx 求值 kvs 附加日志字段，其余方法原样透传内层。
type logWrapper struct {
	Logger
	kvs []any // 附加字段的 key-value 列表，value 可为 Valuer（按调用求值）
}

// NewLogWrapper 包装 Logger，使 Error 系列方法在遇到带栈错误
// （pkg/errors / fmt.Formatter 实现）时打出完整堆栈。
// NewForConfig 用它包裹返回的实例，全仓日志统一获得错误栈能力。
func NewLogWrapper(logger Logger) Logger {
	// 包装层会引入一帧，补偿到 caller 定位，避免 file 字段指向本文件。
	if a, ok := logger.(CallerSkipAdjuster); ok {
		logger = a.WithCallerSkip(1)
	}
	return &logWrapper{Logger: logger}
}

// With 对齐 kratos/log.With：给 logger 附加 key-value 字段，value 可为常量
// 或 Valuer（每次打日志用当前 ctx 求值）。在 main 中声明随请求变化的元数据：
//
//	logger = mlog.With(logger,
//	    "trace.id", mlog.TraceID(),
//	    "span.id", mlog.SpanID(),
//	    "user_name", func(ctx context.Context) any { ... },
//	)
//
// 若 logger 已是 logWrapper（NewForConfig 的产物），把字段合并到同一
// wrapper，避免嵌套包装引入多余调用帧导致 caller 定位漂移。
func With(logger Logger, kvs ...any) Logger {
	if w, ok := logger.(*logWrapper); ok {
		merged := make([]any, 0, len(w.kvs)+len(kvs))
		merged = append(merged, w.kvs...)
		merged = append(merged, kvs...)
		return &logWrapper{Logger: w.Logger, kvs: merged}
	}
	// 裸 logger：补 caller 补偿后包裹（与 NewLogWrapper 相同路径）。
	if a, ok := logger.(CallerSkipAdjuster); ok {
		logger = a.WithCallerSkip(1)
	}
	return &logWrapper{Logger: logger, kvs: append([]any{}, kvs...)}
}

// WithModule 透传给内层并保持包裹（caller 补偿 + module 字段 + 附加字段）。
func (e *logWrapper) WithModule(module string) Logger {
	return &logWrapper{Logger: e.Logger.WithModule(module), kvs: e.kvs}
}

// withFields 求值 kvs 里的 Valuer 并附加到内层 logger，返回可直接打日志的
// logger。只做构造、不引入 wrapper 帧：调用链保持 caller→wrapper→inner 两层，
// caller skip 不漂移（zap/logrus 的 WithCallerSkip 补偿据此设计）。
func (e *logWrapper) withFields(ctx context.Context) Logger {
	fields := evalValuers(e.kvs, ctx)
	if len(fields) == 0 {
		return e.Logger
	}
	if fi, ok := e.Logger.(FieldsInjector); ok {
		return fi.WithFields(fields)
	}
	return e.Logger
}

// 以下 20 个 level 方法均为透传：先按 ctx 求值 kvs 附加字段（非 Ctx 变体用
// context.Background()，Ctx 变体用调用时 ctx），再转发内层 Logger；Error/Fatal
// 系列额外把参数中的 error 展开为完整栈。注释风格与 zap/logrus 后端对齐
// （核心方法注释，Ctx 变体不注释）。

// Debug 透传内层 logger。
func (e *logWrapper) Debug(v ...any) {
	e.withFields(context.Background()).Debug(v...)
}

func (e *logWrapper) DebugCtx(ctx context.Context, v ...any) {
	e.withFields(ctx).DebugCtx(ctx, v...)
}

// Debugf 透传内层 logger。
func (e *logWrapper) Debugf(format string, v ...any) {
	e.withFields(context.Background()).Debugf(format, v...)
}

func (e *logWrapper) DebugCtxf(ctx context.Context, format string, v ...any) {
	e.withFields(ctx).DebugCtxf(ctx, format, v...)
}

// Warning 透传内层 logger。
func (e *logWrapper) Warning(v ...any) {
	e.withFields(context.Background()).Warning(v...)
}

func (e *logWrapper) WarningCtx(ctx context.Context, v ...any) {
	e.withFields(ctx).WarningCtx(ctx, v...)
}

// Warningf 透传内层 logger。
func (e *logWrapper) Warningf(format string, v ...any) {
	e.withFields(context.Background()).Warningf(format, v...)
}

func (e *logWrapper) WarningCtxf(ctx context.Context, format string, v ...any) {
	e.withFields(ctx).WarningCtxf(ctx, format, v...)
}

// Info 透传内层 logger。
func (e *logWrapper) Info(v ...any) {
	e.withFields(context.Background()).Info(v...)
}

func (e *logWrapper) InfoCtx(ctx context.Context, v ...any) {
	e.withFields(ctx).InfoCtx(ctx, v...)
}

// Infof 透传内层 logger。
func (e *logWrapper) Infof(format string, v ...any) {
	e.withFields(context.Background()).Infof(format, v...)
}

func (e *logWrapper) InfoCtxf(ctx context.Context, format string, v ...any) {
	e.withFields(ctx).InfoCtxf(ctx, format, v...)
}

// Error 透传内层 logger，并把参数中的 error 展开为完整栈。
func (e *logWrapper) Error(v ...any) {
	e.withFields(context.Background()).Error(wrapErrors(v)...)
}

func (e *logWrapper) ErrorCtx(ctx context.Context, v ...any) {
	e.withFields(ctx).ErrorCtx(ctx, wrapErrors(v)...)
}

// Errorf 透传内层 logger，并把参数中的 error 展开为完整栈。
func (e *logWrapper) Errorf(format string, v ...any) {
	e.withFields(context.Background()).Errorf(format, wrapErrors(v)...)
}

func (e *logWrapper) ErrorCtxf(ctx context.Context, format string, v ...any) {
	e.withFields(ctx).ErrorCtxf(ctx, format, wrapErrors(v)...)
}

// Fatal 透传内层 logger，并把参数中的 error 展开为完整栈。
func (e *logWrapper) Fatal(v ...any) {
	e.withFields(context.Background()).Fatal(wrapErrors(v)...)
}

func (e *logWrapper) FatalCtx(ctx context.Context, v ...any) {
	e.withFields(ctx).FatalCtx(ctx, wrapErrors(v)...)
}

// Fatalf 透传内层 logger，并把参数中的 error 展开为完整栈。
func (e *logWrapper) Fatalf(format string, v ...any) {
	e.withFields(context.Background()).Fatalf(format, wrapErrors(v)...)
}

func (e *logWrapper) FatalCtxf(ctx context.Context, format string, v ...any) {
	e.withFields(ctx).FatalCtxf(ctx, format, wrapErrors(v)...)
}

var _ Logger = (*logWrapper)(nil)

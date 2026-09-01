package data

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// tracerName 是 data 层统一使用的 OpenTelemetry tracer 名称，遵循 OTel 约定
// 取被插桩代码的 module path，作为 instrumentation scope 在 trace 面板区分插桩来源。
const tracerName = "github.com/duc-cnzj/mars/v6/internal/data"

// tracer 是 data 层全部 repo 方法共享的 tracer，避免各处重复构造。
var tracer = otel.Tracer(tracerName)

// endSpan 结束 span 并记录错误：err 非 nil 时先 RecordError 落异常事件、再把
// 状态置为 Error（描述为错误原文），最后统一 End。
// 供带 error 返回值的 repo 方法以 defer 闭包调用，读取命名返回值 err：
//
//	defer func() { endSpan(span, err) }()
//
// 必须用闭包而非 defer endSpan(span, err)——defer 语句注册时即求值参数，直接把
// err 传参会定格为方法入口处的 nil，永远记录不到错误。
func endSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// spanCall 在 ctx 上创建名为 name 的 child span，包裹一次单调用并返回其结果：
// fn 出错时先 RecordError 落异常事件、再把状态置为 Error（描述为错误原文），
// 返回后统一 End。
// 供 fetchClusterBoard/fetchResourceSnapshot 内逐 k8s List 调用的耗时观测使用，
// 让 trace 面板能看到每个 List 的独立耗时，而不只是整个快照拉取的总时长。
// 每次现取 otel.Tracer 而非复用包级 tracer：包级 tracer 在 init 时捕获、首个
// SetTracerProvider 后即钉死首 provider（setDelegate 只 swap 一次），现取才能
// 在测试中多轮 set/还原全局 provider 捕获 span。
func spanCall[V any](ctx context.Context, name string, fn func(context.Context) (V, error)) (V, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, name)
	defer span.End()
	v, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return v, err
}

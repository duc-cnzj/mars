package data

import (
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

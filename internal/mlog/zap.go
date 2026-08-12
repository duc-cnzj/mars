package mlog

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ Logger = (*zapLogger)(nil)

type zapLogger struct {
	sugar *zap.SugaredLogger
	debug bool
}

// NewZapLogger 构建 zap 日志后端：debug 模式走 DevelopmentConfig（终端彩色输出），
// 生产模式走 ProductionConfig（JSON 输出）。返回实例实现 FieldsInjector/
// CallerSkipAdjuster，供 logWrapper 附加字段与补偿 caller 帧。
func NewZapLogger(debug bool) Logger {
	var (
		logger *zap.Logger
		cfg    zap.Config
	)
	opts := []zap.Option{zap.AddCallerSkip(1)}
	if debug {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}
	// DisableStacktrace 关闭 zap 自动附加的"日志打印位置"运行时栈：wrapper 已用
	// %+v 打出 pkg/errors 完整起源栈，zap 的自动栈（生产默认 StacktraceLevel=
	// ErrorLevel）与之高度重叠，会造成每个 Error 日志双重堆栈（探针实证见
	// TestErrorLogWrapper_NoZapDuplicateStacktrace）。
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.CallerKey = "file"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(timestampLayout)
	// cfg 全部字段为合法值（NewProductionConfig/NewDevelopmentConfig 派生），
	// Build 错误不可达；按 S 级零死代码纪律丢弃 + 注释说明。
	logger, _ = cfg.Build(opts...)

	// Sugar() 在 base 之上 +2 帧（sugar.log/sugar 方法自身），
	// 加上 opts 里的 AddCallerSkip(1)，callerSkip 天然累计在 sugar 里，
	// WithFields/WithModule/WithCallerSkip 不再从 base 重建，也无需记账字段。
	return &zapLogger{sugar: logger.Sugar(), debug: debug}
}

// HandlePanic recover 后打 panic 日志（逻辑见 handleRecovered），debug 模式重抛。
// recover 必须直接在此方法体里调用而不能抽到 helper：Go 的 recover 帧匹配
// 要求 recover 出现在被 defer 的函数内（runtime 按 argp 判断），抽到共享函数后
// 只剩"被 defer 函数恰好内联共享函数"时有效，接口类型 defer 必然穿透。实锤见
// zap/logrus 两后端对照 + cron 消费方 Test_cronManager_wrap 回归。
func (z *zapLogger) HandlePanic(title string) {
	err := recover()
	handleRecovered(z.Errorf, z.debug, title, nil, err)
}

// HandlePanicWithCallback recover 后先触发 callback 再打 panic 日志（逻辑见
// handleRecovered），debug 模式重抛。recover 放方法体的原因同 HandlePanic。
func (z *zapLogger) HandlePanicWithCallback(title string, callback func(error)) {
	err := recover()
	handleRecovered(z.Errorf, z.debug, title, callback, err)
}

// Flush 冲刷 zap 内部缓冲：Sync() 返回底层写入错误。与 logrus 后端 Flush
// 的 no-op 语义对齐（都是"冲刷未落盘数据"的契约，差异由各后端缓冲决定）。
func (z *zapLogger) Flush() error {
	return z.sugar.Sync()
}

// Debug 打印 Debug 级别日志。
func (z *zapLogger) Debug(v ...any) {
	z.sugar.Debug(v...)
}

// DebugCtx 打印 Debug 级别日志。
func (z *zapLogger) DebugCtx(_ context.Context, v ...any) {
	z.sugar.Debug(v...)
}

// Debugf 格式化打印 Debug 级别日志。
func (z *zapLogger) Debugf(format string, v ...any) {
	z.sugar.Debugf(format, v...)
}

// DebugCtxf 格式化打印 Debug 级别日志。
func (z *zapLogger) DebugCtxf(_ context.Context, format string, v ...any) {
	z.sugar.Debugf(format, v...)
}

// Warning 打印 Warning 级别日志。
func (z *zapLogger) Warning(v ...any) {
	z.sugar.Warn(v...)
}

// WarningCtx 打印 Warning 级别日志。
func (z *zapLogger) WarningCtx(_ context.Context, v ...any) {
	z.sugar.Warn(v...)
}

// Warningf 格式化打印 Warning 级别日志。
func (z *zapLogger) Warningf(format string, v ...any) {
	z.sugar.Warnf(format, v...)
}

// WarningCtxf 格式化打印 Warning 级别日志。
func (z *zapLogger) WarningCtxf(_ context.Context, format string, v ...any) {
	z.sugar.Warnf(format, v...)
}

// Info 打印 Info 级别日志。
func (z *zapLogger) Info(v ...any) {
	z.sugar.Info(v...)
}

// InfoCtx 打印 Info 级别日志。
func (z *zapLogger) InfoCtx(_ context.Context, v ...any) {
	z.sugar.Info(v...)
}

// Infof 格式化打印 Info 级别日志。
func (z *zapLogger) Infof(format string, v ...any) {
	z.sugar.Infof(format, v...)
}

// InfoCtxf 格式化打印 Info 级别日志。
func (z *zapLogger) InfoCtxf(_ context.Context, format string, v ...any) {
	z.sugar.Infof(format, v...)
}

// Error 打印 Error 级别日志。
func (z *zapLogger) Error(v ...any) {
	z.sugar.Error(v...)
}

// ErrorCtx 打印 Error 级别日志。
func (z *zapLogger) ErrorCtx(_ context.Context, v ...any) {
	z.sugar.Error(v...)
}

// Errorf 格式化打印 Error 级别日志。
func (z *zapLogger) Errorf(format string, v ...any) {
	z.sugar.Errorf(format, v...)
}

// ErrorCtxf 格式化打印 Error 级别日志。
func (z *zapLogger) ErrorCtxf(_ context.Context, format string, v ...any) {
	z.sugar.Errorf(format, v...)
}

// Fatal 打印 Fatal 级别日志并退出进程。
func (z *zapLogger) Fatal(v ...any) {
	z.sugar.Fatal(v...)
}

// Fatalf 格式化打印 Fatal 级别日志并退出进程。
func (z *zapLogger) Fatalf(format string, v ...any) {
	z.sugar.Fatalf(format, v...)
}

// FatalCtx 打印 Fatal 级别日志并退出进程。
func (z *zapLogger) FatalCtx(_ context.Context, v ...any) {
	z.sugar.Fatal(v...)
}

// FatalCtxf 格式化打印 Fatal 级别日志并退出进程。
func (z *zapLogger) FatalCtxf(_ context.Context, format string, v ...any) {
	z.sugar.Fatalf(format, v...)
}

// WithFields 实现 mlog.FieldsInjector：附加字段后返回新 logger（纯日志原语，
// 不含 ctx 提取逻辑——元数据由 logWrapper 按调用求值后传入）。
// sugar.With 保留 core 的 callerSkip 与已附加字段，链式调用不丢状态。
func (z *zapLogger) WithFields(fields map[string]any) Logger {
	zfs := make([]any, 0, len(fields))
	for k, v := range fields {
		zfs = append(zfs, zap.Any(k, v))
	}
	return &zapLogger{debug: z.debug, sugar: z.sugar.With(zfs...)}
}

// WithModule 在 sugar 链上附加 module 字段并返回新 logger，链式调用保留
// callerSkip 与已附加字段（与 logrus 后端 WithModule 行为对齐）。
func (z *zapLogger) WithModule(module string) Logger {
	return &zapLogger{debug: z.debug, sugar: z.sugar.With(zap.String("module", module))}
}

// WithCallerSkip 实现 mlog.CallerSkipAdjuster：包装层每引入一帧就 +1。
// sugar 的 core 已携带累计 callerSkip（NewZapLogger 的 AddCallerSkip(1) +
// Sugar() 的 +2），WithOptions 在此之上累加，无需单独记账字段。
func (z *zapLogger) WithCallerSkip(skip int) Logger {
	return &zapLogger{debug: z.debug, sugar: z.sugar.WithOptions(zap.AddCallerSkip(skip))}
}

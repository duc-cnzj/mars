package mlog

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	l          *logrus.Logger
	debug      bool
	module     string
	callerSkip int           // 包装层（如 logWrapper）附加的帧数，fields() 定位 caller 时加深
	baseFields logrus.Fields // WithFields 附加的字段，fields() 合并输出
}

var _ Logger = (*logrusLogger)(nil)

// NewLogrusLogger 构建 logrus 日志后端：debug 模式 TextFormatter 终端可读，
// 生产模式 JSONFormatter 结构化输出，统一 timestampLayout 时间戳。返回实例
// 实现 FieldsInjector/CallerSkipAdjuster，供 logWrapper 附加字段与补偿 caller 帧。
func NewLogrusLogger(debug bool) Logger {
	logger := logrus.New()

	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: timestampLayout,
	})

	if debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: timestampLayout,
		})
		logger.SetLevel(logrus.InfoLevel)
	}

	return &logrusLogger{
		debug: debug,
		l:     logger,
	}
}

// WithModule 附加 module 字段并返回新 logger，保留 callerSkip 与 baseFields
// （与 zap 后端 WithModule 行为对齐）。
func (z *logrusLogger) WithModule(module string) Logger {
	return &logrusLogger{
		debug:      z.debug,
		module:     module,
		l:          z.l,
		callerSkip: z.callerSkip,
		baseFields: z.baseFields,
	}
}

// WithCallerSkip 实现 mlog.CallerSkipAdjuster：包装层每引入一帧就 +1。
func (z *logrusLogger) WithCallerSkip(skip int) Logger {
	return &logrusLogger{
		debug:      z.debug,
		module:     z.module,
		l:          z.l,
		callerSkip: z.callerSkip + skip,
		baseFields: z.baseFields,
	}
}

// WithFields 实现 mlog.FieldsInjector：附加字段后返回新 logger（纯日志原语，
// 不含 ctx 提取逻辑——元数据由 logWrapper 按调用求值后传入）。
func (z *logrusLogger) WithFields(fields map[string]any) Logger {
	return &logrusLogger{
		debug:      z.debug,
		module:     z.module,
		l:          z.l,
		callerSkip: z.callerSkip,
		baseFields: fields,
	}
}

// HandlePanic recover 后打 panic 日志（逻辑见 handleRecovered），debug 模式重抛。
// recover 必须直接在此方法体里调用而不能抽到 helper：Go 的 recover 帧匹配
// 要求 recover 出现在被 defer 的函数内（runtime 按 argp 判断），抽到共享函数后
// 只剩"被 defer 函数恰好内联共享函数"时有效，接口类型 defer 必然穿透。实锤见
// zap/logrus 两后端对照 + cron 消费方 Test_cronManager_wrap 回归。
func (z *logrusLogger) HandlePanic(title string) {
	err := recover()
	handleRecovered(z.Errorf, z.debug, title, nil, err)
}

// HandlePanicWithCallback recover 后先触发 callback 再打 panic 日志（逻辑见
// handleRecovered），debug 模式重抛。recover 放方法体的原因同 HandlePanic。
func (z *logrusLogger) HandlePanicWithCallback(title string, callback func(error)) {
	err := recover()
	handleRecovered(z.Errorf, z.debug, title, callback, err)
}

// Flush 为 no-op：logrus 直接写 os.Stdout，无内部缓冲，无需冲刷。
// 与 zap 后端 Flush 返回 Sync() 错误的语义对齐（都是"冲刷未落盘数据"的契约）。
func (z *logrusLogger) Flush() error {
	return nil
}

// Debug 打印 Debug 级别日志。
func (z *logrusLogger) Debug(v ...any) {
	z.l.WithFields(z.fields()).Debug(v...)
}

// Debugf 格式化打印 Debug 级别日志。
func (z *logrusLogger) Debugf(format string, v ...any) {
	z.l.WithFields(z.fields()).Debugf(format, v...)
}

func (z *logrusLogger) DebugCtx(_ context.Context, v ...any) {
	z.l.WithFields(z.fields()).Debug(v...)
}

func (z *logrusLogger) DebugCtxf(_ context.Context, format string, v ...any) {
	z.l.WithFields(z.fields()).Debugf(format, v...)
}

// Warning 打印 Warning 级别日志。
func (z *logrusLogger) Warning(v ...any) {
	z.l.WithFields(z.fields()).Warn(v...)
}

// Warningf 格式化打印 Warning 级别日志。
func (z *logrusLogger) Warningf(format string, v ...any) {
	z.l.WithFields(z.fields()).Warnf(format, v...)
}

func (z *logrusLogger) WarningCtx(_ context.Context, v ...any) {
	z.l.WithFields(z.fields()).Warn(v...)
}

func (z *logrusLogger) WarningCtxf(_ context.Context, format string, v ...any) {
	z.l.WithFields(z.fields()).Warnf(format, v...)
}

// Info 打印 Info 级别日志。
func (z *logrusLogger) Info(v ...any) {
	z.l.WithFields(z.fields()).Info(v...)
}

// Infof 格式化打印 Info 级别日志。
func (z *logrusLogger) Infof(format string, v ...any) {
	z.l.WithFields(z.fields()).Infof(format, v...)
}

func (z *logrusLogger) InfoCtx(_ context.Context, v ...any) {
	z.l.WithFields(z.fields()).Info(v...)
}

func (z *logrusLogger) InfoCtxf(_ context.Context, format string, v ...any) {
	z.l.WithFields(z.fields()).Infof(format, v...)

}

// Error 打印 Error 级别日志。
func (z *logrusLogger) Error(v ...any) {
	z.l.WithFields(z.fields()).Error(v...)
}

// Errorf 格式化打印 Error 级别日志。
func (z *logrusLogger) Errorf(format string, v ...any) {
	z.l.WithFields(z.fields()).Errorf(format, v...)
}

func (z *logrusLogger) ErrorCtx(_ context.Context, v ...any) {
	z.l.WithFields(z.fields()).Error(v...)
}

func (z *logrusLogger) ErrorCtxf(_ context.Context, format string, v ...any) {
	z.l.WithFields(z.fields()).Errorf(format, v...)
}

// Fatal 打印 Fatal 级别日志并退出进程。
func (z *logrusLogger) Fatal(v ...any) {
	z.l.WithFields(z.fields()).Fatal(v...)
}

// Fatalf 格式化打印 Fatal 级别日志并退出进程。
func (z *logrusLogger) Fatalf(format string, v ...any) {
	z.l.WithFields(z.fields()).Fatalf(format, v...)
}

func (z *logrusLogger) FatalCtx(_ context.Context, v ...any) {
	z.l.WithFields(z.fields()).Fatal(v...)
}

func (z *logrusLogger) FatalCtxf(_ context.Context, format string, v ...any) {
	z.l.WithFields(z.fields()).Fatalf(format, v...)
}

// fields 返回 caller 定位（file:line）、module 与 WithFields 附加字段。
func (z *logrusLogger) fields() logrus.Fields {
	pc, _, _, _ := runtime.Caller(2 + z.callerSkip)
	f := map[string]any{
		"file": "unknown",
	}
	if fn := runtime.FuncForPC(pc); fn != nil {
		file, line := fn.FileLine(pc)
		f["file"] = fmt.Sprintf("%s:%d", file, line)
	}
	if z.module != "" {
		f["module"] = z.module
	}
	for k, v := range z.baseFields {
		f[k] = v
	}
	return f
}

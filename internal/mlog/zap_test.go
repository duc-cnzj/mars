package mlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newTestZap 构造 zapLogger：JSON encoder 写入 buffer，调用方可从 buffer 解析断言。
// 编码配置与生产 NewZapLogger 对齐（CallerKey=file、EncodeTime=timestampLayout）。
func newTestZap(t *testing.T) (*zapLogger, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.CallerKey = "file"
	encCfg.TimeKey = "time"
	encCfg.EncodeTime = zapcore.TimeEncoderOfLayout(timestampLayout)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), zapcore.AddSync(buf), zapcore.DebugLevel)
	// AddCaller 等价于生产 cfg.Build 的默认 caller 记录；AddCallerSkip(1) 补偿 sugar 帧。
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &zapLogger{sugar: logger.Sugar(), debug: true}, buf
}

func parseZap(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	assert.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m))
	return m
}

func TestZapLogger_Entry_Levels(t *testing.T) {
	cases := []struct {
		name string
		log  func(l Logger)
		want string
	}{
		{"Debug", func(l Logger) { l.Debug("m") }, "debug"},
		{"Info", func(l Logger) { l.Info("m") }, "info"},
		{"Warning", func(l Logger) { l.Warning("m") }, "warn"},
		{"Error", func(l Logger) { l.Error("m") }, "error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			z, buf := newTestZap(t)
			c.log(z)
			m := parseZap(t, buf)
			assert.Equal(t, "m", m["msg"])
			assert.Equal(t, c.want, m["level"])
		})
	}
}

func TestZapLogger_FileField_NoTrailingDot(t *testing.T) {
	z, buf := newTestZap(t)
	z.Info("x")
	m := parseZap(t, buf)
	file, ok := m["file"].(string)
	assert.True(t, ok)
	assert.Regexp(t, `zap_test\.go:\d+$`, file)
	assert.NotRegexp(t, `\.go:\d+\.$`, file)
}

// 无 panic 时 HandlePanic 不产生任何副作用（不 panic、无日志输出）。
func TestZapLogger_HandlePanic_NoPanic(t *testing.T) {
	z, buf := newTestZap(t)
	z.HandlePanic("boom")
	assert.Empty(t, buf.Bytes())
}

func TestZapLogger_HandlePanic_LogsEntry(t *testing.T) {
	z, buf := newTestZap(t)
	z.debug = false // 避免 re-panic，验证日志内容
	func() {
		defer z.HandlePanic("boom")
		panic("panic value")
	}()
	m := parseZap(t, buf)
	msg, _ := m["msg"].(string)
	assert.Contains(t, msg, "boom")
	assert.Contains(t, msg, "panic value")
}

func TestZapLogger_HandlePanic_DebugRepanic(t *testing.T) {
	z, _ := newTestZap(t)
	assert.PanicsWithValue(t, "boom", func() {
		defer z.HandlePanic("boom")
		panic("boom")
	})
}

// panic 值为非 error/string 类型时 callback 仍须触发（P0-3 锁定）。
func TestZapLogger_HandlePanicWithCallback_NonErrorValue(t *testing.T) {
	z, _ := newTestZap(t)
	z.debug = false
	var got error
	func() {
		defer z.HandlePanicWithCallback("boom", func(e error) { got = e })
		panic(map[string]string{"a": "b"})
	}()
	assert.Error(t, got)
	assert.Contains(t, got.Error(), "a")
}

func TestZapLogger_HandlePanicWithCallback_String(t *testing.T) {
	z, _ := newTestZap(t)
	z.debug = false
	var got error
	func() {
		defer z.HandlePanicWithCallback("boom", func(e error) { got = e })
		panic("str panic")
	}()
	assert.Equal(t, "str panic", got.Error())
}

func TestZapLogger_HandlePanicWithCallback_DebugRepanic(t *testing.T) {
	z, _ := newTestZap(t)
	assert.PanicsWithValue(t, "boom", func() {
		defer z.HandlePanicWithCallback("boom", func(error) {})
		panic("boom")
	})
}

func TestZapLogger_HandlePanicWithCallback_Error(t *testing.T) {
	z, _ := newTestZap(t)
	z.debug = false
	root := errors.New("root")
	var got error
	func() {
		defer z.HandlePanicWithCallback("boom", func(e error) { got = e })
		panic(root)
	}()
	assert.Same(t, root, got)
}

func TestZapLoggerDebugMode(t *testing.T) {
	logger := NewZapLogger(true)
	assert.True(t, logger.(*zapLogger).debug)
}

func TestZapLoggerProductionMode(t *testing.T) {
	logger := NewZapLogger(false)
	assert.False(t, logger.(*zapLogger).debug)
}

func TestZapLoggerFlush(t *testing.T) {
	logger := NewZapLogger(true)
	logger.Flush()
}

// WithModule 在 sugar 链上附加 module 字段，不重建 core、不丢 caller 记录。
func Test_zapLogger_WithModule(t *testing.T) {
	z, buf := newTestZap(t)
	z.WithModule("grpc").Info("x")
	m := parseZap(t, buf)
	assert.Equal(t, "grpc", m["module"])
	// 链式 sugar.With 保留 callerSkip：file 仍指向测试文件真实调用点。
	assert.Regexp(t, `zap_test\.go:\d+$`, m["file"])
}

// WithCallerSkip 实现 mlog.CallerSkipAdjuster：为包装层（logWrapper）引入的每
// 一帧补偿 1 层 caller。双层闭包模拟生产链路（外层=真实调用点，内层=
// logWrapper.Info 包装帧），补偿后 caller 越过包装帧指向外层真实调用点。
func Test_zapLogger_WithCallerSkip(t *testing.T) {
	z, buf := newTestZap(t)
	// 同一结构两次打点：无补偿 vs WithCallerSkip(1)。
	logAt := func(wrapped bool) string {
		buf.Reset()
		outer := func() { // 真实调用点
			inner := func() { // logWrapper.Info 包装帧
				if wrapped {
					z.WithCallerSkip(1).Info("x")
				} else {
					z.Info("y")
				}
			}
			inner()
		}
		outer()
		m := parseZap(t, buf)
		f, _ := m["file"].(string)
		assert.Regexp(t, `zap_test\.go:\d+$`, f, "caller 应指向测试文件真实调用点")
		return f
	}
	// WithCallerSkip(1) 恰好多跳 1 帧：两次 caller 落点不同。
	assert.NotEqual(t, logAt(false), logAt(true))
}

func TestZapLogger_FormatMethods(t *testing.T) {
	cases := []struct {
		name string
		log  func(l Logger)
		want string
	}{
		{"Debugf", func(l Logger) { l.Debugf("x%d", 1) }, "x1"},
		{"Infof", func(l Logger) { l.Infof("x%d", 1) }, "x1"},
		{"Warningf", func(l Logger) { l.Warningf("x%d", 1) }, "x1"},
		{"Errorf", func(l Logger) { l.Errorf("x%d", 1) }, "x1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			z, buf := newTestZap(t)
			c.log(z)
			m := parseZap(t, buf)
			assert.Equal(t, c.want, m["msg"])
		})
	}
}

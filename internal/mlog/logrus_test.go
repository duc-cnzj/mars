package mlog

import (
	"errors"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

// newTestLogrus 构造 debug 可调的 logrusLogger：输出静音（hook 捕获 entry 断言，
// 不污染测试 stdout），与 formatter 无关。
func newTestLogrus(t *testing.T, debug bool) (*logrusLogger, *test.Hook) {
	t.Helper()
	l := NewLogrusLogger(debug).(*logrusLogger)
	l.l.SetOutput(io.Discard)
	hook := test.NewLocal(l.l)
	return l, hook
}

func TestLogrusLogger_Entry_Levels(t *testing.T) {
	cases := []struct {
		name string
		log  func(l Logger)
		want logrus.Level
	}{
		{"Debug", func(l Logger) { l.Debug("m") }, logrus.DebugLevel},
		{"Info", func(l Logger) { l.Info("m") }, logrus.InfoLevel},
		{"Warning", func(l Logger) { l.Warning("m") }, logrus.WarnLevel},
		{"Error", func(l Logger) { l.Error("m") }, logrus.ErrorLevel},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l, hook := newTestLogrus(t, true)
			c.log(l)
			e := hook.LastEntry()
			assert.Equal(t, "m", e.Message)
			assert.Equal(t, c.want, e.Level)
		})
	}
}

// file 字段格式：path/file.go:line，无尾点（P0-1 锁定，与 zap 后端一致）。
func TestLogrusLogger_FileField_NoTrailingDot(t *testing.T) {
	l, hook := newTestLogrus(t, true)
	l.Info("x")
	file, ok := hook.LastEntry().Data["file"].(string)
	assert.True(t, ok)
	assert.Regexp(t, `logrus_test\.go:\d+$`, file)
	assert.NotRegexp(t, `\.go:\d+\.$`, file)
}

func TestLogrusLogger_WithModule_AddsField(t *testing.T) {
	l, hook := newTestLogrus(t, true)
	l.WithModule("grpc").Info("x")
	assert.Equal(t, "grpc", hook.LastEntry().Data["module"])
}

// 无 panic 时 HandlePanicWithCallback 不触发 callback、不产生日志。
func TestLogrusLogger_HandlePanic_NoPanic(t *testing.T) {
	l, hook := newTestLogrus(t, true)
	called := false
	l.HandlePanicWithCallback("boom", func(error) { called = true })
	assert.False(t, called)
	assert.Nil(t, hook.LastEntry())
}

func TestLogrusLogger_HandlePanic_LogsEntry(t *testing.T) {
	l, hook := newTestLogrus(t, false)
	func() {
		defer l.HandlePanic("boom")
		panic("panic value")
	}()
	e := hook.LastEntry()
	assert.Contains(t, e.Message, "boom")
	assert.Contains(t, e.Message, "panic value")
}

func TestLogrusLogger_HandlePanic_DebugRepanic(t *testing.T) {
	l, _ := newTestLogrus(t, true)
	assert.PanicsWithValue(t, "boom", func() {
		defer l.HandlePanic("boom")
		panic("boom")
	})
}

// panic 值为非 error/string 类型时 callback 仍须触发（P0-3 锁定）。
func TestLogrusLogger_HandlePanicWithCallback_NonErrorValue(t *testing.T) {
	l, _ := newTestLogrus(t, false)
	var got error
	func() {
		defer l.HandlePanicWithCallback("boom", func(e error) { got = e })
		panic(map[string]string{"a": "b"})
	}()
	assert.Error(t, got)
	assert.Contains(t, got.Error(), "a")
}

func TestLogrusLogger_HandlePanicWithCallback_String(t *testing.T) {
	l, _ := newTestLogrus(t, false)
	var got error
	func() {
		defer l.HandlePanicWithCallback("boom", func(e error) { got = e })
		panic("str panic")
	}()
	assert.Equal(t, "str panic", got.Error())
}

func TestLogrusLogger_HandlePanicWithCallback_DebugRepanic(t *testing.T) {
	l, _ := newTestLogrus(t, true)
	assert.PanicsWithValue(t, "boom", func() {
		defer l.HandlePanicWithCallback("boom", func(error) {})
		panic("boom")
	})
}

func TestLogrusLogger_HandlePanicWithCallback_Error(t *testing.T) {
	l, _ := newTestLogrus(t, false)
	root := errors.New("root")
	var got error
	func() {
		defer l.HandlePanicWithCallback("boom", func(e error) { got = e })
		panic(root)
	}()
	assert.Same(t, root, got)
}

func Test_logrusLogger_Flush(t *testing.T) {
	logger := NewLogrusLogger(true)
	assert.Nil(t, logger.Flush())
}

func Test_logrusLogger_fields(t *testing.T) {
	logger := NewLogrusLogger(true).(*logrusLogger)
	logger.module = "x"
	fields := logger.fields()
	assert.Len(t, fields, 2)
	assert.Equal(t, "x", fields["module"])
	assert.NotEqual(t, "unknown", fields["file"])
}

// WithCallerSkip 实现 mlog.CallerSkipAdjuster：为包装层（logWrapper）引入的每
// 一帧补偿 1 层 caller。与 zap 后端同一行为契约，双层闭包模拟生产链路。
func Test_logrusLogger_WithCallerSkip(t *testing.T) {
	l, hook := newTestLogrus(t, true)
	logAt := func(wrapped bool) string {
		hook.Reset()
		outer := func() { // 真实调用点
			inner := func() { // logWrapper.Info 包装帧
				if wrapped {
					l.WithCallerSkip(1).Info("x")
				} else {
					l.Info("y")
				}
			}
			inner()
		}
		outer()
		file, _ := hook.LastEntry().Data["file"].(string)
		assert.Regexp(t, `logrus_test\.go:\d+$`, file, "caller 应指向测试文件真实调用点")
		return file
	}
	// WithCallerSkip(1) 恰好多跳 1 帧：两次 caller 落点不同。
	assert.NotEqual(t, logAt(false), logAt(true))
}

func TestLogrusLogger_FormatMethods(t *testing.T) {
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
			l, hook := newTestLogrus(t, true)
			c.log(l)
			assert.Equal(t, c.want, hook.LastEntry().Message)
		})
	}
}

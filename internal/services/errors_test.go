package services

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
)

// captureLogs 在 logger 构造窗口内重定向 stdout/stderr，返回捕获的日志输出。
// zap 生产配置写 stderr、logrus 写 stdout，故两者同时重定向（对齐 mlog 包内
// captureOutput 的用法）。
func captureLogs(t *testing.T, mk func() mlog.Logger, fn func(l mlog.Logger)) string {
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

// logError 作为收敛 helper 引入一帧调用栈，必须经 CallerSkipAdjuster 补偿一帧，
// 否则日志 file 字段指向 services/errors.go 而非真实调用方（回归：日志 file 漂移，
// 用户实锤 "file":"services/errors.go:16"）。双后端都要指向本测试文件。
func TestLogError_CallerPointsToCallSite(t *testing.T) {
	for name, channel := range map[string]string{
		"zap":    "zap",
		"logrus": "logrus",
	} {
		t.Run(name, func(t *testing.T) {
			out := captureLogs(t, func() mlog.Logger {
				return mlog.NewForConfig(&config.Config{LogChannel: channel}).WithModule("services/test")
			}, func(l mlog.Logger) {
				logError(context.Background(), l, errors.New("boom"))
			})
			assert.Contains(t, out, "errors_test.go", "file 字段应指向 logError 的真实调用方（本测试文件），而非 errors.go")
			assert.NotContains(t, out, "errors.go", "file 字段不应指向 logError helper 内部")
		})
	}
}

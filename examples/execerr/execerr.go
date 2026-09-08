// Package execerr 提供容器命令执行错误码（ExecError.code）的命名常量与
// 语义描述，供 examples/grpc 与 examples/http 两个 demo 复用，避免各自
// 重复维护同一段错误码解释。语义对齐 api/proto/container 的 ExecError 注释。
package execerr

import "fmt"

// ExecError.code 的固定语义码：
//   - 0-255 为容器内命令的非零退出码（命令已启动并结束）；
//   - -1/-2/-3 为服务端终止或失败信号，见下方常量。
const (
	// CodeTruncated 命令输出超限被服务端强制截断（ExecOnce）。
	CodeTruncated int64 = -1
	// CodeExecFailed 容器 exec 启动/执行失败（如命令在容器内不存在）。
	CodeExecFailed int64 = -2
	// CodeTimeout 命令执行超时被服务端强制终止（ExecOnce，timeout_seconds 上限）。
	CodeTimeout int64 = -3
)

// Describe 返回 ExecError.code 的中文语义描述，便于 demo 直观展示错误帧类型。
func Describe(code int64) string {
	switch code {
	case CodeTruncated:
		return "(命令输出超限被服务端强制截断)"
	case CodeExecFailed:
		return "(exec 启动/执行失败，如命令在容器内不存在)"
	case CodeTimeout:
		return "(命令执行超时被服务端强制终止)"
	default:
		return fmt.Sprintf("(容器内命令退出码 %d)", code)
	}
}

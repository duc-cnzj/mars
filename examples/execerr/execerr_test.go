package execerr

import "testing"

// TestDescribe 锁定 Describe 的语义描述：覆盖三个固定错误码分支与默认退出码分支。
func TestDescribe(t *testing.T) {
	cases := []struct {
		name string
		code int64
		want string
	}{
		{name: "truncated", code: CodeTruncated, want: "(命令输出超限被服务端强制截断)"},
		{name: "exec_failed", code: CodeExecFailed, want: "(exec 启动/执行失败，如命令在容器内不存在)"},
		{name: "timeout", code: CodeTimeout, want: "(命令执行超时被服务端强制终止)"},
		{name: "exit_code_zero", code: 0, want: "(容器内命令退出码 0)"},
		{name: "exit_code_three", code: 3, want: "(容器内命令退出码 3)"},
		{name: "exit_code_max", code: 255, want: "(容器内命令退出码 255)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Describe(tc.code); got != tc.want {
				t.Fatalf("Describe(%d) = %q，期望 %q", tc.code, got, tc.want)
			}
		})
	}
}

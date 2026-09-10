package biz

import (
	"reflect"
	"sort"
	"testing"
)

// TestPublicMethods_AlignsWithAccessControlDoc 是公开白名单的契约测试：断言白名单与
// doc/access_control.md §4.1「免登录服务」清单逐行一致。新增免登录方法必须同时更新
// 本测试、publicMethods 与文档三处，任何一处漏改都会在此失败（防契约与实现漂移）。
func TestPublicMethods_AlignsWithAccessControlDoc(t *testing.T) {
	want := []string{
		"/auth.Auth/Exchange",
		"/auth.Auth/Login",
		"/auth.Auth/Settings",
		"/cluster.Cluster/ClusterInfo",
		"/picture.Picture/Background",
		"/version.Version/Version",
	}
	// 白名单全 key 排序后与文档清单比对（排序逻辑内联于此，生产代码不为此保留导出函数）。
	got := make([]string, 0, len(publicMethods))
	for name := range publicMethods {
		got = append(got, name)
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("publicMethods 与 doc §4.1 漂移\n got: %v\nwant: %v", got, want)
	}
	for _, name := range want {
		if !IsPublicMethod(name) {
			t.Errorf("白名单缺少公开方法: %s", name)
		}
	}
	// 非白名单方法（如私有服务的命名空间级方法）一律不命中，防止误放行。
	// Info 需登录（拦截器注入 ctx 用户）不在白名单，回退为"登录即可"级别而非免登录。
	if IsPublicMethod("/namespace.Namespace/List") || IsPublicMethod("/auth.Auth/Info") {
		t.Errorf("私有方法不应命中白名单")
	}
}

package biz

import (
	authpb "github.com/duc-cnzj/mars/api/v6/proto/auth"
	clusterpb "github.com/duc-cnzj/mars/api/v6/proto/cluster"
	picturepb "github.com/duc-cnzj/mars/api/v6/proto/picture"
	versionpb "github.com/duc-cnzj/mars/api/v6/proto/version"
)

// publicMethods 是 mars 全部免登录 gRPC 方法的 fullMethodName 白名单：登录拦截器命中
// 白名单时跳过 Bearer token 校验直接放行，未命中一律要求有效 token。
// 未导出：外部包统一经 IsPublicMethod 查询，白名单本体只有本包（IsPublicMethod 与
// 契约测试）直接持有，避免暴露可变全局 map 给外部读写。
//
// 归属 biz 层：与 AccessBiz.RequireAdmin 的 allowlist 同属「按 fullMethodName 判访问
// 级别」的访问控制契约——RequireAdmin 管 admin 豁免，本白名单管免登录放行，两类契约
// 全收口业务层；middlewares 登录拦截器只消费 IsPublicMethod，不在传输层定义访问策略。
//
// 与 access.go 同属访问控制域但分工不同，故独立成文件：access.go 是依赖注入的
// AccessBiz 实例判定服务（运行时门卫），本文件是无状态的免登录白名单静态契约
// （编译期确定）——分开保持变更隔离，白名单改动不波及门卫逻辑。
//
// 直接引用 proto 生成的 *_FullMethodName 常量（grpc-go 官方产物，与 ServiceDesc/
// 拦截器 FullMethod 完全一致）：方法名写错即编译失败，零运行时反射，无需手写
// "/pkg.Svc/Method" 字符串路径。
//
// 新增免登录方法：在下方 map 追加一行对应服务的 *_FullMethodName 常量即可。
// 白名单与 doc/access_control.md §4.1「免登录服务」清单逐行一致，public_methods_test.go
// 的 TestPublicMethods_AlignsWithAccessControlDoc 契约测试会在二者漂移时失败。
//
// 相比原先的 guest 内嵌（AuthFuncOverride 无条件放行整个服务）：白名单把"公开"从
// 服务粒度收窄到方法粒度——新方法默认私有（安全默认），"公开"判定单一归属本处。
var publicMethods = map[string]struct{}{
	authpb.Auth_Login_FullMethodName:             {},
	authpb.Auth_Settings_FullMethodName:          {},
	authpb.Auth_Exchange_FullMethodName:          {},
	clusterpb.Cluster_ClusterInfo_FullMethodName: {},
	picturepb.Picture_Background_FullMethodName:  {},
	versionpb.Version_Version_FullMethodName:     {},
}

// IsPublicMethod 判断 fullMethodName 是否命中公开白名单（免登录放行）：对外统一
// 查询入口，middlewares 登录拦截器等跨包消费方不直接接触 publicMethods 本体。
func IsPublicMethod(fullMethodName string) bool {
	_, ok := publicMethods[fullMethodName]
	return ok
}

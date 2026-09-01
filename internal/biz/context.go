package biz

// context.go 承载用户信息的 context 存取（原 internal/auth 包整体并入）：
// SetUser/GetUser/MustGetUser + 私有 ctxTokenInfo key。auth 此前是单文件微包，
// 唯一依赖就是本包的 UserInfo；由于 auth→biz 反向引用会成环，biz 无法引用
// auth.MustGetUser，取用户回调才被逼到传输层绑定。把这三函数收进 biz 后，
// 环物理消失：NewAccessBiz 内部直接绑 MustGetUser，wire.Value 跨包装配整个删除。

import (
	"context"
	"errors"
)

// ctxTokenInfo 是 context 中注入用户信息所用的私有 key 类型。
// 用私有类型而非字符串作 key，避免与其他包在 context 中撞 key。
// 用空结构体【值】而非指针作 key：空结构体值恒相等（规范保证），
// 而不同函数里的 &ctxTokenInfo{} 指针是否相等是规范未定义的，依赖 gc 优化。
type ctxTokenInfo struct{}

// SetUser 将用户信息注入 context，返回携带用户的新 context。
// info 可为 nil：此时 GetUser 仍会返回错误，调用方可感知注入失效。
func SetUser(ctx context.Context, info *UserInfo) context.Context {
	return context.WithValue(ctx, ctxTokenInfo{}, info)
}

// GetUser 从 context 中取出 SetUser 注入的用户信息。
// 未注入用户，或注入值为 nil（含 (*UserInfo)(nil) 这种带类型 nil 指针）时，
// 返回 "user not found" 错误——不能让调用方拿到 nil 用户却以为成功。
func GetUser(ctx context.Context) (*UserInfo, error) {
	if info, ok := ctx.Value(ctxTokenInfo{}).(*UserInfo); ok && info != nil {
		return info, nil
	}

	return nil, errors.New("user not found")
}

// MustGetUser 的语义是"调用方保证当前上下文已注入用户"：拿不到用户即编程错误，
// 直接 panic 而非返回 nil——返回 nil 会把空指针一路传进下游，要么隐式 nil-deref，
// 要么被当作"非管理员"悄悄放行，掩盖鉴权链路本身的故障。
func MustGetUser(ctx context.Context) *UserInfo {
	info, err := GetUser(ctx)
	if err != nil {
		panic(err)
	}
	return info
}

// authenticate 校验 bearer token 并把用户注入 ctx（角色取登录身份/JWT），返回新 ctx。
// 这是 Authenticate 的纯校验基座：只做 token 校验与用户注入、不读取用户表，供
// Authenticate 内部调用。不单独导出——避免出现第二条鉴权路径导致生效角色策略漂移。
func authenticate(ctx context.Context, auth AuthBiz, token string) (context.Context, error) {
	user, err := auth.VerifyToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return SetUser(ctx, user), nil
}

// Authenticate 是 gRPC 拦截器与 HTTP 中间件共用的唯一鉴权核心：两个传输层各自负责
// "取 token"（gRPC metadata / HTTP Authorization header），校验与生效角色计算统一
// 收敛本处。先经 authenticate 校验 token 并注入用户，再用 auth.EffectiveRoles 按
// users 表 roles_override 接管状态计算生效角色并覆盖注入用户的 Roles——使后台手动
// 升降级真正生效：降权后用户即使 JWT 仍带 mars_admin，生效角色也不含管理员（对应用户
// 表读取失败回落登录身份角色：不阻断鉴权，DB 恢复后接管自动生效；空邮箱同样回落）。
func Authenticate(ctx context.Context, auth AuthBiz, token string) (context.Context, error) {
	ctx, err := authenticate(ctx, auth, token)
	if err != nil {
		return nil, err
	}
	user := MustGetUser(ctx)
	roles, err := auth.EffectiveRoles(ctx, user.Email, user.Roles)
	if err != nil {
		// 用户表读取失败（DB 抖动）：回落登录身份角色（JWT），不锁死已登录用户；
		// 手动接管在 DB 恢复后由下一次请求自动生效，不在此处吞错留无声漂移。
		return ctx, nil
	}
	user.Roles = roles
	return SetUser(ctx, user), nil
}

package auth

import (
	"context"
	"errors"

	"github.com/duc-cnzj/mars/v6/internal/biz"
)

// ctxTokenInfo 是 context 中注入用户信息所用的私有 key 类型。
// 用私有类型而非字符串作 key，避免与其他包在 context 中撞 key。
// 用空结构体【值】而非指针作 key：空结构体值恒相等（规范保证），
// 而不同函数里的 &ctxTokenInfo{} 指针是否相等是规范未定义的，依赖 gc 优化。
type ctxTokenInfo struct{}

// SetUser 将用户信息注入 context，返回携带用户的新 context。
// info 可为 nil：此时 GetUser 仍会返回错误，调用方可感知注入失效。
func SetUser(ctx context.Context, info *biz.UserInfo) context.Context {
	return context.WithValue(ctx, ctxTokenInfo{}, info)
}

// GetUser 从 context 中取出 SetUser 注入的用户信息。
// 未注入用户，或注入值为 nil（含 (*biz.UserInfo)(nil) 这种带类型 nil 指针）时，
// 返回 "user not found" 错误——不能让调用方拿到 nil 用户却以为成功。
func GetUser(ctx context.Context) (*biz.UserInfo, error) {
	if info, ok := ctx.Value(ctxTokenInfo{}).(*biz.UserInfo); ok && info != nil {
		return info, nil
	}

	return nil, errors.New("user not found")
}

// MustGetUser 的语义是"调用方保证当前上下文已注入用户"：拿不到用户即编程错误，
// 直接 panic 而非返回 nil——返回 nil 会把空指针一路传进下游，要么隐式 nil-deref，
// 要么被当作"非管理员"悄悄放行，掩盖鉴权链路本身的故障。
func MustGetUser(ctx context.Context) *biz.UserInfo {
	info, err := GetUser(ctx)
	if err != nil {
		panic(err)
	}
	return info
}

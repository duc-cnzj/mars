package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/picture"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

var _ picture.PictureServer = (*pictureSvc)(nil)

// pictureSvc 是 picture.PictureServer 的 gRPC 实现：返回随机背景图，由 NewPictureSvc 构造。
type pictureSvc struct {
	logger mlog.Logger
	picBiz biz.PictureBiz
	picture.UnimplementedPictureServer
}

// PictureSvcDeps 收口 NewPictureSvc 的构造依赖，由 wire 按字段注入。
type PictureSvcDeps struct {
	PicBiz biz.PictureBiz
	Logger mlog.Logger
}

// NewPictureSvc 收口背景图服务的构造依赖，由 wire 按字段注入。
func NewPictureSvc(deps PictureSvcDeps) picture.PictureServer {
	return &pictureSvc{picBiz: deps.PicBiz, logger: deps.Logger.WithModule("services/picture")}
}

// Background 返回登录页背景图：random 为真时随机取一张，否则取默认图。
func (p *pictureSvc) Background(ctx context.Context, req *picture.BackgroundRequest) (*picture.BackgroundResponse, error) {
	one, err := p.picBiz.Get(ctx, req.Random)
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}

	return &picture.BackgroundResponse{
		Url:       one.Url,
		Copyright: one.Copyright,
	}, nil
}

package biz

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// PictureBiz 收口图片获取业务用例，供传输层取首页图。
type PictureBiz interface {
	// Get 返回一张图片，random 为真时随机挑选。
	Get(ctx context.Context, random bool) (*PictureItem, error)
}

// PictureGetter 提供 Get 方法，用于获取图片（窄视图，供 provider 暴露取图能力）。
type PictureGetter interface {
	// Get 返回一张图片，random 为真时随机挑选。
	Get(ctx context.Context, random bool) (*PictureItem, error)
}

// PictureProvider 提供 PictureGetter，供依赖方按需取图。
type PictureProvider interface {
	// Picture 返回取图能力的窄视图。
	Picture() PictureGetter
}

type pictureBiz struct {
	logger mlog.Logger
	pp     PictureProvider
}

// NewPictureBiz 构造 picture biz。
func NewPictureBiz(logger mlog.Logger, pp PictureProvider) PictureBiz {
	return &pictureBiz{
		logger: logger.WithModule("biz/picture"),
		pp:     pp,
	}
}

// Get 获取一张图片（透传 provider）。
func (p *pictureBiz) Get(ctx context.Context, random bool) (*PictureItem, error) {
	return p.pp.Picture().Get(ctx, random)
}

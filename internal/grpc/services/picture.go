package services

import (
	"context"

	"github.com/duc-cnzj/mars-client/v4/picture"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
)

type PictureSvc struct {
	picture.UnimplementedPictureServer
}

func (p *PictureSvc) Background(ctx context.Context, req *picture.BackgroundRequest) (*picture.BackgroundResponse, error) {
	one, err := plugins.GetPicture().Get(ctx, req.Random)
	if err != nil {
		return nil, err
	}

	return &picture.BackgroundResponse{
		Url:       one.Url,
		Copyright: one.Copyright,
	}, nil
}

func (p *PictureSvc) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}

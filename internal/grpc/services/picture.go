package services

import (
	"context"

	"github.com/duc-cnzj/mars/client/picture"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
)

type Picture struct {
	picture.UnimplementedPictureServer
}

func (p *Picture) Background(ctx context.Context, req *picture.BackgroundRequest) (*picture.BackgroundResponse, error) {
	one, err := plugins.GetPicturePlugin().Get(req.Random)
	if err != nil {
		return nil, err
	}

	return &picture.BackgroundResponse{
		Url:       one.Url,
		Copyright: one.Copyright,
	}, nil
}

func (p *Picture) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}

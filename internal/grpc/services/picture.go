package services

import (
	"context"

	"google.golang.org/grpc"

	"github.com/duc-cnzj/mars/api/v4/picture"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		picture.RegisterPictureServer(s, new(pictureSvc))
	})
	RegisterEndpoint(picture.RegisterPictureHandlerFromEndpoint)
}

type pictureSvc struct {
	guest

	picture.UnimplementedPictureServer
}

func (p *pictureSvc) Background(ctx context.Context, req *picture.BackgroundRequest) (*picture.BackgroundResponse, error) {
	one, err := plugins.GetPicture().Get(ctx, req.Random)
	if err != nil {
		return nil, err
	}

	return &picture.BackgroundResponse{
		Url:       one.Url,
		Copyright: one.Copyright,
	}, nil
}

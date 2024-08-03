package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v4/picture"
	"github.com/duc-cnzj/mars/v4/internal/repo"
)

var _ picture.PictureServer = (*pictureSvc)(nil)

type pictureSvc struct {
	guest

	picRepo repo.PictureRepo
	picture.UnimplementedPictureServer
}

func NewPictureSvc(picRepo repo.PictureRepo) picture.PictureServer {
	return &pictureSvc{picRepo: picRepo}
}

func (p *pictureSvc) Background(ctx context.Context, req *picture.BackgroundRequest) (*picture.BackgroundResponse, error) {
	one, err := p.picRepo.Get(ctx, req.Random)
	if err != nil {
		return nil, err
	}

	return &picture.BackgroundResponse{
		Url:       one.Url,
		Copyright: one.Copyright,
	}, nil
}

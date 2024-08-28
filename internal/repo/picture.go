package repo

import (
	"context"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
)

type PictureRepo interface {
	Get(ctx context.Context, random bool) (*application.PictureItem, error)
}

type pictureRepo struct {
	logger mlog.Logger
	pl     application.PluginManger
}

var _ PictureRepo = (*pictureRepo)(nil)

func NewPictureRepo(logger mlog.Logger, pl application.PluginManger) PictureRepo {
	return &pictureRepo{
		logger: logger.WithModule("repo/picture"),
		pl:     pl,
	}
}

func (p *pictureRepo) Get(ctx context.Context, random bool) (*application.PictureItem, error) {
	return p.pl.Picture().Get(ctx, random)
}

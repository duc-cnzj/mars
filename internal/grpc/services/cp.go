package services

import (
	"context"
	"errors"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/models"

	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/cp"
)

type CopyToPod struct {
	cp.UnimplementedCpServer
}

func (c *CopyToPod) CopyToPod(ctx context.Context, request *cp.CopyToPodRequest) (*cp.CopyToPodResponse, error) {
	if running, reason := utils.IsPodRunning(request.Namespace, request.Pod); !running {
		return nil, errors.New(reason)
	}

	var file models.File
	if err := app.DB().First(&file, request.FileId).Error; err != nil {
		return nil, err
	}
	res, err := utils.CopyFileToPod(request.Namespace, request.Pod, request.Container, file.Path, "")
	if err != nil {
		return nil, err
	}
	file.DeleteFile()

	return &cp.CopyToPodResponse{
		PodFilePath: res.TargetDir,
		Output:      res.Output,
	}, err
}

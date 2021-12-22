package services

import (
	"context"
	"errors"
	"fmt"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/pkg/event"
	"github.com/dustin/go-humanize"

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

	app.DB().Model(&file).Updates(map[string]interface{}{
		"namespace":      request.Namespace,
		"pod":            request.Pod,
		"container":      request.Container,
		"container_path": res.ContainerPath,
	})

	AuditLog(MustGetUser(ctx).Name,
		event.ActionType_Create,
		fmt.Sprintf("上传文件到 pod: %s/%s/%s, 容器路径: '%s', 大小: %s。",
			request.Namespace,
			request.Pod,
			request.Container,
			res.ContainerPath,
			humanize.Bytes(file.Size),
		))

	return &cp.CopyToPodResponse{
		PodFilePath: res.TargetDir,
		Output:      res.ErrOut,
	}, err
}

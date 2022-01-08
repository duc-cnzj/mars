package services

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/pkg/event"
	"github.com/dustin/go-humanize"

	"github.com/duc-cnzj/mars/internal/utils"
	cp "github.com/duc-cnzj/mars/pkg/container_copy"
)

type ContainerCopy struct {
	cp.UnimplementedContainerCopyServer
}

func (c *ContainerCopy) CopyToPod(ctx context.Context, request *cp.CopyToPodRequest) (*cp.CopyToPodResponse, error) {
	if running, reason := utils.IsPodRunning(request.Namespace, request.Pod); !running {
		return nil, status.Error(codes.NotFound, reason)
	}

	var file models.File
	if err := app.DB().First(&file, request.FileId).Error; err != nil {
		return nil, err
	}
	res, err := utils.CopyFileToPod(request.Namespace, request.Pod, request.Container, file.Path, "")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

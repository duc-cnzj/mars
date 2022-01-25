package services

import (
	"context"
	"errors"
	"fmt"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	eventpb "github.com/duc-cnzj/mars/pkg/event"
	"github.com/duc-cnzj/mars/pkg/file"
	"github.com/dustin/go-humanize"
	"gopkg.in/yaml.v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type FileSvc struct {
	file.UnimplementedFileSvcServer
}

func (m *FileSvc) DiskInfo(ctx context.Context, request *file.DiskInfoRequest) (*file.DiskInfoResponse, error) {
	size, err := app.Uploader().DirSize(app.Config().UploadDir)
	if err != nil {
		return nil, err
	}
	return &file.DiskInfoResponse{
		Usage:         size,
		HumanizeUsage: humanize.Bytes(uint64(size)),
	}, nil
}

type listFiles []*file.File

type item struct {
	Name string `yaml:"name"`
	Size string `yaml:"size"`
}

func (l listFiles) PrettyYaml() string {
	var items = make([]item, 0, len(l))
	for _, f := range l {
		items = append(items, item{
			Name: f.Path,
			Size: f.HumanizeSize,
		})
	}
	marshal, _ := yaml.Marshal(items)
	return string(marshal)
}

func (m *FileSvc) DeleteUndocumentedFiles(ctx context.Context, _ *file.DeleteUndocumentedFilesRequest) (*file.DeleteUndocumentedFilesResponse, error) {
	var (
		files       []models.File
		mapFilePath = make(map[string]struct{})

		clearList = make(listFiles, 0)
	)

	app.DB().Select("ID", "Path").Find(&files)
	for _, f := range files {
		mapFilePath[f.Path] = struct{}{}
	}

	directoryFiles, _ := app.Uploader().AllDirectoryFiles(app.Config().UploadDir)
	for _, directoryFile := range directoryFiles {
		if _, ok := mapFilePath[directoryFile.Path()]; !ok {
			clearList = append(clearList, &file.File{
				Path:         directoryFile.Path(),
				HumanizeSize: humanize.Bytes(directoryFile.Size()),
				Size:         int64(directoryFile.Size()),
			})
			if err := app.Uploader().Delete(directoryFile.Path()); err != nil {
				mlog.Error(err)
			}
		}
	}
	events.AuditLog(MustGetUser(ctx).Name, eventpb.ActionType_Delete, "删除未被记录的文件", nil, clearList)

	return &file.DeleteUndocumentedFilesResponse{Files: clearList}, nil
}

func (*FileSvc) Delete(ctx context.Context, request *file.FileDeleteRequest) (*file.FileDeleteResponse, error) {
	var f = &models.File{ID: int(request.Id)}
	if err := app.DB().First(&f).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	f.DeleteFile()
	AuditLog(
		MustGetUser(ctx).Name,
		eventpb.ActionType_Delete,
		fmt.Sprintf("删除文件: '%s', 该文件由 %s 上传, 大小是 %s", f.Path, f.Username, humanize.Bytes(f.Size)))

	return &file.FileDeleteResponse{}, nil
}

func (m *FileSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}

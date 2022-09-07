package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dustin/go-humanize"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars-client/v4/file"
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/scopes"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		file.RegisterFileServer(s, new(File))
	})
	RegisterEndpoint(file.RegisterFileHandlerFromEndpoint)
}

type File struct {
	file.UnimplementedFileServer
}

func (m *File) List(ctx context.Context, request *file.ListRequest) (*file.ListResponse, error) {
	var (
		page     = int(request.Page)
		pageSize = int(request.PageSize)
		files    []models.File
		count    int64
	)
	ss := []func(*gorm.DB) *gorm.DB{
		func(db *gorm.DB) *gorm.DB {
			if !request.GetWithoutDeleted() {
				db = db.Unscoped()
			}
			return db
		},
	}
	if err := app.DB().Scopes(append(ss, scopes.Paginate(&page, &pageSize))...).Order("`id` DESC").Find(&files).Error; err != nil {
		return nil, err
	}
	app.DB().Model(&models.File{}).Scopes(ss...).Count(&count)

	var res = make([]*types.FileModel, 0, len(files))
	for _, ff := range files {
		res = append(res, ff.ProtoTransform())
	}

	return &file.ListResponse{
		Page:     int64(page),
		PageSize: int64(pageSize),
		Items:    res,
		Count:    count,
	}, nil
}

func (m *File) DiskInfo(ctx context.Context, request *file.DiskInfoRequest) (*file.DiskInfoResponse, error) {
	size, err := app.Uploader().DirSize()
	if err != nil {
		return nil, err
	}
	return &file.DiskInfoResponse{
		Usage:         size,
		HumanizeUsage: humanize.Bytes(uint64(size)),
	}, nil
}

func (m *File) ShowRecords(ctx context.Context, request *file.ShowRecordsRequest) (*file.ShowRecordsResponse, error) {
	var f models.File
	if err := app.DB().First(&f, request.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	rc, err := f.Uploader().Read(f.Path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return &file.ShowRecordsResponse{Items: transformToRecords(rc)}, nil
}

func transformToRecords(rd io.Reader) []string {
	var (
		data   []string
		lists  []string
		reader = bufio.NewReader(rd)
	)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				lists = append(lists, strings.Join(data, "\n"))
			}
			mlog.Debug(err)
			break
		}
		if strings.HasPrefix(string(line), `{"version": 2,`) {
			if len(data) > 0 {
				lists = append(lists, strings.Join(data, "\n"))
			}
			data = []string{string(line)}
		} else {
			data = append(data, string(line))
		}
	}
	return lists
}

func (*File) Delete(ctx context.Context, request *file.DeleteRequest) (*file.DeleteResponse, error) {
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
		types.EventActionType_Delete,
		fmt.Sprintf("删除文件: '%s', 该文件由 %s 上传, 大小是 %s", f.Path, f.Username, humanize.Bytes(f.Size)))

	return &file.DeleteResponse{}, nil
}

func (*File) MaxUploadSize(ctx context.Context, request *file.MaxUploadSizeRequest) (*file.MaxUploadSizeResponse, error) {
	return &file.MaxUploadSizeResponse{
		HumanizeSize: humanize.Bytes(app.Config().MaxUploadSize()),
		Bytes:        app.Config().MaxUploadSize(),
	}, nil
}

func (m *File) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if strings.Contains(fullMethodName, "MaxUploadSize") {
		return ctx, nil
	}

	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}

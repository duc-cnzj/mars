package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/duc-cnzj/mars/api/v4/file"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ file.FileServer = (*fileSvc)(nil)

type fileSvc struct {
	file.UnimplementedFileServer
	eventRepo repo.EventRepo
	fileRepo  repo.FileRepo
	logger    mlog.Logger
}

func NewFileSvc(eventRepo repo.EventRepo, fileRepo repo.FileRepo, logger mlog.Logger) file.FileServer {
	return &fileSvc{eventRepo: eventRepo, fileRepo: fileRepo, logger: logger.WithModule("services/file")}
}

func (m *fileSvc) List(ctx context.Context, request *file.ListRequest) (*file.ListResponse, error) {
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	files, pag, err := m.fileRepo.List(ctx, &repo.ListFileInput{
		Page:           page,
		PageSize:       size,
		OrderIDDesc:    lo.ToPtr(true),
		WithSoftDelete: request.WithoutDeleted,
	})
	if err != nil {
		m.logger.ErrorCtx(ctx, err)
		return nil, err
	}

	return &file.ListResponse{
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Items:    serialize.Serialize(files, transformer.FromFile),
		Count:    pag.Count,
	}, nil
}

func (m *fileSvc) DiskInfo(ctx context.Context, request *file.DiskInfoRequest) (*file.DiskInfoResponse, error) {
	size, err := m.fileRepo.DiskInfo()
	if err != nil {
		m.logger.ErrorCtx(ctx, err)
		return nil, err
	}
	return &file.DiskInfoResponse{
		Usage:         size,
		HumanizeUsage: humanize.Bytes(uint64(size)),
	}, nil
}

func (m *fileSvc) ShowRecords(ctx context.Context, request *file.ShowRecordsRequest) (*file.ShowRecordsResponse, error) {
	records, err := m.fileRepo.ShowRecords(ctx, int(request.Id))
	if err != nil {
		m.logger.ErrorCtx(ctx, err)
		return nil, err
	}
	defer records.Close()

	return &file.ShowRecordsResponse{Items: m.transformToRecords(records)}, nil
}

func (m *fileSvc) transformToRecords(rd io.Reader) []string {
	var (
		data   []string
		lists  []string
		reader = bufio.NewReader(rd)
	)
	for {
		line, err := reader.ReadString('\n')
		line = strings.Trim(line, "\n")
		if err != nil {
			if err == io.EOF {
				data = append(data, line)
				lists = append(lists, strings.Join(data, "\n"))
			}
			m.logger.Debug(err)
			break
		}
		if line != "" {
			if strings.HasPrefix(line, `{"version": 2,`) {
				if len(data) > 0 {
					lists = append(lists, strings.Join(data, "\n"))
				}
				data = []string{line}
			} else {
				data = append(data, line)
			}
		}
	}
	return lists
}

func (m *fileSvc) Delete(ctx context.Context, request *file.DeleteRequest) (*file.DeleteResponse, error) {
	f, err := m.fileRepo.GetByID(ctx, int(request.Id))
	if err != nil {
		return nil, err
	}
	if err := m.fileRepo.Delete(ctx, int(request.Id)); err != nil {
		m.logger.ErrorCtx(ctx, err)
		return nil, err
	}
	m.eventRepo.FileAuditLog(
		types.EventActionType_Delete,
		MustGetUser(ctx).Name,
		fmt.Sprintf("删除文件: '%s', 该文件由 %s 上传, 大小是 %s", f.Path, f.Username, humanize.Bytes(f.Size)),
		f.ID,
	)

	return &file.DeleteResponse{}, nil
}

func (m *fileSvc) MaxUploadSize(ctx context.Context, request *file.MaxUploadSizeRequest) (*file.MaxUploadSizeResponse, error) {
	size := m.fileRepo.MaxUploadSize()
	return &file.MaxUploadSizeResponse{
		HumanizeSize: humanize.Bytes(size),
		Bytes:        uint32(size),
	}, nil
}

func (m *fileSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if strings.Contains(fullMethodName, "MaxUploadSize") {
		return ctx, nil
	}

	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}

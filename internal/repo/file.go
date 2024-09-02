package repo

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/cache"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/ent/file"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/filters"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/duc-cnzj/mars/v5/internal/util/rand"
	"github.com/duc-cnzj/mars/v5/internal/util/serialize"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type File struct {
	ID            int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	UploadType    schematype.UploadType
	Path          string
	Size          uint64
	Username      string
	Namespace     string
	Pod           string
	Container     string
	ContainerPath string

	HumanizeSize string
}

func ToFile(file *ent.File) *File {
	if file == nil {
		return nil
	}
	return &File{
		ID:            file.ID,
		CreatedAt:     file.CreatedAt,
		UpdatedAt:     file.UpdatedAt,
		DeletedAt:     file.DeletedAt,
		UploadType:    file.UploadType,
		Path:          file.Path,
		Size:          file.Size,
		Username:      file.Username,
		Namespace:     file.Namespace,
		Pod:           file.Pod,
		Container:     file.Container,
		ContainerPath: file.ContainerPath,
	}
}

type Recorder interface {
	Resize(cols, rows uint16)
	Write(p []byte) (n int, err error)
	Close() error
	SetShell(string)
	GetShell() string
	File() *File
	Duration() time.Duration
	User() *auth.UserInfo
	Container() *Container
}

type FileRepo interface {
	MaxUploadSize() uint64
	Delete(ctx context.Context, id int) error
	ShowRecords(ctx context.Context, id int) (io.ReadCloser, error)
	DiskInfo(force bool) (int64, error)
	List(ctx context.Context, input *ListFileInput) ([]*File, *pagination.Pagination, error)
	GetByID(ctx context.Context, id int) (*File, error)
	Create(todo context.Context, input *CreateFileInput) (*File, error)
	NewDisk(disk string) uploader.Uploader
	NewFile(fpath string) (uploader.File, error)
	NewRecorder(user *auth.UserInfo, container *Container) Recorder
	Update(ctx context.Context, i *UpdateFileRequest) (*File, error)
	StreamUploadFile(ctx context.Context, input *StreamUploadFileRequest) (*File, error)
}

var _ FileRepo = (*fileRepo)(nil)

type fileRepo struct {
	logger   mlog.Logger
	uploader uploader.Uploader
	timer    timer.Timer

	cache         cache.Cache
	maxUploadSize uint64
	data          data.Data
}

func NewFileRepo(
	logger mlog.Logger,
	data data.Data,
	cache cache.Cache,
	uploader uploader.Uploader,
	timer timer.Timer,
) FileRepo {
	return &fileRepo{
		cache:         cache,
		logger:        logger.WithModule("repo/file"),
		uploader:      uploader,
		timer:         timer,
		data:          data,
		maxUploadSize: data.Config().MaxUploadSize(),
	}
}

type ListFileInput struct {
	Page, PageSize int32
	OrderIDDesc    *bool
	WithSoftDelete bool
}

func (repo *fileRepo) List(ctx context.Context, input *ListFileInput) ([]*File, *pagination.Pagination, error) {
	var db = repo.data.DB()
	queryCtx := ctx
	if input.WithSoftDelete {
		queryCtx = mixin.SkipSoftDelete(queryCtx)
	}
	query := db.File.Query().
		Where(filters.IfOrderByIDDesc(input.OrderIDDesc))
	files := query.Clone().
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).
		AllX(queryCtx)
	count := query.Clone().
		CountX(queryCtx)

	return serialize.Serialize(files, ToFile), pagination.NewPagination(input.Page, input.PageSize, count), nil
}

type CreateFileInput struct {
	Path       string
	Username   string
	Size       uint64
	UploadType schematype.UploadType

	Namespace string
	Pod       string
	Container string
}

func (repo *fileRepo) Create(todo context.Context, input *CreateFileInput) (*File, error) {
	var db = repo.data.DB()
	save, err := db.File.Create().
		SetPath(input.Path).
		SetUsername(input.Username).
		SetNamespace(input.Namespace).
		SetPod(input.Pod).
		SetContainer(input.Container).
		SetSize(input.Size).
		SetUploadType(input.UploadType).
		Save(todo)
	return ToFile(save), err
}

func (repo *fileRepo) GetByID(ctx context.Context, id int) (*File, error) {
	var db = repo.data.DB()
	first, err := db.File.Query().Where(file.ID(id)).First(ctx)
	return ToFile(first), err
}

type UpdateFileRequest struct {
	ID            int
	ContainerPath string
	Namespace     string
	Pod           string
	Container     string
}

func (repo *fileRepo) Update(ctx context.Context, i *UpdateFileRequest) (*File, error) {
	var db = repo.data.DB()
	first, err := db.File.UpdateOneID(i.ID).
		SetContainerPath(i.ContainerPath).
		SetNamespace(i.Namespace).
		SetPod(i.Pod).
		SetContainer(i.Container).
		Save(ctx)
	return ToFile(first), err
}

func (repo *fileRepo) MaxUploadSize() uint64 {
	return repo.maxUploadSize
}

func (repo *fileRepo) Delete(ctx context.Context, id int) error {
	var db = repo.data.DB()
	file, err := db.File.Query().Where(file.ID(id)).First(ctx)
	if err != nil {
		return err
	}
	if err = db.File.DeleteOneID(id).Exec(ctx); err != nil {
		return err
	}
	return repo.uploader.Delete(file.Path)
}

func (repo *fileRepo) ShowRecords(ctx context.Context, id int) (io.ReadCloser, error) {
	var db = repo.data.DB()
	file, err := db.File.Query().Where(file.ID(id)).First(ctx)
	if err != nil {
		return nil, err
	}
	var up uploader.Uploader
	switch file.UploadType {
	case repo.uploader.LocalUploader().Type():
		up = repo.uploader.LocalUploader()
	case repo.uploader.Type():
		up = repo.uploader
	}
	return up.Read(file.Path)
}

var DirSizeCacheSeconds = int((15 * time.Minute).Seconds())

func (repo *fileRepo) DiskInfo(force bool) (int64, error) {
	remember, err := repo.cache.Remember(cache.NewKey("dir-size"), DirSizeCacheSeconds, func() ([]byte, error) {
		size, err := repo.uploader.DirSize()
		return int64ToByte(size), err
	}, force)

	return byteToInt64(remember), err
}

func (repo *fileRepo) NewRecorder(user *auth.UserInfo, container *Container) Recorder {
	return &recorder{
		fileRepo:      repo,
		logger:        repo.logger,
		timer:         repo.timer,
		container:     container,
		user:          user,
		localUploader: repo.uploader.LocalUploader(),
		uploader:      repo.uploader,
	}
}

func (repo *fileRepo) NewFile(fpath string) (uploader.File, error) {
	return repo.uploader.NewFile(fpath)
}

func (repo *fileRepo) NewDisk(disk string) uploader.Uploader {
	return repo.uploader.Disk(disk)
}

const StreamUploadFileDisk = "grpc_upload"

type StreamUploadFileRequest struct {
	Namespace, Pod, Container string
	Username                  string
	FileName                  string
	FileData                  chan []byte
}

// StreamUploadFile
// 1. 用户上传文件
func (repo *fileRepo) StreamUploadFile(ctx context.Context, input *StreamUploadFileRequest) (*File, error) {
	tracer := otel.Tracer("")
	ctx, span := tracer.Start(ctx, "fileRepo/StreamUploadFile")
	defer span.End()
	span.SetAttributes(
		attribute.Key("username").String(input.Username),
		attribute.Key("namespace").String(input.Namespace),
		attribute.Key("pod").String(input.Pod),
		attribute.Key("container").String(input.Container),
		attribute.Key("file_name").String(input.FileName),
	)
	disk := repo.uploader.Disk(StreamUploadFileDisk)
	// 某个用户/那天/时间/文件名称
	// duc/2006-01-02/15-03-04-random-str/xxx.tgz
	now := repo.timer.Now()
	p := fmt.Sprintf("%s/%s/%s/%s",
		input.Username,
		now.Format("2006-01-02"),
		fmt.Sprintf("%s-%s", now.Format("15-04-05"), rand.String(20)),
		filepath.Base(input.FileName))
	fpath := disk.AbsolutePath(p)
	err := disk.MkDir(filepath.Dir(fpath), true)
	if err != nil {
		return nil, err
	}
	newFile, err := repo.uploader.NewFile(fpath)
	if err != nil {
		return nil, err
	}
	defer newFile.Close()
	for data := range input.FileData {
		if _, err := newFile.Write(data); err != nil {
			return nil, err
		}
	}
	stat, _ := newFile.Stat()
	return repo.Create(ctx, &CreateFileInput{
		Path:       newFile.Name(),
		Username:   input.Username,
		Size:       uint64(stat.Size()),
		UploadType: disk.Type(),
		Namespace:  input.Namespace,
		Pod:        input.Pod,
		Container:  input.Container,
	})
}

type recorder struct {
	sync.RWMutex

	file     *File
	fileRepo FileRepo

	logger    mlog.Logger
	timer     timer.Timer
	container *Container
	f         uploader.File
	startTime time.Time

	user *auth.UserInfo
	once sync.Once

	buffer *bufio.Writer

	shellMu sync.RWMutex
	shell   string

	rcMu       sync.RWMutex
	cols, rows uint16

	localUploader uploader.Uploader
	uploader      uploader.Uploader
}

func (r *recorder) Container() *Container {
	return r.container
}

func (r *recorder) User() *auth.UserInfo {
	return r.user
}

func (r *recorder) Duration() time.Duration {
	return time.Since(r.startTime)
}

func (r *recorder) File() *File {
	r.Lock()
	defer r.Unlock()
	return r.file
}

func max[T int | uint16 | uint64](a, b T) T {
	if a < b {
		return b
	}
	return a
}

func (r *recorder) GetShell() string {
	r.shellMu.RLock()
	defer r.shellMu.RUnlock()
	return r.shell
}

func (r *recorder) SetShell(sh string) {
	r.shellMu.Lock()
	defer r.shellMu.Unlock()
	r.shell = sh
}

func (r *recorder) Resize(cols, rows uint16) {
	r.HeadLineColRow(cols, rows)
}

func (r *recorder) HeadLineColRow(cols, rows uint16) {
	r.rcMu.Lock()
	defer r.rcMu.Unlock()
	r.cols = max(r.cols, cols)
	r.rows = max(r.rows, rows)
}

func (r *recorder) Write(data []byte) (n int, err error) {
	r.Lock()
	defer r.Unlock()
	r.once.Do(func() {
		var file uploader.File
		file, err = r.localUploader.Disk("tmp").NewFile(fmt.Sprintf("%s/%s/%s",
			r.user.Name,
			r.timer.Now().Format("2006-01-02"),
			fmt.Sprintf("recorder-%s-%s-%s-%s.cast.tmp", r.container.Namespace, r.container.Pod, r.container.Container, rand.String(20))))
		if err != nil {
			return
		}
		r.f = file
		r.buffer = bufio.NewWriterSize(r.f, 1024*20)
		r.startTime = r.timer.Now()
		r.HeadLineColRow(106, 25)
	})
	if err != nil {
		return 0, err
	}
	marshal, _ := json.Marshal(string(data))
	_, err = r.buffer.WriteString(fmt.Sprintf(writeLine, float64(time.Since(r.startTime).Microseconds())/1000000, string(marshal)))
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

var (
	startLine = "{\"version\": 2, \"width\": %d, \"height\": %d, \"timestamp\": %d, \"env\": {\"SHELL\": \"%s\", \"TERM\": \"xterm-256color\"}}\n"
	writeLine = "[%.6f, \"o\", %s]\n"
)

func (r *recorder) Close() error {
	r.Lock()
	defer r.Unlock()
	defer r.logger.Info("recorder close")
	var (
		err error

		localUploader = r.localUploader
		uploader      = r.uploader
	)
	if r.buffer == nil || r.startTime.IsZero() {
		return nil
	}
	if err := r.buffer.Flush(); err != nil {
		r.logger.Error(err)
		return err
	}

	upFile, err := uploader.Disk("shell").NewFile(fmt.Sprintf("%s/%s/%s",
		r.user.Name,
		r.timer.Now().Format("2006-01-02"),
		fmt.Sprintf("recorder-%s-%s-%s-%s.cast", r.container.Namespace, r.container.Pod, r.container.Container, rand.String(20))))
	if err != nil {
		r.logger.Error(err)
		return err
	}
	func() {
		defer func() {
			r.f.Close()
			localUploader.Delete(r.f.Name())
		}()
		r.f.Seek(0, 0)
		func() {
			r.rcMu.RLock()
			defer r.rcMu.RUnlock()
			upFile.WriteString(fmt.Sprintf(startLine, r.cols, r.rows, r.startTime.Unix(), r.shell))
		}()
		if _, err := io.Copy(upFile, r.f); err != nil {
			r.logger.Error(err)
		}
	}()

	stat, e := upFile.Stat()
	if e != nil {
		upFile.Close()
		uploader.Delete(upFile.Name())
		r.logger.Error(e)
		return e
	}
	var emptyFile bool = true
	defer func() {
		err = upFile.Close()
		if emptyFile {
			uploader.Delete(upFile.Name())
		}
	}()
	if stat.Size() > 0 {
		r.file, err = r.fileRepo.Create(context.TODO(), &CreateFileInput{
			UploadType: uploader.Type(),
			Path:       upFile.Name(),
			Size:       uint64(stat.Size()),
			Username:   r.user.Name,
			Namespace:  r.container.Namespace,
			Pod:        r.container.Pod,
			Container:  r.container.Container,
		})
		if err != nil {
			r.logger.Error(err)
			return err
		}

		emptyFile = false
	}
	return err
}

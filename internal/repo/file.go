package repo

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/file"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v4/internal/filters"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
	"github.com/duc-cnzj/mars/v4/internal/util/rand"
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
)

type Recorder interface {
	Resize(cols, rows uint16)
	Write(data string) (err error)
	Close() error
	SetShell(string)
	GetShell() string
}

type FileRepo interface {
	MaxUploadSize() uint64
	Delete(ctx context.Context, id int) error
	ShowRecords(ctx context.Context, id int) (io.ReadCloser, error)
	DiskInfo() (int64, error)
	List(ctx context.Context, input *ListFileInput) ([]*ent.File, *pagination.Pagination, error)
	GetByID(ctx context.Context, id int) (*ent.File, error)
	Create(todo context.Context, input *CreateFileInput) (*ent.File, error)
	NewDisk(disk string) uploader.Uploader
	NewFile(fpath string) (uploader.File, error)
	NewRecorder(action types.EventActionType, user *auth.UserInfo, container *Container) Recorder
}

var _ FileRepo = (*fileRepo)(nil)

type fileRepo struct {
	logger   mlog.Logger
	uploader uploader.Uploader
	timer    timer.Timer
	crRepo   CronRepo

	maxUploadSize uint64
	data          data.Data
}

func NewFileRepo(
	crRepo CronRepo,
	logger mlog.Logger,
	data data.Data,
	uploader uploader.Uploader,
	timer timer.Timer,
) FileRepo {
	return &fileRepo{
		crRepo:        crRepo,
		logger:        logger,
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

func (repo *fileRepo) List(ctx context.Context, input *ListFileInput) ([]*ent.File, *pagination.Pagination, error) {
	var db = repo.data.DB()
	queryCtx := context.TODO()
	if input.WithSoftDelete {
		queryCtx = mixin.SkipSoftDelete(queryCtx)
	}
	query := db.File.Query().Where(filters.IfOrderByDesc("id")(input.OrderIDDesc))
	files := query.Clone().
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).AllX(queryCtx)
	count := query.Clone().CountX(queryCtx)

	return files, pagination.NewPagination(input.Page, input.PageSize, count), nil
}

type CreateFileInput struct {
	Path       string
	Username   string
	Size       uint64
	UploadType schematype.UploadType
}

func (repo *fileRepo) Create(todo context.Context, input *CreateFileInput) (*ent.File, error) {
	var db = repo.data.DB()
	return db.File.Create().
		SetPath(input.Path).
		SetUsername(input.Username).
		SetSize(input.Size).
		SetUploadType(input.UploadType).
		Save(todo)
}

func (repo *fileRepo) GetByID(ctx context.Context, id int) (*ent.File, error) {
	var db = repo.data.DB()
	return db.File.Query().Where(file.ID(id)).First(ctx)
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

func (repo *fileRepo) DiskInfo() (int64, error) {
	return repo.crRepo.DiskInfo()
}

func (repo *fileRepo) NewRecorder(action types.EventActionType, user *auth.UserInfo, container *Container) Recorder {
	return &recorder{
		RWMutex:       sync.RWMutex{},
		data:          repo.data,
		logger:        repo.logger,
		action:        action,
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

type recorder struct {
	sync.RWMutex

	data      data.Data
	logger    mlog.Logger
	action    types.EventActionType
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

func (r *recorder) Write(data string) (err error) {
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
		return err
	}
	marshal, _ := json.Marshal(data)
	_, err = r.buffer.WriteString(fmt.Sprintf(writeLine, float64(time.Since(r.startTime).Microseconds())/1000000, string(marshal)))
	return err
}

var (
	startLine = "{\"version\": 2, \"width\": %d, \"height\": %d, \"timestamp\": %d, \"env\": {\"SHELL\": \"%s\", \"TERM\": \"xterm-256color\"}}\n"
	writeLine = "[%.6f, \"o\", %s]\n"
)

func (r *recorder) Close() error {
	r.RLock()
	defer r.RUnlock()
	var (
		err error

		db            = r.data.DB()
		localUploader = r.localUploader
		uploader      = r.uploader
	)
	if r.buffer == nil || r.startTime.IsZero() {
		return nil
	}
	r.buffer.Flush()

	upFile, _ := uploader.Disk("shell").NewFile(fmt.Sprintf("%s/%s/%s",
		r.user.Name,
		r.timer.Now().Format("2006-01-02"),
		fmt.Sprintf("recorder-%s-%s-%s-%s.cast", r.container.Namespace, r.container.Pod, r.container.Container, rand.String(20))))

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
		io.Copy(upFile, r.f)
	}()

	stat, e := upFile.Stat()
	if e != nil {
		upFile.Close()
		uploader.Delete(upFile.Name())
		r.logger.Error(e)
		return e
	}
	var emptyFile bool = true
	if stat.Size() > 0 {
		file, _ := db.File.Create().
			SetUploadType(uploader.Type()).
			SetPath(upFile.Name()).
			SetSize(uint64(stat.Size())).
			SetUsername(r.user.Name).
			SetNamespace(r.container.Namespace).
			SetPod(r.container.Pod).
			SetContainer(r.container.Container).
			Save(context.TODO())

		db.Event.Create().SetAction(r.action).SetUsername(r.user.Name).
			SetMessage(fmt.Sprintf("用户进入容器执行命令，container: '%s', namespace: '%s', pod： '%s'", r.container.Container, r.container.Namespace, r.container.Pod)).
			SetFileID(file.ID).
			SetDuration(date.HumanDuration(time.Since(r.startTime))).
			Save(context.TODO())
		emptyFile = false
	}
	err = upFile.Close()
	if emptyFile {
		uploader.Delete(upFile.Name())
	}
	return err
}

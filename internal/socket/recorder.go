package socket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/internal/utils/date"
)

var (
	startLine = "{\"version\": 2, \"width\": %d, \"height\": %d, \"timestamp\": %d, \"env\": {\"SHELL\": \"%s\", \"TERM\": \"xterm-256color\"}}\n"
	writeLine = "[%.6f, \"o\", %s]\n"
)

type timer interface {
	Now() time.Time
}

type realTimer struct{}

func (r realTimer) Now() time.Time {
	return time.Now()
}

type Recorder struct {
	sync.RWMutex
	timer     timer
	container contracts.Container
	f         contracts.File
	startTime time.Time

	t    *MyPtyHandler
	once sync.Once

	buffer *bufio.Writer

	shellMu sync.RWMutex
	shell   string

	rcMu       sync.RWMutex
	cols, rows uint16
}

func max[T int | uint16 | uint64](a, b T) T {
	if a < b {
		return b
	}
	return a
}

func (r *Recorder) GetShell() string {
	r.shellMu.RLock()
	defer r.shellMu.RUnlock()
	return r.shell
}

func (r *Recorder) SetShell(sh string) {
	r.shellMu.Lock()
	defer r.shellMu.Unlock()
	r.shell = sh
}

func (r *Recorder) Resize(cols, rows uint16) {
	r.HeadLineColRow(cols, rows)
}

func (r *Recorder) HeadLineColRow(cols, rows uint16) {
	r.rcMu.Lock()
	defer r.rcMu.Unlock()
	r.cols = max(r.cols, cols)
	r.rows = max(r.rows, rows)
}

func (r *Recorder) Write(data string) (err error) {
	r.Lock()
	defer r.Unlock()
	r.once.Do(func() {
		var file contracts.File
		file, err = app.LocalUploader().Disk("tmp").NewFile(fmt.Sprintf("%s/%s/%s",
			r.t.conn.GetUser().Name,
			r.timer.Now().Format("2006-01-02"),
			fmt.Sprintf("recorder-%s-%s-%s-%s.cast.tmp", r.t.Container().Namespace, r.t.Container().Pod, r.t.Container().Container, utils.RandomString(20))))
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

func (r *Recorder) Close() error {
	r.RLock()
	defer r.RUnlock()
	var (
		err           error
		localUploader = app.LocalUploader()
		uploader      = app.Uploader()
	)
	if r.buffer == nil || r.startTime.IsZero() {
		return nil
	}
	r.buffer.Flush()

	upFile, _ := uploader.Disk("shell").NewFile(fmt.Sprintf("%s/%s/%s",
		r.t.conn.GetUser().Name,
		r.timer.Now().Format("2006-01-02"),
		fmt.Sprintf("recorder-%s-%s-%s-%s.cast", r.t.Container().Namespace, r.t.Container().Pod, r.t.Container().Container, utils.RandomString(20))))

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
		mlog.Error(e)
		return e
	}
	var emptyFile bool = true
	if stat.Size() > 0 {
		file := &models.File{
			UploadType: uploader.Type(),
			Path:       upFile.Name(),
			Size:       uint64(stat.Size()),
			Username:   r.t.conn.GetUser().Name,
			Namespace:  r.container.Namespace,
			Pod:        r.container.Pod,
			Container:  r.container.Container,
		}
		app.DB().Create(file)
		var emodal = models.Event{
			Action:   uint8(types.EventActionType_Shell),
			Username: r.t.conn.GetUser().Name,
			Message:  fmt.Sprintf("用户进入容器执行命令，container: '%s', namespace: '%s', pod： '%s'", r.container.Container, r.container.Namespace, r.container.Pod),
			FileID:   &file.ID,
			Duration: date.HumanDuration(time.Since(r.startTime)),
		}

		app.DB().Create(&emodal)
		emptyFile = false
	}
	err = upFile.Close()
	if emptyFile {
		uploader.Delete(upFile.Name())
	}
	return err
}

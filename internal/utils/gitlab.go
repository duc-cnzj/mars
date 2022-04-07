package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
)

func DownloadFiles(pid any, commit string, files []string) (string, func(), error) {
	id := fmt.Sprintf("%v", pid)
	dir := fmt.Sprintf("mars_tmp_%s", RandomString(10))
	if err := app.Uploader().MkDir(dir, false); err != nil {
		return "", nil, err
	}

	return DownloadFilesToDir(id, commit, files, app.Uploader().AbsolutePath(dir))
}

func DownloadFilesToDir(pid any, commit string, files []string, dir string) (string, func(), error) {
	uploader := app.Uploader()

	wg := &sync.WaitGroup{}
	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			defer HandlePanic("DownloadFilesToDir")
			raw, err := plugins.GetGitServer().GetFileContentWithSha(fmt.Sprintf("%v", pid), commit, file)
			if err != nil {
				mlog.Error(err)
			}
			fp := filepath.Join(dir, file)
			if _, err := uploader.Put(fp, strings.NewReader(raw)); err != nil {
				mlog.Errorf("[DownloadFilesToDir]: err '%s'", err.Error())
			}
		}(file)
	}
	wg.Wait()

	return dir, func() {
		err := app.Uploader().DeleteDir(dir)
		if err != nil {
			mlog.Warning(err)
			return
		}
		mlog.Debug("remove " + dir)
	}, nil
}

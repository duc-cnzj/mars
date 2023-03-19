package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	app "github.com/duc-cnzj/mars/v4/internal/app/helper"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
	"github.com/duc-cnzj/mars/v4/internal/utils/recovery"
)

func DownloadFiles(pid any, commit string, files []string) (string, func(), error) {
	var localUploader = app.LocalUploader()
	id := fmt.Sprintf("%v", pid)
	dir := fmt.Sprintf("mars_tmp_%s", RandomString(10))
	if err := localUploader.MkDir(dir, false); err != nil {
		return "", nil, err
	}

	return DownloadFilesToDir(id, commit, files, localUploader.AbsolutePath(dir))
}

func DownloadFilesToDir(pid any, commit string, files []string, dir string) (string, func(), error) {
	var localUploader = app.LocalUploader()
	wg := &sync.WaitGroup{}
	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			defer recovery.HandlePanic("DownloadFilesToDir")
			raw, err := plugins.GetGitServer().GetFileContentWithSha(fmt.Sprintf("%v", pid), commit, file)
			if err != nil {
				mlog.Error(err)
			}
			localPath := filepath.Join(dir, file)
			if _, err := localUploader.Put(localPath, strings.NewReader(raw)); err != nil {
				mlog.Errorf("[DownloadFilesToDir]: err '%s'", err.Error())
			}
		}(file)
	}
	wg.Wait()

	return dir, func() {
		err := localUploader.DeleteDir(dir)
		if err != nil {
			mlog.Warning(err)
			return
		}
		mlog.Debug("remove " + dir)
	}, nil
}

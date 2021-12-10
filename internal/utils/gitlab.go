package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/duc-cnzj/mars/internal/plugins"

	"github.com/duc-cnzj/mars/internal/mlog"
)

func DownloadFiles(pid interface{}, commit string, files []string) (string, func(), error) {
	id := fmt.Sprintf("%v", pid)
	dir, err := os.MkdirTemp("", "mars_tmp_*")
	if err != nil {
		return "", nil, err
	}

	return DownloadFilesToDir(id, commit, files, dir)
}

func DownloadFilesToDir(pid interface{}, commit string, files []string, dir string) (string, func(), error) {
	wg := &sync.WaitGroup{}
	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			raw, err := plugins.GetGitServer().GetFileContentWithSha(fmt.Sprintf("%v", pid), commit, file)
			if err != nil {
				mlog.Error(err)
			}
			fp := filepath.Join(dir, file)
			s := filepath.Dir(fp)
			if !FileExists(s) {
				os.MkdirAll(s, 0700)
			}
			openFile, err := os.OpenFile(fp, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
			if err != nil {
				mlog.Error(err)
				return
			}
			defer openFile.Close()
			openFile.Write([]byte(raw))
		}(file)
	}
	wg.Wait()

	return dir, func() {
		if FileExists(dir) {
			err := os.RemoveAll(dir)
			if err != nil {
				mlog.Warning(err)
				return
			}
			mlog.Debug("remove " + dir)
		}
	}, nil
}

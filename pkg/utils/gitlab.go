package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/duc-cnzj/mars/pkg/mlog"
	"github.com/xanzy/go-gitlab"
)

func GetDirectoryFiles(pid interface{}, commit string, path string) []string {
	var files []string

	opt := &gitlab.ListTreeOptions{
		Path:      gitlab.String(path),
		Recursive: gitlab.Bool(true),
	}
	if commit != "" {
		opt.Ref = gitlab.String(commit)
	}

	tree, _, _ := GitlabClient().Repositories.ListTree(pid, opt)

	for _, node := range tree {
		if node.Type == "blob" {
			files = append(files, node.Path)
		}
	}

	return files
}

var TmpDirName = "/tmp/gitlab_files"

func DownloadFiles(pid interface{}, commit string, files []string) (string, func()) {
	dir := fmt.Sprintf("%s/tmp_%s/", TmpDirName, RandomString(10))
	os.MkdirAll(dir, 0755)
	wg := &sync.WaitGroup{}
	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			opt := gitlab.GetRawFileOptions{}
			if commit != "" {
				opt.Ref = gitlab.String(commit)
			}
			raw, _, err := GitlabClient().RepositoryFiles.GetRawFile(pid, file, &opt)
			if err != nil {
				mlog.Error(err)
			}
			fp := filepath.Join(dir, file)
			s := filepath.Dir(fp)
			if !FileExists(s) {
				os.MkdirAll(s, 0755)
			}
			openFile, err := os.OpenFile(fp, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
			if err != nil {
				mlog.Error(err)
				return
			}
			defer openFile.Close()
			openFile.Write(raw)
		}(file)
	}
	wg.Wait()

	return dir, func() {
		if FileExists(dir) && strings.HasPrefix(dir, "/tmp") {
			err := os.RemoveAll(dir)
			if err != nil {
				mlog.Warning(err)
				return
			}
			mlog.Debug("remove " + dir)
		}
	}
}

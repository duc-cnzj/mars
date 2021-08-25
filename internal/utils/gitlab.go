package utils

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/xanzy/go-gitlab"
)

func GetDirectoryFiles(pid interface{}, commit string, path string) []string {
	var files []string

	// TODO: 坑, GitlabClient().Repositories.ListTree 带分页！！凸(艹皿艹 )
	opt := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
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

func DownloadFiles(pid interface{}, commit string, files []string) (string, func()) {
	dir, err := os.MkdirTemp("", "mars_tmp_*")
	if err != nil {
		return "", nil
	}

	return DownloadFilesToDir(pid, commit, files, dir)
}

func DownloadFilesToDir(pid interface{}, commit string, files []string, dir string) (string, func()) {
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
				os.MkdirAll(s, 0700)
			}
			openFile, err := os.OpenFile(fp, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
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
		if FileExists(dir) {
			err := os.RemoveAll(dir)
			if err != nil {
				mlog.Warning(err)
				return
			}
			mlog.Debug("remove " + dir)
		}
	}
}

func GetAllBranches(pid interface{}, options ...gitlab.RequestOptionFunc) ([]*gitlab.Branch, error) {
	var branches []*gitlab.Branch
	page := 1
	for page != -1 {
		b, r, e := GitlabClient().Branches.ListBranches(pid, &gitlab.ListBranchesOptions{ListOptions: gitlab.ListOptions{PerPage: 100, Page: page}}, options...)
		if e != nil {
			return nil, e
		}
		nextPage := r.Header.Get("x-next-page")
		if nextPage == "" {
			page = -1
		} else {
			page, _ = strconv.Atoi(nextPage)
		}
		branches = append(branches, b...)
	}

	return branches, nil
}

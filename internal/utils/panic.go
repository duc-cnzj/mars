package utils

import "github.com/duc-cnzj/mars/internal/mlog"

func HandlePanic()  {
	err := recover()
	if err != nil {
		mlog.Errorf("[Panic]: %v", err)
	}
}

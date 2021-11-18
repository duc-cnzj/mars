package utils

import (
	"github.com/duc-cnzj/mars/internal/mlog"
)

func HandlePanic(title string) {
	err := recover()
	if err != nil {
		mlog.Errorf("[Panic]: title: %v, err: %v", title, err)
	}
}

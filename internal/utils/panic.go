package utils

import (
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
)

func HandlePanic(title string) {
	err := recover()
	if err != nil {
		mlog.Errorf("[Panic]: title: %v, err: %v", title, err)
		if app.App().IsDebug() {
			panic(err)
		}
	}
}

package contracts

//go:generate mockgen -destination ../mock/mock_log.go -package mock github.com/duc-cnzj/mars/internal/contracts LoggerInterface

type LoggerInterface interface {
	Debug(v ...any)
	Debugf(format string, v ...any)
	Warning(v ...any)
	Warningf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
}

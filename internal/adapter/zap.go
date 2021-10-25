package adapter

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	app   contracts.ApplicationInterface
	sugar *zap.SugaredLogger
}

func NewZapLogger(app contracts.ApplicationInterface) *ZapLogger {
	var (
		logger *zap.Logger
		cfg    zap.Config
	)

	opts := []zap.Option{zap.AddStacktrace(zapcore.ErrorLevel), zap.AddCallerSkip(2)}
	if app.IsDebug() {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	logger, _ = cfg.Build(opts...)

	app.RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
		logger.Info("zap synchronized.")
		logger.Sync()
	})

	return &ZapLogger{app: app, sugar: logger.Sugar()}
}

func (z *ZapLogger) Debug(v ...interface{}) {
	z.sugar.Debug(v...)
}

func (z *ZapLogger) Debugf(format string, v ...interface{}) {
	z.sugar.Debugf(format, v...)
}

func (z *ZapLogger) Warning(v ...interface{}) {
	z.sugar.Warn(v...)
}

func (z *ZapLogger) Warningf(format string, v ...interface{}) {
	z.sugar.Warnf(format, v...)
}

func (z *ZapLogger) Info(v ...interface{}) {
	z.sugar.Info(v...)
}

func (z *ZapLogger) Infof(format string, v ...interface{}) {
	z.sugar.Infof(format, v...)
}

func (z *ZapLogger) Error(v ...interface{}) {
	z.sugar.Error(v...)
}

func (z *ZapLogger) Errorf(format string, v ...interface{}) {
	z.sugar.Errorf(format, v...)
}

func (z *ZapLogger) Fatal(v ...interface{}) {
	z.sugar.Fatal(v...)
}

func (z *ZapLogger) Fatalf(format string, v ...interface{}) {
	z.sugar.Fatalf(format, v...)
}

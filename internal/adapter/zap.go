package adapter

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

type zapLogger struct {
	app   contracts.ApplicationInterface
	sugar *zap.SugaredLogger
}

// NewZapLogger impl contracts.LoggerInterface.
func NewZapLogger(app contracts.ApplicationInterface) contracts.LoggerInterface {
	var (
		logger *zap.Logger
		cfg    zap.Config
	)

	opts := []zap.Option{zap.AddStacktrace(zapcore.ErrorLevel), zap.AddCallerSkip(2)}
	if app.IsDebug() {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.CallerKey = "file"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	logger, _ = cfg.Build(opts...)

	app.RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
		logger.Info("zap synchronized.")
		logger.Sync()
	})

	return &zapLogger{app: app, sugar: logger.Sugar()}
}

// Debug print debug msg
func (z *zapLogger) Debug(v ...any) {
	z.sugar.Debug(v...)
}

// Debugf printf debug msg
func (z *zapLogger) Debugf(format string, v ...any) {
	z.sugar.Debugf(format, v...)
}

// Warning print Warning msg
func (z *zapLogger) Warning(v ...any) {
	z.sugar.Warn(v...)
}

// Warningf prints Warning msg
func (z *zapLogger) Warningf(format string, v ...any) {
	z.sugar.Warnf(format, v...)
}

// Info print info msg
func (z *zapLogger) Info(v ...any) {
	z.sugar.Info(v...)
}

// Infof printf info msg
func (z *zapLogger) Infof(format string, v ...any) {
	z.sugar.Infof(format, v...)
}

// Error print err msg
func (z *zapLogger) Error(v ...any) {
	z.sugar.Error(v...)
}

// Errorf printf err msg
func (z *zapLogger) Errorf(format string, v ...any) {
	z.sugar.Errorf(format, v...)
}

// Fatal fatal err.
func (z *zapLogger) Fatal(v ...any) {
	z.sugar.Fatal(v...)
}

// Fatalf fatalf err.
func (z *zapLogger) Fatalf(format string, v ...any) {
	z.sugar.Fatalf(format, v...)
}

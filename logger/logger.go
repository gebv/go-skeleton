package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level zapcore.Level) (logger *zap.Logger) {
	atom := zap.NewAtomicLevelAt(level)
	encoderCfg := zap.NewDevelopmentEncoderConfig()

	logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))
	return
}

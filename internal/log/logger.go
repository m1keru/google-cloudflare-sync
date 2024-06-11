package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SharedLogger - Logger Instance
var SharedLogger *zap.SugaredLogger
var LogLevel zap.AtomicLevel

// InitLogger - InitLogger
func InitLogger(logLevel zapcore.Level) {
	LogLevel = zap.NewAtomicLevel()
	LogLevel.SetLevel(logLevel)
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel)
	SharedLogger = zap.New(core, zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()
}

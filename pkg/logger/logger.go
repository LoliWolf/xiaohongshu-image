package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 定义 Asia/Shanghai 时区
var shanghaiTZ *time.Location

func init() {
	var err error
	shanghaiTZ, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果加载失败，使用 UTC+8
		shanghaiTZ = time.FixedZone("CST", 8*60*60)
	}
}

func New(mode string) (*zap.Logger, error) {
	var config zap.Config

	if mode == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 使用自定义时区编码器（Asia/Shanghai GMT+8）
	config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(shanghaiTZ).Format("2006-01-02T15:04:05.000Z07:00"))
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func NewProduction() (*zap.Logger, error) {
	return zap.NewProduction()
}

func NewDevelopment() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

func NewJSONEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(shanghaiTZ).Format("2006-01-02T15:04:05.000Z07:00"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func NewConsoleEncoder() zapcore.Encoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(shanghaiTZ).Format("2006-01-02T15:04:05.000Z07:00"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func NewSyncer(filepath string) (zapcore.WriteSyncer, error) {
	if filepath == "" {
		return os.Stdout, nil
	}
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(file), nil
}

package logger

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

const (
	// LogDir ...
	LogDir = "/var/log/httpserver"
	// LogRotateDaysEnvKey 日志保留最大时长环境变量名
	LogRotateDaysEnvKey = "LOG_ROTATE_DAYS"
	// DefaultLogRotateDays 日志保留最大时长默认天数
	DefaultLogRotateDays = 3
	// MinLogRotateDays 日志保留最大时长最小天数
	MinLogRotateDays = 1
)

var (
	// 服务名称
	gServiceName = ""
)

// Factory is the default logging wrapper that can create
// logger instances either for a given Context or context-less.
type Factory struct {
	logger *zap.Logger
	ctx    context.Context
}

var f *Factory

var once sync.Once

// Bg creates a context-unaware logger.
func (b *Factory) Bg() Logger {
	return logger(*b)
}

// For returns a context-aware Logger. If the context
// echo-ed into the span.
func (b *Factory) For(ctx context.Context) Logger {
	b.ctx = ctx
	return b.Bg()
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (b *Factory) With(fields ...zapcore.Field) Factory {
	return Factory{logger: b.logger.With(fields...)}
}

var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

// Config ...
type Config struct {
	ServiceName string
	LogLevel    string
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

// Init initialize the logger singleton
// serviceIDs 标记服务唯一ID，具体见common/constant/constant.go文件定义
func Init(config *Config) {
	once.Do(func() {
		serviceName := config.ServiceName
		gServiceName = config.ServiceName
		fileName := LogDir + "/" + serviceName + ".%Y%m%d%H" + ".log"

		maxAgeHours := time.Hour * time.Duration(getLogRotateDays()*24)
		timeRotateWriter, _ := rotatelogs.New(
			fileName,
			rotatelogs.WithMaxAge(maxAgeHours),
			rotatelogs.WithRotationTime(24*time.Hour),
		)
		syncer := zapcore.AddSync(timeRotateWriter)
		level := getLoggerLevel(config.LogLevel)
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(getEncoderConfig()),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), syncer),
			zap.NewAtomicLevelAt(level),
		)

		zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
		_ = &Factory{
			logger: zapLogger,
			ctx:    context.TODO(),
		}
	})
}

// 自定义zap encoder
// 输出格式：时间|服务标识ID|日志等级|TraceID|时间 日志内容
// 如： 20171122174631|8000000|WARN||2017-11-22 17:46:31.022 http-nio-12012-exec-2 BizAlarmTest - OrderServiceResponse(orderId=0000000001,resultCode=1300001000,resultMessage=参数配置为空)
func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		NameKey:      "name",
		MessageKey:   "msg",
		LevelKey:     "level",
		EncodeLevel:  func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {},
		TimeKey:      "s",
		EncodeTime:   func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeName: func(n string, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(n)
		},
	}
}

func For(ctx context.Context) Logger {
	return f.For(ctx)
}

func init() {
	// ensure we always has a logger
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	f = &Factory{
		logger: zapLogger,
		ctx:    context.TODO(),
	}
}

// getLogRotateDays 获取日志最大保存时长
// 当前日志初始化在获取配置文件之前，无法通过配置文件获取
// 通过环境变量获取
func getLogRotateDays() int {
	envDays := os.Getenv(LogRotateDaysEnvKey)
	if envDays == "" {
		return DefaultLogRotateDays
	}

	days, err := strconv.Atoi(envDays)
	if err != nil {
		return DefaultLogRotateDays
	}
	return max(days, MinLogRotateDays)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

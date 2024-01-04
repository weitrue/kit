package xzap

import (
	"errors"
	"os"
	"path"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DebugFilename  = "debug.log"
	AccessFilename = "access.log"
	ErrorFilename  = "error.log"
	WarnFilename   = "warn.log"

	ConsoleMode = "console"
	VolumeMode  = "volume"

	MaxSize   = 30
	MaxBackup = 5

	DefaultTimeLayout = "2006-01-02T15:04:05.000Z07"
	DefaultKeepDays   = 7
)

var (
	// ErrLogPathNotSet is an error that indicates the log path is not set.
	ErrLogPathNotSet = errors.New("log path must be set")
	// ErrLogServiceNameNotSet is an error that indicates that the service name is not set.
	ErrLogServiceNameNotSet = errors.New("log service name must be set")
)

// SetUp 初始化zap Logger
func SetUp(c Config, opts ...Option) (*zap.Logger, error) {
	if c.KeepDays == 0 {
		c.KeepDays = DefaultKeepDays
	}

	opt := &Opt{
		fields: make(map[string]string),
	}

	for _, f := range opts {
		if f != nil {
			f(opt)
		}
	}

	if len(c.Path) == 0 {
		return nil, ErrLogPathNotSet
	}

	if len(c.ServiceName) == 0 {
		return nil, ErrLogServiceNameNotSet
	}

	switch c.Mode {
	case ConsoleMode:
		withLogLevel(c.Level)
		return setupWithConsole(opt), nil
	case VolumeMode:
		return setupWithFiles(c, opt), nil
	default:
		return setupWithFiles(c, opt), nil
	}
}

func setupWithConsole(opt *Opt) *zap.Logger {
	consoleDebugging := zapcore.Lock(os.Stdout)
	core := zapcore.NewTee(
		zapcore.NewCore(ZapConsoleEncoder(), consoleDebugging, zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= opt.level
		})),
	)

	log := zap.New(core, zap.AddCaller())
	for key, value := range opt.fields {
		log = log.WithOptions(zap.Fields(zapcore.Field{Key: key, Type: zapcore.StringType, String: value}))
	}

	return log
}

func setupWithFiles(c Config, opt *Opt) *zap.Logger {
	accessPath := path.Join(c.Path, AccessFilename)
	errorPath := path.Join(c.Path, ErrorFilename)
	severePath := path.Join(c.Path, WarnFilename)
	debugPath := path.Join(c.Path, DebugFilename)

	errPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl > zapcore.WarnLevel
	})
	warnPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.WarnLevel
	})
	infoPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel
	})
	debugPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.DebugLevel
	})

	core := zapcore.NewTee(
		zapcore.NewCore(ZapFileEncoder(), ZapLogWriter(accessPath, MaxSize, MaxBackup, c.KeepDays, c.Compress), infoPriority),
		zapcore.NewCore(ZapFileEncoder(), ZapLogWriter(errorPath, MaxSize, MaxBackup, c.KeepDays, c.Compress), errPriority),
		zapcore.NewCore(ZapFileEncoder(), ZapLogWriter(severePath, MaxSize, MaxBackup, c.KeepDays, c.Compress), warnPriority),
		zapcore.NewCore(ZapFileEncoder(), ZapLogWriter(debugPath, MaxSize, MaxBackup, c.KeepDays, c.Compress), debugPriority),
	)

	stderr := zapcore.Lock(os.Stderr) // lock for concurrent safe

	log := zap.New(core, zap.AddCaller(), zap.ErrorOutput(stderr))
	for key, value := range opt.fields {
		log = log.WithOptions(zap.Fields(zapcore.Field{Key: key, Type: zapcore.StringType, String: value}))
	}
	return log
}

func ZapLogWriter(fileName string, maxSize, maxBackups, maxAge int, isCompress bool) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   isCompress,
	}

	return zapcore.AddSync(lumberJackLogger)
}

func ZapFileEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "@timestamp",
		LevelKey:       "level",
		NameKey:        "Logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // Level 序列化为小写字符串
		EncodeTime:     TimeEncoder,                    // 记录时间设置为2006-01-02T15:04:05Z07:00
		EncodeDuration: zapcore.SecondsDurationEncoder, //  耗时设置为浮点秒数
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
	})
}

func ZapConsoleEncoder() zapcore.Encoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// TimeEncoder 设置时间格式化方式
func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(DefaultTimeLayout))
}

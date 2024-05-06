package proof

import (
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	TimeDivision = "time"
	SizeDivision = "size"

	ConsoleEncoder = "console"
	JSONEncoder    = "json"

	defaultEncoding = ConsoleEncoder
	defaultDivision = TimeDivision
	defaultUnit     = Day
)

type Options struct {
	Encoding      string
	InfoFilename  string
	ErrorFilename string
	MaxSize       int
	MaxBackups    int
	MaxAge        int
	Compress      bool
	Division      string
	LevelSeparate bool
	TimeUnit      TimeUnit

	closeDisplay int
	caller       bool
	capitalColor bool
}

func New() *Options {
	return &Options{
		Division:      defaultDivision,
		LevelSeparate: false,
		TimeUnit:      defaultUnit,
		Encoding:      defaultEncoding,
		caller:        false,
	}
}

func (o *Options) SetDivision(division string) {
	o.Division = division
}

func (o *Options) CloseConsoleDisplay() {
	o.closeDisplay = 1
}

func (o *Options) SetCaller(b bool) {
	o.caller = b
}

func (o *Options) SetTimeUnit(t TimeUnit) {
	o.TimeUnit = t
}

func (o *Options) SetMaxAge(a int) {
	o.MaxAge = a
}

func (o *Options) SetMaxSize(size int) {
	o.SetMaxSize(size)
}

func (o *Options) SetMaxBackups(n int) {
	o.SetMaxBackups(n)
}

func (o *Options) SetCompress(b bool) {
	o.Compress = b
}

func (o *Options) SetErrorFile(path string) {
	o.LevelSeparate = true
	o.ErrorFilename = path
}

func (o *Options) SetInfoFile(path string) {
	o.InfoFilename = path
}

func (o *Options) SetEncoding(encoding string) {
	o.Encoding = encoding
}

func (o *Options) SetCapitalColor(b bool) {
	o.capitalColor = b
}

// isOutput whether set output file
func (o *Options) isOutput() bool {
	return o.InfoFilename != ""
}

func (o *Options) Logger() *zap.Logger {
	var (
		core               zapcore.Core
		infoHook, warnHook io.Writer
		wsInfo             []zapcore.WriteSyncer
		wsWarn             []zapcore.WriteSyncer
	)

	if o.Encoding == "" {
		o.Encoding = defaultEncoding
	}

	encoder := encoderNameToConstructor[o.Encoding]

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ztime",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "file",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		FunctionKey:    zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	if o.capitalColor {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if o.closeDisplay == 0 {
		wsInfo = append(wsInfo, zapcore.AddSync(os.Stdout))
		wsWarn = append(wsWarn, zapcore.AddSync(os.Stdout))
	}

	if o.isOutput() {
		switch o.Division {
		case TimeDivision:
			infoHook = o.timeDivisionWriter(o.InfoFilename)
			if o.LevelSeparate {
				warnHook = o.timeDivisionWriter(o.ErrorFilename)
			}
		case SizeDivision:
			infoHook = o.sizeDivisionWriter(o.InfoFilename)
			if o.LevelSeparate {
				warnHook = o.sizeDivisionWriter(o.ErrorFilename)
			}
		}
		wsInfo = append(wsInfo, zapcore.AddSync(infoHook))
	}

	if o.ErrorFilename != "" {
		wsWarn = append(wsWarn, zapcore.AddSync(warnHook))
	}

	if o.LevelSeparate {
		core = zapcore.NewTee(
			zapcore.NewCore(encoder(encoderConfig), zapcore.NewMultiWriteSyncer(wsInfo...), infoLevel()),
			zapcore.NewCore(encoder(encoderConfig), zapcore.NewMultiWriteSyncer(wsWarn...), warnLevel()),
		)
	} else {
		core = zapcore.NewCore(encoder(encoderConfig), zapcore.NewMultiWriteSyncer(wsInfo...), zap.InfoLevel)
	}

	development := zap.Development()
	stackTrace := zap.AddStacktrace(zapcore.WarnLevel)

	// field := zap.Fields(
	//	zap.String("serviceName", os.Getenv("project")),
	//	zap.String("hostName", os.Getenv("HOSTNAME")),
	// )

	var logger *zap.Logger
	if o.caller {
		logger = zap.New(core, zap.AddCaller(), development, stackTrace)
	} else {
		logger = zap.New(core, development, stackTrace)
	}

	return logger
}

func (o *Options) sizeDivisionWriter(filename string) io.Writer {
	hook := &lumberjack.Logger{
		Filename:   filename,     // 日志文件路径
		MaxSize:    o.MaxSize,    // 每个文件保存的最大尺寸 单位：MB
		MaxBackups: o.MaxBackups, // 日志文件最多保存多少个备份
		MaxAge:     o.MaxAge,     // 文件最多保存多少天
		Compress:   o.Compress,   // 是否压缩
	}
	return hook
}

func (o *Options) timeDivisionWriter(filename string) io.Writer {
	hook, err := rotatelogs.New(
		filename+o.TimeUnit.Format(),
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Duration(int64(24*time.Hour)*int64(o.MaxAge))),
		rotatelogs.WithRotationTime(o.TimeUnit.RotationGap()),
	)

	if err != nil {
		panic(err)
	}
	return hook
}

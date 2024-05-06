package proof

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	encoderNameToConstructor = map[string]func(zapcore.EncoderConfig) zapcore.Encoder{
		ConsoleEncoder: func(encoderConfig zapcore.EncoderConfig) zapcore.Encoder {
			return zapcore.NewConsoleEncoder(encoderConfig)
		},
		JSONEncoder: func(encoderConfig zapcore.EncoderConfig) zapcore.Encoder {
			return zapcore.NewJSONEncoder(encoderConfig)
		},
	}
)

func infoLevel() zap.LevelEnablerFunc {
	return func(level zapcore.Level) bool {
		return level < zapcore.WarnLevel
	}
}

func warnLevel() zap.LevelEnablerFunc {
	return func(level zapcore.Level) bool {
		return level >= zapcore.WarnLevel
	}
}

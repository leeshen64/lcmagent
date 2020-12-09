package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// default log file folders
const (
	defaultLogFolder = (".")
)

// Logger ...
type Logger struct {
	*zap.Logger
}

var (
	logFolder      string
	logRotateCount int
	once           sync.Once
	once1          sync.Once
	instance       *Logger
)

func customConsoleEncoder() zapcore.Encoder {
	consoleEncoderConfig := zap.NewProductionConfig().EncoderConfig
	consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // ISO8601 UTC
	consoleEncoderConfig.EncodeLevel = customLevelEncoder
	consoleEncoderConfig.TimeKey = "time"
	consoleEncoderConfig.LevelKey = "level"
	consoleEncoderConfig.NameKey = "logger"
	consoleEncoderConfig.CallerKey = "caller"

	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	return consoleEncoder
}

func customJSONEncoder() zapcore.Encoder {
	jsonEncoderConfig := zap.NewProductionConfig().EncoderConfig
	jsonEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // ISO8601 UTC
	jsonEncoderConfig.TimeKey = "time"
	jsonEncoderConfig.LevelKey = "level"
	jsonEncoderConfig.NameKey = "logger"
	jsonEncoderConfig.CallerKey = "caller"

	jsonEncoder := zapcore.NewJSONEncoder(jsonEncoderConfig)
	return jsonEncoder
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%s]", level.String()))
}

// SetEnvParam ...
func SetEnvParam(folderString string, rotateCount int) {
	once1.Do(func() {
		logFolder = folderString
		logRotateCount = rotateCount
	})
}

// New ...
// logger.New().Info("This is an Info message", zap.String("customField", "123"))
func New() *Logger {
	once.Do(func() {
		// all of the levels
		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.DebugLevel
		})

		appName := getAppName()
		if _, err := os.Stat(logFolder); os.IsNotExist(err) {
			// does not exist
			os.MkdirAll(logFolder, os.ModePerm)
		}

		fileRotateHook := zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFolder + string(os.PathSeparator) + appName + ".log",
			MaxSize:    1, // MB
			MaxBackups: logRotateCount,
			MaxAge:     30, // day
			Compress:   false,
		})
		consoleHook := zapcore.Lock(os.Stdout)

		consoleEncoder := customConsoleEncoder()

		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, fileRotateHook, lowPriority),
			zapcore.NewCore(consoleEncoder, consoleHook, lowPriority))

		instance = &Logger{zap.New(core, zap.AddCaller())}
	})

	return instance
}

func getCurrentDirectory() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return strings.Replace(dir, "\\", "/", -1), err
}

func getAppName() string {
	full := os.Args[0]
	full = strings.Replace(full, "\\", "/", -1)
	splits := strings.Split(full, "/")
	if len(splits) >= 1 {
		name := splits[len(splits)-1]
		name = strings.TrimSuffix(name, ".exe")
		return name
	}

	return ""
}

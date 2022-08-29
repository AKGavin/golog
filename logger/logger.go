package logger

import (
	"io"
	"os"
	"path"
	"sync/atomic"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapFIled = zapcore.Field

var ccbLog *zap.Logger
var appName string = "api"
var maxAge int = 7 * 24
var logInited int32

var logLevel = zap.NewAtomicLevel()

const (
	DEBUG = "debug"
	INFO  = "info"
	WARN  = "warn"
	ERROR = "error"
	PANIC = "panic"
)

/*
func Init(logConf map[string]interface{}) {

	mkdirIfNotExist(logConf["logPath"].(string))

	if logConf["logName"] != nil {
		appName = logConf["logName"].(string)
	}

	level := logConf["level"].(string)
	SetLevel(level)

	filePath := path.Join(logConf["logPath"].(string), appName)

	infoWriter := getWriter(filePath)
	w := zapcore.AddSync(infoWriter)

	config := zap.NewProductionEncoderConfig()
	config.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		w,
		logLevel,
	)

	apmCore := &apmzap.Core{}
	ccbLog = zap.New(apmCore.WrapCore(core), zap.AddCaller(), zap.AddCallerSkip(1))
	atomic.StoreInt32(&logInited, 1)

	ccbLog.Info("log init......")
}*/

func Init2(logConf map[string]interface{}) {
	level := "info"
	if logConf["level"] != nil {
		level = logConf["level"].(string)
	}
	SetLevel(level)

	if logConf["logName"] != nil {
		appName = logConf["logName"].(string)
	}
	filePath := path.Join(logConf["logPath"].(string), appName)
	mkdirIfNotExist(logConf["logPath"].(string))

	if logConf["maxAge"] != nil {
		maxAge = logConf["maxAge"].(int)
	}
	// debug level
	debugPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= logLevel.Level()
	})
	// error level
	errorPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})

	cores := [...]zapcore.Core{
		getEncoderCore(filePath, maxAge, debugPriority),
		getEncoderCore(filePath+"_error", maxAge, errorPriority),
	}
	ccbLog = zap.New(zapcore.NewTee(cores[:]...), zap.AddCaller(), zap.AddCallerSkip(1))

	atomic.StoreInt32(&logInited, 1)
}

// 获取Encoder的zapcore.Core
func getEncoderCore(fileName string, maxAge int, level zapcore.LevelEnabler) (core zapcore.Core) {
	config := zap.NewProductionEncoderConfig()
	config.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	writer := getWriter(fileName, maxAge) // 使用file-rotatelogs进行日志分割

	return zapcore.NewCore(zapcore.NewJSONEncoder(config), zapcore.AddSync(writer), level)
}

func mkdirIfNotExist(path string) {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		os.MkdirAll(path, 0774)
	}
}

func getWriter(filename string, maxAge int) io.Writer {
	// 保存7天内的日志，每24小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		filename+".%Y-%m-%d.log",
		rotatelogs.WithMaxAge(time.Duration(24*maxAge)*time.Hour),
		rotatelogs.WithLinkName(filename+".log"),
		rotatelogs.WithRotationTime(time.Hour*24),
	)

	if err != nil {
		panic(err)
	}
	return hook
}

func SetLevel(level string) {
	l := zapcore.DebugLevel
	switch level {
	case DEBUG:
		l = zap.DebugLevel
	case INFO:
		l = zap.InfoLevel
	case WARN:
		l = zap.WarnLevel
	case ERROR:
		l = zap.ErrorLevel
	default:
		l = zap.InfoLevel
	}

	logLevel.SetLevel(zapcore.Level(l))
}

/*
 * Copy zap.Logger and can set unique fields
 */
func LogAppend(fields ...ZapFIled) *zap.Logger {
	return ccbLog.With(fields...)
}

func Debug(msg string, fields ...ZapFIled) {
	ccbLog.Debug(msg, fields...)
}

func Info(msg string, fields ...ZapFIled) {
	ccbLog.Info(msg, fields...)
}

func Warn(msg string, fields ...ZapFIled) {
	ccbLog.Warn(msg, fields...)
}
func Error(msg string, fields ...ZapFIled) {
	ccbLog.Error(msg, fields...)
}

func Panic(msg string, fields ...ZapFIled) {
	ccbLog.DPanic(msg, fields...)
}

// func GetClientIP(r *http.Request, headers ...string) string {
// 	for _, header := range headers {
// 		ip := r.Header.Get(header)
// 		if ip != "" {
// 			return strings.Split(ip, ",")[0]
// 		}
// 	}
// 	return strings.Split(r.RemoteAddr, ":")[0]
// }

/*
 * 注意 程序退出时需要调用该方法
 */
func Sync() {
	if atomic.LoadInt32(&logInited) > 0 {
		ccbLog.Sync()
	}
}

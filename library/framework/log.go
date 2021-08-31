package framework

import (
	common2 "github.com/pingcap-inc/tiem/library/common"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogRecord struct {
	defaultLogEntry *log.Entry

	forkFileEntry map[string]*log.Entry

	LogLevel      string
	LogOutput     string
	LogFileRoot   string
	LogFileName   string
	LogMaxSize    int
	LogMaxAge     int
	LogMaxBackups int
	LogLocalTime  bool
	LogCompress   bool
}

const (
	// LogDebug debug level
	LogDebug = "debug"
	// LogInfo info level
	LogInfo = "info"
	// LogWarn warn level
	LogWarn = "warn"
	// LogError error level
	LogError = "error"
	// LogFatal fatal level
	LogFatal = "fatal"
)

const (
	// OutputConsole log console output
	OutputConsole = "console"
	// OutputFile log file output
	OutputFile = "file"
)

const (
	// RecordSysField record sys name
	RecordSysField = "sys"
	// RecordModField record mod name
	RecordModField = "mod"
	// RecordFunField record fun name
	RecordFunField = "fun"
	// RecordFileField record file name
	RecordFileField = "file"
	// RecordLineField record line number
	RecordLineField = "line"
)

func DefaultLogRecord() *LogRecord {
	lr := &LogRecord{
		LogLevel:      "info",
		LogOutput:     "file",
		LogFileRoot:   "." + common2.LogDirPrefix,
		LogFileName:   "default",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
	}

	// Record sys and mod default init
	lr.defaultLogEntry = lr.forkEntry(lr.LogFileName)
	lr.forkFileEntry = map[string]*log.Entry{lr.LogFileName: lr.defaultLogEntry}
	//.WithField(RecordSysField, lr.RecordSysName).WithField(RecordModField, lr.RecordModName)
	return lr
}

func NewLogRecordFromArgs(serviceName ServiceNameEnum, args *ClientArgs) *LogRecord {
	lr := &LogRecord{
		LogLevel:      args.LogLevel,
		LogOutput:     "file",
		LogFileRoot:   args.DataDir + common2.LogDirPrefix,
		LogFileName:   serviceName.ServerName(),
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
	}

	// Record sys and mod default init
	lr.defaultLogEntry = lr.forkEntry(lr.LogFileName)
	lr.forkFileEntry = map[string]*log.Entry{lr.LogFileName: lr.defaultLogEntry}

	//.WithField(RecordSysField, lr.RecordSysName).WithField(RecordModField, lr.RecordModName)
	return lr
}

func (lr *LogRecord) forkEntry(fileName string) *log.Entry {
	logger := log.New()

	// Set log format
	logger.SetFormatter(&log.JSONFormatter{})
	// Set log level
	logger.SetLevel(getLogLevel(lr.LogLevel))

	// Define output type writer
	writers := []io.Writer{os.Stdout}

	// Determine whether the log output contains the file type
	if strings.Contains(strings.ToLower(lr.LogOutput), OutputFile) {
		writers = append(writers, &lumberjack.Logger{
			Filename: lr.LogFileRoot + fileName + ".log",
			MaxSize: lr.LogMaxSize,
			MaxAge: lr.LogMaxAge,
			MaxBackups: lr.LogMaxBackups,
			LocalTime: lr.LogLocalTime,
			Compress: lr.LogCompress,
		})
	}
	// remove the os.Stdout output
	writers = writers[1:]

	// Set log output
	logger.SetOutput(io.MultiWriter(writers...))
	return log.NewEntry(logger)
}

func (lr *LogRecord) ForkFile(fileName string) *log.Entry {
	if entry, ok := lr.forkFileEntry[fileName]; ok {
		return entry
	} else {

		lr.forkFileEntry[fileName] = lr.forkEntry(fileName)
		return lr.forkFileEntry[fileName]
	}
}

// Tool method to get log level
func getLogLevel(level string) log.Level {
	switch strings.ToLower(level) {
	case LogDebug:
		return log.DebugLevel
	case LogInfo:
		return log.InfoLevel
	case LogWarn:
		return log.WarnLevel
	case LogError:
		return log.ErrorLevel
	case LogFatal:
		return log.FatalLevel
	}
	return log.DebugLevel
}

type logCtxKeyType struct{}

var logCtxKey logCtxKeyType

func (lr *LogRecord) Record(key string, value interface{}) *log.Entry {
	return lr.defaultLogEntry.WithField(key, value)
}

func (lr *LogRecord) Records(fields log.Fields) *log.Entry {
	return lr.defaultLogEntry.WithFields(fields)
}

func (lr *LogRecord) RecordSys(sys string) *LogRecord {
	lr.defaultLogEntry = lr.defaultLogEntry.WithField(RecordSysField, sys)
	return lr
}

func (lr *LogRecord) RecordMod(mod string) *LogRecord {
	lr.defaultLogEntry = lr.defaultLogEntry.WithField(RecordModField, mod)
	return lr
}

func (lr *LogRecord) RecordFun() *log.Entry {
	logEntry := lr.defaultLogEntry
	if pc, file, line, ok := runtime.Caller(2); ok {
		ptr := runtime.FuncForPC(pc)
		//fmt.Println(ptr.Name(), file, line)
		logEntry = lr.defaultLogEntry.WithField(RecordFunField, ptr.Name()).
			WithField(RecordFileField, path.Base(file)).WithField(RecordLineField, line)
	}
	return logEntry
}

func (lr *LogRecord) Debug(args ...interface{}) {
	lr.RecordFun().Debug(args...)
}

func (lr *LogRecord) Debugf(format string, args ...interface{}) {
	lr.RecordFun().Debugf(format, args...)
}

func (lr *LogRecord) Debugln(args ...interface{}) {
	lr.RecordFun().Debugln(args...)
}

func (lr *LogRecord) Info(args ...interface{}) {
	lr.RecordFun().Info(args...)
}

func (lr *LogRecord) Infof(format string, args ...interface{}) {
	lr.RecordFun().Infof(format, args...)
}

func (lr *LogRecord) Infoln(args ...interface{}) {
	lr.RecordFun().Infoln(args...)
}

func (lr *LogRecord) Warn(args ...interface{}) {
	lr.RecordFun().Warn(args...)
}

func (lr *LogRecord) Warnf(format string, args ...interface{}) {
	lr.RecordFun().Warnf(format, args...)
}

func (lr *LogRecord) Warnln(args ...interface{}) {
	lr.RecordFun().Warnln(args...)
}

func (lr *LogRecord) Warning(args ...interface{}) {
	lr.RecordFun().Warning(args...)
}

func (lr *LogRecord) Warningf(format string, args ...interface{}) {
	lr.RecordFun().Warningf(format, args...)
}

func (lr *LogRecord) Warningln(args ...interface{}) {
	lr.RecordFun().Warningln(args...)
}

func (lr *LogRecord) Error(args ...interface{}) {
	lr.RecordFun().Error(args...)
}

func (lr *LogRecord) Errorf(format string, args ...interface{}) {
	lr.RecordFun().Errorf(format, args...)
}

func (lr *LogRecord) Errorln(args ...interface{}) {
	lr.RecordFun().Errorln(args...)
}

func (lr *LogRecord) Fatal(args ...interface{}) {
	lr.RecordFun().Fatal(args...)
}

func (lr *LogRecord) Fatalf(format string, args ...interface{}) {
	lr.RecordFun().Fatalf(format, args...)
}

func (lr *LogRecord) Fatalln(args ...interface{}) {
	lr.RecordFun().Fatalln(args...)
}

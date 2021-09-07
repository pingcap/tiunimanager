package framework

import (
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	common2 "github.com/pingcap-inc/tiem/library/common"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type RootLogger struct {
	defaultLogEntry *logrus.Entry

	forkFileEntry map[string]*logrus.Entry

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
	RecordSysField = "source_sys"
	// RecordModField record mod name
	RecordModField = "source_mod"
	// RecordFileField record file name
	RecordFileField = "source_file"
	// RecordFunField record fun name
	RecordFunField = "source_fun"
	// RecordLineField record line number
	RecordLineField = "source_line"
)

func DefaultRootLogger() *RootLogger {
	lr := &RootLogger{
		LogLevel:      "info",
		LogOutput:     "file",
		LogFileRoot:   "." + common2.LogDirPrefix,
		LogFileName:   "default-server",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
	}

	// WithField sys and mod default init
	lr.defaultLogEntry = lr.forkEntry(lr.LogFileName)
	lr.forkFileEntry = map[string]*logrus.Entry{lr.LogFileName: lr.defaultLogEntry}
	//.WithField(RecordSysField, lr.RecordSysName).WithField(RecordModField, lr.RecordModName)
	return lr
}

func NewLogRecordFromArgs(serviceName ServiceNameEnum, args *ClientArgs) *RootLogger {
	lr := &RootLogger{
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

	// WithField sys and mod default init
	lr.defaultLogEntry = lr.forkEntry(lr.LogFileName)
	lr.forkFileEntry = map[string]*logrus.Entry{lr.LogFileName: lr.defaultLogEntry}

	//.WithField(RecordSysField, lr.RecordSysName).WithField(RecordModField, lr.RecordModName)
	return lr
}

func (lr *RootLogger) ForkFile(fileName string) *logrus.Entry {
	if entry, ok := lr.forkFileEntry[fileName]; ok {
		return entry
	} else {

		lr.forkFileEntry[fileName] = lr.forkEntry(fileName)
		return lr.forkFileEntry[fileName]
	}
}

func (lr *RootLogger) Entry() *logrus.Entry {
	return lr.defaultLogEntry
}

func Caller() logrus.Fields {
	if pc, file, line, ok := runtime.Caller(1); ok {
		ptr := runtime.FuncForPC(pc)
		return map[string]interface{}{
			RecordFunField: ptr.Name(),
			RecordFileField: path.Base(file),
			RecordLineField: line,
		}
	}
	return map[string]interface{}{}
}

func (lr *RootLogger) forkEntry(fileName string) *logrus.Entry {
	logger := logrus.New()

	// Set log format
	logger.SetFormatter(&logrus.JSONFormatter{})
	// Set log level
	logger.SetLevel(getLogLevel(lr.LogLevel))

	// Define output type writer
	writers := []io.Writer{os.Stdout}

	// Determine whether the log output contains the file type
	if strings.Contains(strings.ToLower(lr.LogOutput), OutputFile) {
		writers = append(writers, &lumberjack.Logger{
			Filename:   lr.LogFileRoot + fileName + ".log",
			MaxSize:    lr.LogMaxSize,
			MaxAge:     lr.LogMaxAge,
			MaxBackups: lr.LogMaxBackups,
			LocalTime:  lr.LogLocalTime,
			Compress:   lr.LogCompress,
		})
	}
	// remove the os.Stdout output
	writers = writers[1:]

	// Set log output
	logger.SetOutput(io.MultiWriter(writers...))
	return logrus.NewEntry(logger)
}

// Tool method to get log level
func getLogLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case LogDebug:
		return logrus.DebugLevel
	case LogInfo:
		return logrus.InfoLevel
	case LogWarn:
		return logrus.WarnLevel
	case LogError:
		return logrus.ErrorLevel
	case LogFatal:
		return logrus.FatalLevel
	}
	return logrus.DebugLevel
}

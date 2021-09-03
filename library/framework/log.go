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

type RootLogger struct {
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

func DefaultLogRecord() *RootLogger {
	lr := &RootLogger{
		LogLevel:      "info",
		LogOutput:     "file",
		LogFileRoot:   "." + common2.LogDirPrefix,
		LogFileName:   "service",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
	}

	// WithField sys and mod default init
	lr.defaultLogEntry = lr.forkEntry(lr.LogFileName)
	lr.forkFileEntry = map[string]*log.Entry{lr.LogFileName: lr.defaultLogEntry}
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
	lr.forkFileEntry = map[string]*log.Entry{lr.LogFileName: lr.defaultLogEntry}

	//.WithField(RecordSysField, lr.RecordSysName).WithField(RecordModField, lr.RecordModName)
	return lr
}

func (lr *RootLogger) forkEntry(fileName string) *log.Entry {
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

func (lr *RootLogger) ForkFile(fileName string) *log.Entry {
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

func (lr *RootLogger) defaultRecord() *log.Entry {
	logEntry := lr.defaultLogEntry
	if pc, file, line, ok := runtime.Caller(2); ok {
		ptr := runtime.FuncForPC(pc)
		//fmt.Println(ptr.Name(), file, line)
		logEntry = lr.defaultLogEntry.WithField(RecordFunField, ptr.Name()).
			WithField(RecordFileField, path.Base(file)).WithField(RecordLineField, line)
	}
	return logEntry
}

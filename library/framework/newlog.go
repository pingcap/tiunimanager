package framework

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)
var ErrRecordNotFound = errors.New("record not found")

type LogLevel int

const (
	Silent LogLevel = iota + 1
	Error
	Warn
	Info
)
// Writer log writer interface
type Writer interface {
	Printf(string, ...interface{})
}

type Config struct {
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	LogLevel                  LogLevel
}

type Logger struct {
	Writer
	Config
	traceStr, traceErrStr, traceWarnStr string
}

func New(writer Writer, config Config) *Logger {
	l := &Logger{
		Writer: writer,
		Config: config,
	}
	return l
}

func (p *Logger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *p
	newlogger.LogLevel = LogLevel(level)
	return &newlogger
}

func (p *Logger) Info(ctx context.Context, s string, i ...interface{}) {
	LogWithContext(ctx).Infof(s, i...)
}

func (p *Logger) Warn(ctx context.Context, s string, i ...interface{}) {
	LogWithContext(ctx).Warnf(s, i...)
}

func (p *Logger) Error(ctx context.Context, s string, i ...interface{}) {
	LogWithContext(ctx).Errorf(s, i...)
}

func (p *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if p.LogLevel <= Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && p.LogLevel >= Error && (!errors.Is(err, ErrRecordNotFound) || !p.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			p.Printf(p.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			p.Printf(p.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > p.SlowThreshold && p.SlowThreshold != 0 && p.LogLevel >= Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", p.SlowThreshold)
		if rows == -1 {
			p.Printf(p.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			p.Printf(p.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case p.LogLevel == Info:
		sql, rows := fc()
		if rows == -1 {
			p.Printf(p.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			p.Printf(p.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func (p *Logger) Printf(format string, args ...interface{}) {
	Log().Infof(format, args...)
}
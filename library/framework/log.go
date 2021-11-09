
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package framework

import (
	"context"
	"fmt"
	"github.com/asim/go-micro/v3/server"
	"github.com/pingcap-inc/tiem/library/common"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

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
		LogFileRoot:   "." + common.LogDirPrefix,
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
		LogFileRoot:   args.DataDir + common.LogDirPrefix,
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

func (p *RootLogger) ForkFile(fileName string) *logrus.Entry {
	if entry, ok := p.forkFileEntry[fileName]; ok {
		return entry
	} else {

		p.forkFileEntry[fileName] = p.forkEntry(fileName)
		return p.forkFileEntry[fileName]
	}
}

func (p *RootLogger) Entry() *logrus.Entry {
	return p.defaultLogEntry
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

func (p *RootLogger) forkEntry(fileName string) *logrus.Entry {
	logger := logrus.New()

	// Set log format
	logger.SetFormatter(&logrus.JSONFormatter{})
	// Set log level
	logger.SetLevel(getLogLevel(p.LogLevel))

	// Define output type writer
	writers := []io.Writer{os.Stdout}

	// Determine whether the log output contains the file type
	if strings.Contains(strings.ToLower(p.LogOutput), OutputFile) {
		writers = append(writers, &lumberjack.Logger{
			Filename:   p.LogFileRoot + fileName + ".log",
			MaxSize:    p.LogMaxSize,
			MaxAge:     p.LogMaxAge,
			MaxBackups: p.LogMaxBackups,
			LocalTime:  p.LogLocalTime,
			Compress:   p.LogCompress,
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

// logWrapper returns a Handler Wrapper for log
func logWrapper(bf *BaseFramework) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := fmt.Sprintf("%s.%s", req.Service(), req.Method())
			microLog := bf.LogWithContext(ctx).
				WithField("micro-method", name)

			microLog.Infof("micro-method start, [request]%s", req.Body())

			if err := h(ctx, req, rsp); err != nil {
				microLog.Errorf("micro-method failed, [request]%s, [error]%s", req.Body(), err.Error())
			} else {
				microLog.Infof("micro-method end, [response]%s", rsp)
			}
			return nil
		}
	}
}

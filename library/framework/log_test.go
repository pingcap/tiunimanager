package framework

import (
	log "github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestDefaultLogRecord(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		got := DefaultLogRecord()
		Assert(got != nil)
		Assert(got.LogLevel == "info")
		Assert(got.LogLevel == "info")
	})
}

func TestRootLogger_ForkFile(t *testing.T) {
	InitBaseFrameworkForUt(MetaDBService)
	GetRootLogger().ForkFile("aaa").Info("some")
	GetRootLogger().ForkFile("aaa").Info("another")
}

func TestRootLogger_RecordFun(t *testing.T) {
	type fields struct {
		defaultLogEntry *log.Entry
		forkFileEntry   map[string]*log.Entry
		LogLevel        string
		LogOutput       string
		LogFileRoot     string
		LogFileName     string
		LogMaxSize      int
		LogMaxAge       int
		LogMaxBackups   int
		LogLocalTime    bool
		LogCompress     bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *log.Entry
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lr := &RootLogger{
				defaultLogEntry: tt.fields.defaultLogEntry,
				forkFileEntry:   tt.fields.forkFileEntry,
				LogLevel:        tt.fields.LogLevel,
				LogOutput:       tt.fields.LogOutput,
				LogFileRoot:     tt.fields.LogFileRoot,
				LogFileName:     tt.fields.LogFileName,
				LogMaxSize:      tt.fields.LogMaxSize,
				LogMaxAge:       tt.fields.LogMaxAge,
				LogMaxBackups:   tt.fields.LogMaxBackups,
				LogLocalTime:    tt.fields.LogLocalTime,
				LogCompress:     tt.fields.LogCompress,
			}
			if got := lr.DefaultRecord(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRootLogger_forkEntry(t *testing.T) {
	//lr.forkEntry()
}

func Test_getLogLevel(t *testing.T) {
	Assert(getLogLevel("info") == log.InfoLevel)
}

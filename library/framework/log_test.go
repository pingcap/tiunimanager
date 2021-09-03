package framework

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultLogRecord(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		got := DefaultRootLogger()
		Assert(got != nil)
		Assert(got.LogLevel == "info")
		Assert(got.LogLevel == "info")
	})
}

func TestRootLogger_ForkFile(t *testing.T) {
	InitBaseFrameworkForUt(MetaDBService)
	LogForkFile("aaa").Info("some")
	LogForkFile("aaa").Info("another")
}

func TestRootLogger_forkEntry(t *testing.T) {
	//lr.forkEntry()
}

func Test_getLogLevel(t *testing.T) {
	Assert(getLogLevel("info") == log.InfoLevel)
	Assert(getLogLevel("debug") == log.DebugLevel)
	Assert(getLogLevel("warn") == log.WarnLevel)
	Assert(getLogLevel("error") == log.ErrorLevel)
	Assert(getLogLevel("fatal") == log.FatalLevel)
	Assert(getLogLevel("aaaa") == log.DebugLevel)
}

func TestCaller(t *testing.T) {
	got := Caller()
	assert.Equal(t, 3, len(got))

}
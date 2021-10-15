
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package framework

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultLogRecord(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		got := DefaultRootLogger()
		assert.NotNil(t, got)
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
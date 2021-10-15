
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
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInitBaseFrameworkForUt(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		f := InitBaseFrameworkForUt(MetaDBService,
			func(d *BaseFramework) error {
				d.GetServiceMeta().ServicePort = 99999
				return nil
			},
		)
		if f.GetServiceMeta().ServicePort != 99999 {
			t.Errorf("InitBaseFrameworkFromArgs() service port wrong, want = %v, got %v", 99999, f.GetServiceMeta().ServicePort)
		}
	})
}

func TestGetRootLogger(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		Current = nil
		got := GetRootLogger()
		assert.Equal(t, "default-server", got.LogFileName)
	})
	t.Run("service", func(t *testing.T) {
		Current = nil
		InitBaseFrameworkForUt(MetaDBService)
		got := GetRootLogger()
		assert.Equal(t, MetaDBService.ServerName(), got.LogFileName)
	})
}

func TestLog(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		Current = nil
		entry := Log().WithFields(Caller())
		_, ok := entry.Data[RecordFunField]
		assert.True(t, ok, "RecordFunField not found")
	})
	t.Run("service", func(t *testing.T) {
		Current = nil
		InitBaseFrameworkForUt(MetaDBService)
		entry := Log().WithFields(Caller())
		_, ok := entry.Data[RecordFunField]
		assert.True(t, ok, "RecordFunField not found")

	})
}

func TestBaseFramework_GetLoggerWithContext(t *testing.T) {
	ctx := &gin.Context{}
	ctx.Set(TiEM_X_TRACE_ID_NAME, "111")

	got := LogWithContext(ctx)
	assert.Equal(t, "111", got.Data[TiEM_X_TRACE_ID_NAME])
}

func TestBaseFramework_Get(t *testing.T) {
	f := InitBaseFrameworkForUt(MetaDBService)
	assert.NotNil(t, f.GetClientArgs())
	assert.NotNil(t, f.GetConfiguration())
	assert.NotNil(t, f.GetTracer())
	assert.NotNil(t, f.GetDataDir())
	assert.NotNil(t, f.GetDeployDir())

	assert.NotNil(t, f.GetServiceMeta())
	assert.NoError(t, f.Shutdown())
	assert.NoError(t, f.StopService())

	ctx := &gin.Context{}
	ctx.Set(TiEM_X_TRACE_ID_NAME, "111")
	assert.Equal(t, "111", f.LogWithContext(ctx).Data[TiEM_X_TRACE_ID_NAME])
}


func TestBaseFramework_loadCert(t *testing.T) {
	b := BaseFramework{
		certificate: &CertificateInfo{
			"./../../bin/cert/server.crt",
			"./../../bin/cert/server.key",
		},
	}
	got := b.loadCert()
	assert.NotNil(t, got)
}

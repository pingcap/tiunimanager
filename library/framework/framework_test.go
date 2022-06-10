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
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetRootLogger(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		Current = nil
		got := GetRootLogger()
		assert.Equal(t, "default-server", got.LogFileName)
	})
}

func TestLog(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		Current = nil
		entry := Log().WithFields(Caller())
		_, ok := entry.Data[RecordFunField]
		assert.True(t, ok, "RecordFunField not found")
	})
}

func TestBaseFramework_GetLoggerWithContext(t *testing.T) {
	ctx := &gin.Context{}
	ctx.Set(TiEM_X_TRACE_ID_KEY, "111")

	got := LogWithContext(ctx)
	assert.Equal(t, "111", got.Data[TiEM_X_TRACE_ID_KEY])
}

func TestBaseFramework_loadCert(t *testing.T) {
	b := BaseFramework{
		certificate: &CertificateInfo{
			"./../../bin/cert/server.crt",
			"./../../bin/cert/server.key",
		},
	}
	got := b.loadCert("grpc")
	assert.NotNil(t, got)
}

func Test_GetPrivateKeyFilePath(t *testing.T) {
	InitBaseFrameworkForUt(ClusterService)
	privPath := GetPrivateKeyFilePath("test_tiunimanager")
	assert.Equal(t, "/home/test_tiunimanager/.ssh/tiup_rsa", privPath)
}

func Test_GetPublicKeyFilePath(t *testing.T) {
	InitBaseFrameworkForUt(ClusterService)
	publicPath := GetPublicKeyFilePath("test_tiunimanager")
	assert.Equal(t, "/home/test_tiunimanager/.ssh/id_rsa.pub", publicPath)
}

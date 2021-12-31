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
 ******************************************************************************/

package models

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetReaderWriter(t *testing.T) {
	assert.NotEmpty(t, GetBRReaderWriter())
	SetBRReaderWriter(nil)
	assert.Empty(t, GetBRReaderWriter())

	assert.NotEmpty(t, GetChangeFeedReaderWriter())
	SetChangeFeedReaderWriter(nil)
	assert.Empty(t, GetChangeFeedReaderWriter())

	assert.NotEmpty(t, GetWorkFlowReaderWriter())
	SetWorkFlowReaderWriter(nil)
	assert.Empty(t, GetWorkFlowReaderWriter())

	assert.NotEmpty(t, GetImportExportReaderWriter())
	SetImportExportReaderWriter(nil)
	assert.Empty(t, GetImportExportReaderWriter())

	assert.NotEmpty(t, GetResourceReaderWriter())

	assert.NotEmpty(t, GetClusterReaderWriter())
	SetClusterReaderWriter(nil)
	assert.Empty(t, GetClusterReaderWriter())

	assert.NotEmpty(t, GetConfigReaderWriter())
	SetConfigReaderWriter(nil)
	assert.Empty(t, GetConfigReaderWriter())

	assert.NotEmpty(t, GetSecondPartyOperationReaderWriter())
	SetSecondPartyOperationReaderWriter(nil)
	assert.Empty(t, GetSecondPartyOperationReaderWriter())

	assert.NotEmpty(t, GetParameterGroupReaderWriter())
	SetParameterGroupReaderWriter(nil)
	assert.Empty(t, GetParameterGroupReaderWriter())

	assert.NotEmpty(t, GetClusterParameterReaderWriter())
	SetClusterParameterReaderWriter(nil)
	assert.Empty(t, GetClusterParameterReaderWriter())

	assert.NotEmpty(t, GetAccountReaderWriter())
	SetAccountReaderWriter(nil)
	assert.Empty(t, GetAccountReaderWriter())

	assert.NotEmpty(t, GetTenantReaderWriter())
	SetTenantReaderWriter(nil)
	assert.Empty(t, GetTenantReaderWriter())

	assert.NotEmpty(t, GetTokenReaderWriter())
	SetTokenReaderWriter(nil)
	assert.Empty(t, GetTokenReaderWriter())

	assert.NotEmpty(t, GetMirrorReaderWriter())
	SetMirrorReaderWriter(nil)
	assert.Empty(t, GetMirrorReaderWriter())

	assert.NotEmpty(t, GetProductReaderWriter())
	SetProductReaderWriter(nil)
	assert.Empty(t, GetProductReaderWriter())
}

func TestMain(m *testing.M) {
	var testFilePath string
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			testFilePath = d.GetDataDir()
			os.MkdirAll(testFilePath, 0755)

			return Open(d, false)
		},
	)
	code := m.Run()
	os.RemoveAll(testFilePath)

	os.Exit(code)
}

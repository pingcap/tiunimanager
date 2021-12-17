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

package importexport

import (
	"context"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDataImportConfig(t *testing.T) {
	meta := &handler.ClusterMeta{}
	info := &importInfo{
		ClusterId:   "test-cls",
		UserName:    "root",
		Password:    "root",
		FilePath:    "/home/em/import",
		RecordId:    "",
		StorageType: "s3",
		ConfigPath:  "/home/em/import",
	}
	config := NewDataImportConfig(context.TODO(), meta, info)
	assert.NotNil(t, config)
}

/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

package management

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClusterInstance_GetDir(t *testing.T) {
	instance := ClusterInstance {
		ClusterID: "cluster1",
		DiskPath: "/sda",
		Type: string(constants.ComponentIDTiKV),
	}
	t.Run("deploy", func(t *testing.T) {
		assert.Equal(t, "/sda/cluster1/tikv-deploy",instance.GetDeployDir())
	})
	t.Run("data", func(t *testing.T) {
		assert.Equal(t, "/sda/cluster1/tikv-data",instance.GetDataDir())
	})
	t.Run("log", func(t *testing.T) {
		assert.Equal(t, "/sda/cluster1/tikv-deploy/cluster1/tidb-log", instance.GetLogDir())
	})
}

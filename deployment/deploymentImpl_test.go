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
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: deploymentImpl_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/17
*******************************************************************************/

package deployment

import (
	"context"
	"testing"
)

var manager *Manager

func init() {
	manager = &Manager{
		TiUPBinPath: "mock_tiup",
	}
}

func TestManager_Deploy(t *testing.T) {
	_, err := manager.Deploy(context.TODO(), TiUPComponentTypeCluster, TestClusterID, TestVersion,
		TestTiDBTopo, TestTiUPHome, TestWorkFlowID, []string{"-u", "root", "-i", "/root/.ssh/tiup_rsa"}, 360)
	if err == nil {
		t.Error("err should not be nil")
	}
}

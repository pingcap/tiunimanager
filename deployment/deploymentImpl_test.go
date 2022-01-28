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
		TestTiDBTopo, testTiUPHome, TestWorkFlowID, []string{"-u", "root", "-i", "/root/.ssh/tiup_rsa"}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_Start(t *testing.T) {
	_, err := manager.Start(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, TestWorkFlowID, []string{}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_Stop(t *testing.T) {
	_, err := manager.Stop(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, TestWorkFlowID, []string{}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_Restart(t *testing.T) {
	_, err := manager.Restart(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, TestWorkFlowID, []string{}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_Upgrade(t *testing.T) {
	_, err := manager.Upgrade(context.TODO(), TiUPComponentTypeCluster, TestClusterID, TestDstVersion, testTiUPHome, TestWorkFlowID, []string{}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_ScaleOut(t *testing.T) {
	_, err := manager.ScaleOut(context.TODO(), TiUPComponentTypeCluster, TestClusterID, TestTiDBScaleOutTopo, testTiUPHome, TestWorkFlowID, []string{"-u", "root", "-i", "/root/.ssh/tiup_rsa"}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_ScaleIn(t *testing.T) {
	_, err := manager.ScaleIn(context.TODO(), TiUPComponentTypeCluster, TestClusterID, TestNodeID, testTiUPHome, TestWorkFlowID, []string{"--force"}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_Destroy(t *testing.T) {
	_, err := manager.Destroy(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, TestWorkFlowID, []string{}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_EditConfig(t *testing.T) {
	_, err := manager.EditConfig(context.TODO(), TiUPComponentTypeCluster, TestClusterID, TestTiDBEditConfigTopo, testTiUPHome, TestWorkFlowID, []string{}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_Reload(t *testing.T) {
	_, err := manager.Reload(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, TestWorkFlowID, []string{}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_List(t *testing.T) {
	_, err := manager.List(context.TODO(), TiUPComponentTypeCluster, testTiUPHome, TestWorkFlowID, []string{"--format", "json"}, 360)
	if err == nil {
		t.Error("nil error")
	}
}

func TestManager_Display(t *testing.T) {
	_, err := manager.Display(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, TestWorkFlowID, []string{"--format", "json"}, 360)
	if err == nil {
		t.Error("nil error")
	}
}

func TestManager_ShowConfig(t *testing.T) {
	_, err := manager.ShowConfig(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, TestWorkFlowID, []string{"--format", "json"}, 360)
	if err == nil {
		t.Error("nil error")
	}
}

func TestManager_Dumpling(t *testing.T) {
	_, err := manager.Dumpling(context.TODO(), testTiUPHome, TestWorkFlowID, []string{"-u", "root",
		"-P", "10000",
		"--host", "127.0.01",
		"--filetype", "sql",
		"-t", "8",
		"-o", "/tmp/test",
		"-r", "200000",
		"-F", "256MiB"}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_Lightning(t *testing.T) {
	_, err := manager.Lightning(context.TODO(), testTiUPHome, TestWorkFlowID, []string{"-config", "tidb-lightning.toml"}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_Push(t *testing.T) {
	_, err := manager.Push(context.TODO(), TiUPComponentTypeCluster, TestClusterID, TestTiDBPushTopo, "/conf/input_tidb.yml", testTiUPHome, TestWorkFlowID, []string{"-N", "127.0.01"}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_Ctl(t *testing.T) {
	_, err := manager.Ctl(context.TODO(), TiUPComponentTypeCluster, TestVersion, "tidb", testTiUPHome, []string{}, 360)
	if err == nil {
		t.Error("nil err")
	}
}

func TestManager_Exec(t *testing.T) {
	_, err := manager.Exec(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, TestWorkFlowID, []string{}, 360)
	if err != nil {
		t.Error(err)
	}
}

func TestManager_GetStatus(t *testing.T) {
	_, err := manager.GetStatus(context.TODO(), "")
	if err == nil {
		t.Error("nil err")
	}
}

func TestManager_startAsyncOperation(t *testing.T) {
	t.Run("no cmd", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "mock_tiup",
		}
		m.startAsyncOperation(context.TODO(), "", "", "", 0)
	})
	t.Run("right cmd", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "ls",
		}
		m.startAsyncOperation(context.TODO(), "", "", "-lh", 1)
	})
}

func TestManager_startSyncOperation(t *testing.T) {
	t.Run("no cmd", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "mock_tiup",
		}
		m.startSyncOperation("", "", 0)
	})
	t.Run("right cmd", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "ls",
		}
		m.startSyncOperation("", "-lh", 1)
	})
}

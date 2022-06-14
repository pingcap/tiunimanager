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
	"errors"
	"os"
	"os/exec"
	"testing"

	asserts "github.com/stretchr/testify/assert"
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
	_, err := manager.Display(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, []string{"--format", "json"}, 360)
	if err == nil {
		t.Error("nil error")
	}
}

func TestManager_ShowConfig(t *testing.T) {
	_, err := manager.ShowConfig(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, []string{"--format", "json"}, 360)
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

func TestManager_Pull(t *testing.T) {
	_, err := manager.Pull(context.TODO(), TiUPComponentTypeCluster, TestClusterID, "/remote/path", testTiUPHome, []string{}, 360)
	if err == nil {
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

func TestManager_CheckConfig(t *testing.T) {
	_, err := manager.CheckConfig(context.TODO(), TiUPComponentTypeCluster, TestTiDBCheckTopo, testTiUPHome, []string{}, 360)
	if err == nil {
		t.Error("nil err")
	}
}

func TestManager_ExtractCheckResult(t *testing.T) {
	type want struct {
		result string
	}
	tests := []struct {
		testName string
		results  []string
		want     want
	}{
		{"Test_WithResult", []string{"{\"prefix\":", "{\"prefix\":", "{\"prefix\":", "{\"result\":", "{\"exit_code\":"}, want{"{\"result\":"}},
		{"Test_WithoutResult", []string{"{\"prefix\":", "{\"prefix\":", "{\"prefix\":", "{\"exit_code\":"}, want{""}},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := manager.extractCheckResult(tt.results)
			if result != tt.want.result {
				t.Errorf("case extract check result %s . err(expected %v, actual %v)", tt.testName, tt.want.result, result)
			}
		})
	}
}

func TestManager_CheckCluster(t *testing.T) {
	_, err := manager.CheckCluster(context.TODO(), TiUPComponentTypeCluster, "clusterId", testTiUPHome, []string{}, 360)
	if err == nil {
		t.Error("nil err")
	}
}

func TestManager_Prune(t *testing.T) {
	_, err := manager.Prune(context.TODO(), TiUPComponentTypeCluster, TestClusterID, testTiUPHome, TestWorkFlowID, []string{}, 360)
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

func TestManager_sensitiveCmd(t *testing.T) {
	t.Run("dumpling", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "mock_tiup",
		}
		asserts.True(t, m.sensitiveCmd("dumpling sth"))
	})
	t.Run("tidb-lightning", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "mock_tiup",
		}
		asserts.True(t, m.sensitiveCmd("tidb-lightning sth"))
	})
	t.Run("deploy", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "mock_tiup",
		}
		asserts.False(t, m.sensitiveCmd("deploy sth"))
	})
}

func TestManager_ExitStatusZero(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "mock_tiup",
		}
		err := exec.ExitError{
			&os.ProcessState{},
			[]byte{},
		}
		asserts.True(t, m.ExitStatusZero(&err))
	})

	t.Run("nonzero", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "mock_tiup",
		}
		err := errors.New("")
		asserts.False(t, m.ExitStatusZero(err))
	})
}

func TestManager_startSyncOperation(t *testing.T) {
	t.Run("no cmd", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "mock_tiup",
		}
		m.startSyncOperation("", "", 0, false)
	})
	t.Run("right cmd", func(t *testing.T) {
		m := &Manager{
			TiUPBinPath: "ls",
		}
		m.startSyncOperation("", "-lh", 1, false)
	})
}

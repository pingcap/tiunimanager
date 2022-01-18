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
 * @File: deploymentImpl
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/7
*******************************************************************************/

package deployment

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/pingcap-inc/tiem/util/disk"

	"github.com/pingcap-inc/tiem/deployment/operation"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

type Manager struct {
	TiUPBinPath string
}

func (m *Manager) Deploy(ctx context.Context, componentType TiUPComponentType, clusterID, version, configYaml, home,
	workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)
	configYamlFilePath, err := disk.NewTmpFileWithContent("em-topology", "yaml", []byte(configYaml))
	if err != nil {
		return "", err
	}

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %s %s %d %s", componentType, CMDDeploy, clusterID, version, configYamlFilePath, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	op, err := GetOperationReaderWriter().Create(ctx, CMDDeploy, workFlowID)
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, op.ID, home, tiUPArgs, timeout)
	return op.ID, nil
}

func (m *Manager) startAsyncOperation(ctx context.Context, id, home, tiUPArgs string, timeoutS int) {
	go func() {
		cmd, cancelFunc := genCommand(home, m.TiUPBinPath, tiUPArgs, timeoutS)
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		defer cancelFunc()

		t0 := time.Now()
		if err := cmd.Start(); err != nil {
			updateStatus(ctx, id, fmt.Sprintf("operation starts err: %+v, errStr: %s", err, stderr.String()), operation.Error, t0)
			return
		}
		updateStatus(ctx, id, "operation processing", operation.Processing, time.Time{})

		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					if status.ExitStatus() == 0 {
						updateStatus(ctx, id, "operation finished", operation.Finished, t0)
						return
					}
				}
			}
			updateStatus(ctx, id, fmt.Sprintf("operation failed with err: %+v, errstr: %s", err, stderr.String()), operation.Error, t0)
			return
		}

		updateStatus(ctx, id, "operation finished", operation.Finished, t0)
		return
	}()
}

func genCommand(home, tiUPBinPath, tiUPArgs string, timeoutS int) (cmd *exec.Cmd, cancelFunc context.CancelFunc) {
	if timeoutS != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutS)*time.Second)
		cancelFunc = cancel
		cmd = exec.CommandContext(ctx, tiUPBinPath, tiUPArgs)
	} else {
		cmd = exec.Command(tiUPBinPath, tiUPArgs)
		cancelFunc = func() {}
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("TIUP_HOME=%s", home))
	cmd.SysProcAttr = genSysProcAttr()
	return
}

func updateStatus(ctx context.Context, operationID, msg string, status operation.Status, t0 time.Time) {
	if t0.IsZero() {
		framework.LogWithContext(ctx).Info(msg)
	} else {
		framework.LogWithContext(ctx).Infof("%s, time cost %v", msg, time.Since(t0))
	}
	GetOperationReaderWriter().Update(ctx, &operation.Operation{
		ID:     operationID,
		Status: status,
	})
}

func (m *Manager) Start(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) Restart(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) Upgrade(ctx context.Context, componentType TiUPComponentType, clusterID, version, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) ScaleIn(ctx context.Context, componentType TiUPComponentType, clusterID, nodeID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) ScaleOut(ctx context.Context, componentType TiUPComponentType, clusterID, configYaml, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) Destroy(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) Reload(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) EditClusterConfig(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, configs map[string]map[string]interface{}, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) EditInstanceConfig(ctx context.Context, componentType TiUPComponentType, clusterID, component, host, home, workFlowID string, config map[string]interface{}, args []string, port, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) List(ctx context.Context, componentType TiUPComponentType, home, workFlowID string, args []string, timeout int) (result string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) Display(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (result string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) ShowConfig(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (spec *spec.Specification, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) Dumpling(ctx context.Context, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) Lightning(ctx context.Context, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) Push(ctx context.Context, componentType TiUPComponentType, clusterID, collectorYaml, remotePath, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) Ctl(ctx context.Context, componentType TiUPComponentType, version, component, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) Exec(ctx context.Context, componentType TiUPComponentType, home, workFlowID string, args []string, timeout int) (result string, err error) {
	//TODO implement me
	panic("implement me")
}
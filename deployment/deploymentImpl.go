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

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

type Manager struct {
	TiUPBinPath string
}

// Deploy
// @Description: wrapper of `tiup <component> deploy`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter version
// @Parameter configYaml
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Deploy(ctx context.Context, componentType TiUPComponentType, clusterID, version, configYaml, home,
	workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)
	// todo: delete the outdated topo file
	configYamlFilePath, err := disk.CreateWithContent("", "em-topology", "yaml", []byte(configYaml))
	if err != nil {
		return "", err
	}

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %s %s %d %s", componentType, CMDDeploy, clusterID, version, configYamlFilePath, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDDeploy,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

func (m *Manager) startAsyncOperation(ctx context.Context, id, home, tiUPArgs string, timeoutS int) {
	go func() {
		cmd, cancelFunc := genCommand(home, m.TiUPBinPath, tiUPArgs, timeoutS)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		defer cancelFunc()

		t0 := time.Now()
		if err := cmd.Start(); err != nil {
			updateStatus(ctx, id, fmt.Sprintf("operation starts err: %+v, errStr: %s", err, stderr.String()), Error, t0)
			return
		}
		updateStatus(ctx, id, "operation processing", Processing, time.Time{})

		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					if status.ExitStatus() == 0 {
						updateStatus(ctx, id, "operation finished", Finished, t0)
						return
					}
				}
			}
			updateStatus(ctx, id, fmt.Sprintf("operation failed with err: %+v, errstr: %s", err, stderr.String()), Error, t0)
			return
		}

		updateStatus(ctx, id, "operation finished", Finished, t0)
		return
	}()
}

func genCommand(home, tiUPBinPath, tiUPArgs string, timeoutS int) (cmd *exec.Cmd, cancelFunc context.CancelFunc) {
	if timeoutS != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutS)*time.Second)
		cancelFunc = cancel
		cmd = exec.CommandContext(ctx, tiUPBinPath, strings.Fields(tiUPArgs)...)
	} else {
		cmd = exec.Command(tiUPBinPath, strings.Fields(tiUPArgs)...)
		cancelFunc = func() {}
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("TIUP_HOME=%s", home))
	cmd.SysProcAttr = genSysProcAttr()
	return
}

func updateStatus(ctx context.Context, operationID, msg string, status Status, t0 time.Time) {
	if t0.IsZero() {
		framework.LogWithContext(ctx).Info(msg)
	} else {
		framework.LogWithContext(ctx).Infof("%s, time cost %v", msg, time.Since(t0))
	}
	op, err := Read(operationID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("Fail update %s to %s: %v", operationID, status, err)
		return
	}

	op.Status = status
	if status == Error {
		op.ErrorStr = msg
	} else {
		op.Result = msg
	}
	err = Update(operationID, op)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("Fail update %s to %s: %v", operationID, status, err)
	}
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

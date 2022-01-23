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
	configYamlFilePath, err := disk.CreateWithContent("", "tidb-deploy-topology", "yaml", []byte(configYaml))
	if err != nil {
		return "", err
	}

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %s %s %d %s", componentType, CMDDeploy, clusterID, version, configYamlFilePath,
		strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
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

// Start
// @Description: wrapper of `tiup <component> start`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Start(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDStart, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDStart,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// Stop
// @Description: wrapper of `tiup <component> stop`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Stop(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDStop, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDStop,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// Restart
// @Description: wrapper of `tiup <component> restart`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Restart(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDRestart, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDRestart,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// Upgrade
// @Description: wrapper of `tiup <component> upgrade`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter version
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Upgrade(ctx context.Context, componentType TiUPComponentType, clusterID, version, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %s %d %s", componentType, CMDUpgrade, clusterID, version, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDUpgrade,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// ScaleOut
// @Description: wrapper of `tiup <component> scale-out`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter configYaml
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) ScaleOut(ctx context.Context, componentType TiUPComponentType, clusterID, configYaml, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)
	// todo: delete the outdated topo file
	configYamlFilePath, err := disk.CreateWithContent("", "tidb-scale-out-topology", "yaml", []byte(configYaml))
	if err != nil {
		return "", err
	}

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %s %d %s", componentType, CMDScaleOut, clusterID, configYamlFilePath, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDScaleOut,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// ScaleIn
// @Description: wrapper of `tiup <component> scale-in`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter nodeID
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) ScaleIn(ctx context.Context, componentType TiUPComponentType, clusterID, nodeID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %s %s %d %s", componentType, CMDScaleIn, clusterID, CMDNode, nodeID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDScaleIn,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// Destroy
// @Description: wrapper of `tiup <component> destroy`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Destroy(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDDestroy, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDDestroy,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// EditConfig
// @Description: wrapper of `tiup <component> edit-config`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter configYaml
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) EditConfig(ctx context.Context, componentType TiUPComponentType, clusterID, configYaml, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)
	// todo: delete the outdated topo file
	configYamlFilePath, err := disk.CreateWithContent("", "tidb-edit-config-topology", "yaml", []byte(configYaml))
	if err != nil {
		return "", err
	}

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %s %s %d %s", componentType, CMDEditConfig, clusterID, CMDTopologyFile, configYamlFilePath, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDEditConfig,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// Reload
// @Description: wrapper of `tiup <component> reload`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Reload(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDReload, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDReload,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// EditClusterConfig
// @Description:
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter workFlowID
// @Parameter configs
// @Parameter args
// @Parameter timeout
// @return ID
// @return err
func (m *Manager) EditClusterConfig(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, configs map[string]map[string]interface{}, args []string, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

// EditInstanceConfig
// @Description:
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter component
// @Parameter host
// @Parameter home
// @Parameter workFlowID
// @Parameter config
// @Parameter args
// @Parameter port
// @Parameter timeout
// @return ID
// @return err
func (m *Manager) EditInstanceConfig(ctx context.Context, componentType TiUPComponentType, clusterID, component, host, home, workFlowID string, config map[string]interface{}, args []string, port, timeout int) (ID string, err error) {
	//TODO implement me
	panic("implement me")
}

// List
// @Description: wrapper of `tiup <component> list`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return result, the response from command
// @return err
func (m *Manager) List(ctx context.Context, componentType TiUPComponentType, home, workFlowID string, args []string, timeout int) (result string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %d %s", componentType, CMDList, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	return m.startSyncOperation(home, tiUPArgs, timeout)
}

// Display
// @Description: wrapper of `tiup <component> display`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return result, the response from command
// @return err
func (m *Manager) Display(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (result string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDDisplay, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	return m.startSyncOperation(home, tiUPArgs, timeout)
}

// ShowConfig
// @Description: wrapper of `tiup <component> show-config`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return result, the response from command
// @return err
func (m *Manager) ShowConfig(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (result string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDShowConfig, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	return m.startSyncOperation(home, tiUPArgs, timeout)
}

// Dumpling
// @Description: wrapper of `tiup dumpling`
// @Receiver m
// @Parameter ctx
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Dumpling(ctx context.Context, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s", CMDDumpling, strings.Join(args, " "))
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDDumpling,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// Lightning
// @Description: wrapper of `tiup lightning`
// @Receiver m
// @Parameter ctx
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Lightning(ctx context.Context, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s", CMDLightning, strings.Join(args, " "))
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDLightning,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// Push
// @Description: wrapper of `tiup cluster push`
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter collectorYaml
// @Parameter remotePath
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Push(ctx context.Context, componentType TiUPComponentType, clusterID, collectorYaml, remotePath, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)
	// todo: delete the outdated topo file
	configYamlFilePath, err := disk.CreateWithContent("", "tidb-push-topology", "yaml", []byte(collectorYaml))
	if err != nil {
		return "", err
	}

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %s %s %d %s", componentType, CMDPush, clusterID, configYamlFilePath, remotePath, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDPush,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// Ctl
// @Description: wrapper of `tiup ctl:<TiDB version> <TiDB component>`
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter version
// @Parameter component
// @Parameter home
// @Parameter args
// @Parameter timeout
// @return result
// @return err
func (m *Manager) Ctl(ctx context.Context, componentType TiUPComponentType, version, component, home string, args []string, timeout int) (result string, err error) {
	logInFunc := framework.LogWithContext(ctx)

	tiUPArgs := fmt.Sprintf("%s:%s %s %s", string(componentType), version, component, strings.Join(args, " "))
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	return m.startSyncOperation(home, tiUPArgs, timeout)
}

// Exec
// @Description: wrapper of `tiup <component> exec`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter workFlowID
// @Parameter args
// @Parameter timeout
// @return ID, operation id to help check the status
// @return err
func (m *Manager) Exec(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDExec, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	id, err := Create(home, Operation{
		Type:       CMDExec,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// GetStatus
// @Description: get status for async operation
// @Receiver m
// @Parameter ctx
// @Parameter ID
// @return op
// @return err
func (m *Manager) GetStatus(ctx context.Context, ID string) (op Operation, err error) {
	framework.LogWithContext(ctx).Infof("getstatus for operationid: %s", ID)
	return Read(ID)
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

func (m *Manager) startSyncOperation(home, tiUPArgs string, timeoutS int) (result string, err error) {
	cmd, cancelFunc := genCommand(home, m.TiUPBinPath, tiUPArgs, timeoutS)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	defer cancelFunc()

	data, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(data), nil
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

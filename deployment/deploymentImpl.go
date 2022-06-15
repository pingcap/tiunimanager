/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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

	"github.com/pingcap/tiunimanager/util/uuidutil"

	"github.com/pingcap/tiunimanager/util/disk"

	"github.com/pingcap/tiunimanager/library/framework"
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDDeploy,
		Operation:  op,
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDStart,
		Operation:  op,
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDStop,
		Operation:  op,
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)
	logInFunc.Infof("env PATH: %s", os.Getenv("PATH"))

	id, err := Create(home, Operation{
		Type:       CMDRestart,
		Operation:  op,
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDUpgrade,
		Operation:  op,
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDScaleOut,
		Operation:  op,
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDScaleIn,
		Operation:  op,
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDDestroy,
		Operation:  op,
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDEditConfig,
		Operation:  op,
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDReload,
		Operation:  op,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
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

	return m.startSyncOperation(home, tiUPArgs, timeout, false)
}

// Display
// @Description: wrapper of `tiup <component> display`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter args
// @Parameter timeout
// @return result, the response from command
// @return err
func (m *Manager) Display(ctx context.Context, componentType TiUPComponentType, clusterID, home string, args []string, timeout int) (result string, err error) {
	logInFunc := framework.LogWithContext(ctx)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDDisplay, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	return m.startSyncOperation(home, tiUPArgs, timeout, false)
}

// ShowConfig
// @Description: wrapper of `tiup <component> show-config`, <component> can be 'cluster', 'dm'
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter args
// @Parameter timeout
// @return result, the response from command
// @return err
func (m *Manager) ShowConfig(ctx context.Context, componentType TiUPComponentType, clusterID, home string, args []string, timeout int) (result string, err error) {
	logInFunc := framework.LogWithContext(ctx)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDShowConfig, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	return m.startSyncOperation(home, tiUPArgs, timeout, false)
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
	//logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s", CMDDumpling, strings.Join(args, " "))
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	// todo op contains password
	//logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDDumpling,
		Operation:  op,
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
	//logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s", CMDLightning, strings.Join(args, " "))
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	// todo op contains password
	//logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDLightning,
		Operation:  op,
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDPush,
		Operation:  op,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// Pull
// @Description: wrapper of `tiup cluster pull`
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter remotePath
// @Parameter localPath
// @Parameter home
// @Parameter args
// @Parameter timeout
// @return result, content of the file pulled from remotePath
// @return err
func (m *Manager) Pull(ctx context.Context, componentType TiUPComponentType, clusterID, remotePath, home string, args []string, timeout int) (result string, err error) {
	logInFunc := framework.LogWithContext(ctx)

	localPath := fmt.Sprintf("/tmp/%s", uuidutil.GenerateID())
	//defer os.Remove(localPath)
	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %s %s %d %s", componentType, CMDPull, clusterID, remotePath, localPath, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	_, err = m.startSyncOperation(home, tiUPArgs, timeout, false)
	if err != nil {
		return
	}

	return disk.ReadFileContent(localPath)
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

	return m.startSyncOperation(home, tiUPArgs, timeout, false)
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
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDExec,
		Operation:  op,
		WorkFlowID: workFlowID,
		Status:     Init,
	})
	if err != nil {
		return "", err
	}

	m.startAsyncOperation(ctx, id, home, tiUPArgs, timeout)
	return id, nil
}

// CheckConfig
// @Description: wrapper of `tiup cluster check topology.yml`
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter configYaml
// @Parameter home
// @Parameter args
// @Parameter timeout
// @return result
// @return err
func (m *Manager) CheckConfig(ctx context.Context, componentType TiUPComponentType, configYaml, home string, args []string, timeout int) (result string, err error) {
	logInFunc := framework.LogWithContext(ctx)
	// todo: delete the outdated topo file
	configYamlFilePath, err := disk.CreateWithContent("", "tidb-check-topology", "yaml", []byte(configYaml))
	if err != nil {
		return "", err
	}

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d", componentType, CMDCheck, configYamlFilePath, strings.Join(args, " "), FlagWaitTimeout, timeout)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("env PATH: %s", os.Getenv("PATH"))

	resp, err := m.startSyncOperation(home, tiUPArgs, timeout, false)
	if err != nil {
		return "", err
	}

	jsons := strings.Split(resp, "\n")
	result = m.extractCheckResult(jsons)
	return
}

// CheckCluster
// @Description: wrapper of `tiup cluster check <cluster-name>`
// @Receiver m
// @Parameter ctx
// @Parameter componentType
// @Parameter clusterID
// @Parameter home
// @Parameter args
// @Parameter timeout
// @return result
// @return err
func (m *Manager) CheckCluster(ctx context.Context, componentType TiUPComponentType, clusterID, home string, args []string, timeout int) (result string, err error) {
	logInFunc := framework.LogWithContext(ctx)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d", componentType, CMDCheck, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout)
	logInFunc.Infof("recv operation req: TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)

	return m.startSyncOperation(home, tiUPArgs, timeout, true)
}

// extract check result from tiup check cluster
func (m *Manager) extractCheckResult(resultJsons []string) (result string) {
	for _, jsonStr := range resultJsons {
		if strings.HasPrefix(jsonStr, "{\"result\":") {
			result = jsonStr
			break
		}
	}
	return
}

func (m *Manager) Prune(ctx context.Context, componentType TiUPComponentType, clusterID, home, workFlowID string, args []string, timeout int) (ID string, err error) {
	logInFunc := framework.LogWithContext(ctx).WithField("workFlowID", workFlowID)

	tiUPArgs := fmt.Sprintf("%s %s %s %s %s %d %s", componentType, CMDPrune, clusterID, strings.Join(args, " "), FlagWaitTimeout, timeout, CMDYes)
	op := fmt.Sprintf("TIUP_HOME=%s %s %s", home, m.TiUPBinPath, tiUPArgs)
	logInFunc.Infof("recv operation req: %s", op)

	id, err := Create(home, Operation{
		Type:       CMDPrune,
		Operation:  op,
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
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		defer cancelFunc()

		t0 := time.Now()
		if err := cmd.Start(); err != nil {
			var detailInfo string
			if m.sensitiveCmd(tiUPArgs) {
				detailInfo = out.String()
			} else {
				detailInfo = fmt.Sprintf("%s\n%s", stderr.String(), out.String())
			}
			updateStatus(ctx, id, fmt.Sprintf("operation starts err: %+v. \ndetail info: %s", err, detailInfo), Error, t0)
			return
		}
		updateStatus(ctx, id, "operation processing", Processing, time.Time{})

		err := cmd.Wait()
		if err != nil && !m.ExitStatusZero(err) {
			var detailInfo string
			if m.sensitiveCmd(tiUPArgs) {
				detailInfo = out.String()
			} else {
				detailInfo = fmt.Sprintf("%s\n%s", stderr.String(), out.String())
			}
			updateStatus(ctx, id, fmt.Sprintf("operation failed with err: %+v. \ndetail info: %s", err, detailInfo), Error, t0)
			return
		}

		updateStatus(ctx, id, "operation finished", Finished, t0)
	}()
}

func (m *Manager) sensitiveCmd(tiUPArgs string) bool {
	cmd := strings.Split(tiUPArgs, " ")[0]
	return cmd == CMDDumpling || cmd == CMDLightning
}

func (m *Manager) ExitStatusZero(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus() == 0
		}
	}
	return false
}

func (m *Manager) startSyncOperation(home, tiUPArgs string, timeoutS int, allInfo bool) (result string, err error) {
	cmd, cancelFunc := genCommand(home, m.TiUPBinPath, tiUPArgs, timeoutS)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	defer cancelFunc()

	data, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s.\ndetail info: %s\n%s", err.Error(), stderr.String(), string(data))
	}
	if allInfo {
		return fmt.Sprintf("%s%s", string(data), stderr.String()), nil
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

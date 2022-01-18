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

/*******************************************************************************
 * @File: lib_tiup_v2
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/8
*******************************************************************************/

package secondparty

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	expect "github.com/google/goexpect"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/workflow/secondparty"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/spec"
	spec2 "github.com/pingcap/tiup/pkg/cluster/spec"
	"gopkg.in/yaml.v2"
)

func (manager *SecondPartyManager) ClusterDeploy(ctx context.Context, tiUPComponent TiUPComponentTypeStr,
	instanceName string, version string, configStrYaml string, timeoutS int, flags []string, workFlowNodeID string, password string) (
	operationID string, err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("clusterdeploy "+
		"tiupcomponent: %s, instancename: %s, version: %s, configstryaml: %s, timeout: %d, flags: %v, workflownodeid: "+
		"%s", string(tiUPComponent), instanceName, version, configStrYaml, timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterDeploy, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var deployReq CmdDeployReq
		deployReq.TiUPComponent = tiUPComponent
		deployReq.InstanceName = instanceName
		deployReq.Version = version
		deployReq.ConfigStrYaml = configStrYaml
		deployReq.TimeoutS = timeoutS
		deployReq.Flags = flags
		deployReq.TiUPPath = manager.TiUPBinPath
		deployReq.Password = password
		deployReq.TiUPHome = GetTiUPHomeForComponent(ctx, tiUPComponent)
		manager.startTiUPDeployOperation(ctx, secondPartyOperation.ID, &deployReq)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPDeployOperation(ctx context.Context, operationID string, req *CmdDeployReq) {
	topologyTmpFilePath, err := newTmpFileWithContent(topologyTmpFilePrefix, []byte(req.ConfigStrYaml))
	if err != nil {
		manager.operationStatusCh <- OperationStatusMember{
			OperationID: operationID,
			Status:      secondparty.OperationStatus_Error,
			Result:      "",
			ErrorStr:    fmt.Sprintln(err),
		}
		return
	}
	go func() {
		//defer os.Remove(topologyTmpFilePath)
		var args []string
		args = append(args, string(req.TiUPComponent), "deploy", req.InstanceName, req.Version, topologyTmpFilePath)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, req.Password)
	}()
}

func (manager *SecondPartyManager) ClusterScaleOut(ctx context.Context, tiUPComponent TiUPComponentTypeStr,
	instanceName string, configStrYaml string, timeoutS int, flags []string, workFlowNodeID string, password string) (
	operationID string, err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("clusterscaleout "+
		"tiupcomponent: %s, instancename: %s, configstryaml: %s, timeout: %d, flags: %v, workflownodeid: %s",
		string(tiUPComponent), instanceName, configStrYaml, timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterScaleOut, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var scaleOutReq CmdScaleOutReq
		scaleOutReq.TiUPComponent = tiUPComponent
		scaleOutReq.InstanceName = instanceName
		scaleOutReq.ConfigStrYaml = configStrYaml
		scaleOutReq.TimeoutS = timeoutS
		scaleOutReq.Flags = flags
		scaleOutReq.TiUPPath = manager.TiUPBinPath
		scaleOutReq.TiUPHome = GetTiUPHomeForComponent(ctx, tiUPComponent)
		scaleOutReq.Password = password
		manager.startTiUPScaleOutOperation(ctx, secondPartyOperation.ID, &scaleOutReq)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPScaleOutOperation(ctx context.Context, operationID string, req *CmdScaleOutReq) {
	topologyTmpFilePath, err := newTmpFileWithContent(topologyTmpFilePrefix, []byte(req.ConfigStrYaml))
	if err != nil {
		manager.operationStatusCh <- OperationStatusMember{
			OperationID: operationID,
			Status:      secondparty.OperationStatus_Error,
			Result:      "",
			ErrorStr:    fmt.Sprintln(err),
		}
		return
	}
	go func() {
		//defer os.Remove(topologyTmpFilePath)
		var args []string
		args = append(args, string(req.TiUPComponent), "scale-out", req.InstanceName, topologyTmpFilePath)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, req.Password)
	}()
}

func (manager *SecondPartyManager) ClusterScaleIn(ctx context.Context, tiUPComponent TiUPComponentTypeStr,
	instanceName string, nodeId string, timeoutS int, flags []string, workFlowNodeID string) (operationID string,
	err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("clusterscalein "+
		"tiupcomponent: %s, instancename: %s, nodeid: %s, timeout: %d, flags: %v, workflownodeid: %s",
		string(tiUPComponent), instanceName, nodeId, timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterScaleIn, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var scaleInReq CmdScaleInReq
		scaleInReq.TiUPComponent = tiUPComponent
		scaleInReq.InstanceName = instanceName
		scaleInReq.NodeId = nodeId
		scaleInReq.TimeoutS = timeoutS
		scaleInReq.Flags = flags
		scaleInReq.TiUPPath = manager.TiUPBinPath
		scaleInReq.TiUPHome = GetTiUPHomeForComponent(ctx, tiUPComponent)
		manager.startTiUPScaleInOperation(ctx, secondPartyOperation.ID, &scaleInReq)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPScaleInOperation(ctx context.Context, operationID string, req *CmdScaleInReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "scale-in", req.InstanceName, "--node", req.NodeId)
		args = append(args, req.Flags...)
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) ClusterStart(ctx context.Context, tiUPComponent TiUPComponentTypeStr,
	instanceName string, timeoutS int, flags []string, workFlowNodeID string) (operationID string, err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("clusterstart "+
		"tiupComponent: %s, instancename: %s, timeout: %d, flags: %v, workflownodeid: %s", string(tiUPComponent),
		instanceName, timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterStart, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var req CmdStartReq
		req.TiUPComponent = tiUPComponent
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiUPPath = manager.TiUPBinPath
		req.TiUPHome = GetTiUPHomeForComponent(ctx, tiUPComponent)
		req.Flags = flags
		manager.startTiUPStartOperation(ctx, secondPartyOperation.ID, &req)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPStartOperation(ctx context.Context, operationID string, req *CmdStartReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "start", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) ClusterRestart(ctx context.Context, tiUPComponent TiUPComponentTypeStr,
	instanceName string, timeoutS int, flags []string, workFlowNodeID string) (operationID string, err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("clusterrestart "+
		"tiupcomponent: %s, instancename: %s, timeout: %d, flags: %v, workflownodeid: %s", string(tiUPComponent),
		instanceName, timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterRestart, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var req CmdStartReq
		req.TiUPComponent = tiUPComponent
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiUPPath = manager.TiUPBinPath
		req.TiUPHome = GetTiUPHomeForComponent(ctx, tiUPComponent)
		req.Flags = flags
		manager.startTiUPRestartOperation(ctx, secondPartyOperation.ID, &req)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPRestartOperation(ctx context.Context, operationID string, req *CmdStartReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "restart", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) ClusterStop(ctx context.Context, tiUPComponent TiUPComponentTypeStr,
	instanceName string, timeoutS int, flags []string, workFlowNodeID string) (operationID string, err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("clusterstop "+
		"tiupComponent: %s, instancename: %s, timeout: %d, flags: %v, workflownodeid: %s", string(tiUPComponent),
		instanceName, timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterStop, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var req CmdStartReq
		req.TiUPComponent = tiUPComponent
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiUPPath = manager.TiUPBinPath
		req.TiUPHome = GetTiUPHomeForComponent(ctx, tiUPComponent)
		req.Flags = flags
		manager.startTiUPStopOperation(ctx, secondPartyOperation.ID, &req)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPStopOperation(ctx context.Context, operationID string, req *CmdStartReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "stop", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) ClusterList(ctx context.Context, tiUPComponent TiUPComponentTypeStr, timeoutS int,
	flags []string) (resp *CmdListResp, err error) {
	framework.LogWithContext(ctx).Infof("clusterlist tiupComponent: %s, timeout: %d, flags: %v",
		string(tiUPComponent), timeoutS, flags)
	var req CmdListReq
	req.TiUPComponent = tiUPComponent
	req.TimeoutS = timeoutS
	req.TiUPPath = manager.TiUPBinPath
	req.TiUPHome = GetTiUPHomeForComponent(ctx, tiUPComponent)
	req.Flags = flags
	cmdListResp, err := manager.startTiUPListOperation(ctx, &req)
	return &cmdListResp, err
}

func (manager *SecondPartyManager) startTiUPListOperation(ctx context.Context, req *CmdListReq) (resp CmdListResp,
	err error) {
	var args []string
	args = append(args, string(req.TiUPComponent), "list")
	args = append(args, req.Flags...)
	args = append(args, "--yes")

	logInFunc := framework.LogWithContext(ctx)
	logInFunc.Info("operation starts processing:", fmt.Sprintf("tiuppath:%s tiupargs:%v timeouts:%d",
		req.TiUPPath, args, req.TimeoutS))
	var cmd *exec.Cmd
	var cancelFp context.CancelFunc
	if req.TimeoutS != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.TimeoutS)*time.Second)
		cancelFp = cancel
		cmd = exec.CommandContext(ctx, req.TiUPPath, args...)
	} else {
		cmd = exec.Command(req.TiUPPath, args...)
		cancelFp = func() {}
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("TIUP_HOME=%s", req.TiUPHome))
	defer cancelFp()
	cmd.SysProcAttr = genSysProcAttr()
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	var data []byte
	if data, err = cmd.Output(); err != nil {
		logInFunc.Errorf("cmd start err: %+v, errstr: %s", err, stderr.String())
		err = fmt.Errorf("cmd start err: %+v, errstr: %s", err, stderr.String())
		return
	}
	resp.ListRespStr = string(data)
	return
}

func (manager *SecondPartyManager) ClusterDestroy(ctx context.Context, tiUPComponent TiUPComponentTypeStr,
	instanceName string, timeoutS int, flags []string, workFlowNodeID string) (operationID string, err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("clusterdestroy "+
		"tiupComponent: %s, instancename: %s, timeout: %d, flags: %v, workflownodeid: %s", string(tiUPComponent),
		instanceName, timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterDestroy, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var req CmdDestroyReq
		req.TiUPComponent = tiUPComponent
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiUPPath = manager.TiUPBinPath
		req.TiUPHome = GetTiUPHomeForComponent(ctx, tiUPComponent)
		req.Flags = flags
		manager.startTiUPDestroyOperation(ctx, secondPartyOperation.ID, &req)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPDestroyOperation(ctx context.Context, operationID string, req *CmdDestroyReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "destroy", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) ClusterDisplay(ctx context.Context, tiUPComponent TiUPComponentTypeStr,
	instanceName string, timeoutS int, flags []string) (resp *CmdDisplayResp, err error) {
	framework.LogWithContext(ctx).Infof("clusterdisplay tiupcomponent: %s,  instanceName: %s, "+
		"timeouts: %d, flags: %v", string(tiUPComponent), instanceName, timeoutS, flags)
	var args []string
	args = append(args, string(tiUPComponent), "display")
	args = append(args, instanceName)
	args = append(args, flags...)
	tiUPHome := GetTiUPHomeForComponent(ctx, tiUPComponent)
	result, err := manager.startSyncTiUPOperation(ctx, args, timeoutS, tiUPHome)
	resp = &CmdDisplayResp{}
	resp.DisplayRespString = result
	return
}

// extract check result from tiup check cluster
func (manager *SecondPartyManager) extractCheckResult(resultJsons []string) (result string) {
	for _, jsonStr := range resultJsons {
		if strings.HasPrefix(jsonStr, "{\"result\":") {
			result = jsonStr
			break
		}
	}
	return
}

func (manager *SecondPartyManager) CheckTopo(ctx context.Context, tiUPComponent TiUPComponentTypeStr, topoStr string,
	timeoutS int, flags []string) (result string, err error) {
	topologyTmpFilePath, err := newTmpFileWithContent(topologyTmpFilePrefix, []byte(topoStr))
	if err != nil {
		return "", err
	}
	framework.LogWithContext(ctx).Infof("check tiupcomponent: %s,  topostr: %s, "+
		"timeouts: %d, flags: %v", string(tiUPComponent), topoStr, timeoutS, flags)
	var args []string
	args = append(args, string(tiUPComponent), "check")
	args = append(args, topologyTmpFilePath)
	args = append(args, flags...)
	tiUPHome := GetTiUPHomeForComponent(ctx, tiUPComponent)

	//defer os.Remove(topologyTmpFilePath)
	resp, err := manager.startSyncTiUPOperation(ctx, args, timeoutS, tiUPHome)
	if err != nil {
		return "", err
	}

	jsons := strings.Split(resp, "\n")
	result = manager.extractCheckResult(jsons)
	return
}

func (manager *SecondPartyManager) ClusterUpgrade(ctx context.Context, tiUPComponent TiUPComponentTypeStr,
	instanceName string, version string, timeoutS int, flags []string, workFlowNodeID string) (operationID string,
	err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("clusterupgrade "+
		"tiupcomponent: %s, instancename: %s, version: %s, timeouts: %d, flags: %v, workflownodeid: %s",
		string(tiUPComponent), instanceName,
		version, timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterUpgrade, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var req CmdUpgradeReq
		req.TiUPComponent = tiUPComponent
		req.InstanceName = instanceName
		req.Version = version
		req.TimeoutS = timeoutS
		req.Flags = flags
		req.TiUPPath = manager.TiUPBinPath
		req.TiUPHome = GetTiUPHomeForComponent(ctx, tiUPComponent)
		manager.startTiUPUpgradeOperation(ctx, secondPartyOperation.ID, &req)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPUpgradeOperation(ctx context.Context, operationID string, req *CmdUpgradeReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "upgrade", req.InstanceName, req.Version)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) ClusterShowConfig(ctx context.Context, req *CmdShowConfigReq) (
	resp *CmdShowConfigResp, err error) {
	framework.LogWithContext(ctx).Infof("clustershowconfig cmdshowconfigreq: %v", req)
	var args []string
	args = append(args, string(req.TiUPComponent), "show-config")
	args = append(args, req.InstanceName)
	args = append(args, req.Flags...)
	tiUPHome := GetTiUPHomeForComponent(ctx, req.TiUPComponent)

	topoStr, err := manager.startSyncTiUPOperation(ctx, args, req.TimeoutS, tiUPHome)
	if err != nil {
		return nil, err
	}

	topo := &spec2.Specification{}
	if err = yaml.UnmarshalStrict([]byte(topoStr), topo); err != nil {
		framework.LogWithContext(ctx).Errorf("parse original config(%s) error: %+v", topoStr, err)
		return
	}

	resp = &CmdShowConfigResp{topo}
	resp.TiDBClusterTopo = topo
	return
}

func (manager *SecondPartyManager) ClusterEditGlobalConfig(ctx context.Context,
	cmdEditGlobalConfigReq CmdEditGlobalConfigReq, workFlowNodeID string) (string, error) {
	framework.LogWithContext(ctx).Infof("clustereditglobalconfig cmdeditglobalconfigreq: %v, workflownodeid: %s",
		cmdEditGlobalConfigReq, workFlowNodeID)

	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterEditGlobalConfig, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	}

	cmdShowConfigReq := CmdShowConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  cmdEditGlobalConfigReq.InstanceName,
	}
	cmdShowConfigResp, err := manager.ClusterShowConfig(ctx, &cmdShowConfigReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("check orignal config error: %+v", err)
		manager.operationStatusCh <- OperationStatusMember{
			OperationID: secondPartyOperation.ID,
			Status:      secondparty.OperationStatus_Error,
			Result:      "",
			ErrorStr:    fmt.Sprintln(err),
		}
		return secondPartyOperation.ID, err
	}
	topo := cmdShowConfigResp.TiDBClusterTopo

	manager.startTiUPEditGlobalConfigOperation(ctx, secondPartyOperation.ID, &cmdEditGlobalConfigReq, topo)
	return secondPartyOperation.ID, nil
}

func (manager *SecondPartyManager) startTiUPEditGlobalConfigOperation(ctx context.Context, operationID string,
	req *CmdEditGlobalConfigReq, topo *spec2.Specification) {
	var componentServerConfigs map[string]interface{}

	for _, globalComponentConfig := range req.GlobalComponentConfigs {
		switch globalComponentConfig.TiDBClusterComponent {
		case spec.TiDBClusterComponent_TiDB:
			componentServerConfigs = topo.ServerConfigs.TiDB
		case spec.TiDBClusterComponent_TiKV:
			componentServerConfigs = topo.ServerConfigs.TiKV
		case spec.TiDBClusterComponent_PD:
			componentServerConfigs = topo.ServerConfigs.PD
		case spec.TiDBClusterComponent_TiFlash:
			componentServerConfigs = topo.ServerConfigs.TiFlash
		case spec.TiDBClusterComponent_TiFlashLearner:
			componentServerConfigs = topo.ServerConfigs.TiFlashLearner
		case spec.TiDBClusterComponent_Pump:
			componentServerConfigs = topo.ServerConfigs.Pump
		case spec.TiDBClusterComponent_Drainer:
			componentServerConfigs = topo.ServerConfigs.Drainer
		case spec.TiDBClusterComponent_CDC:
			componentServerConfigs = topo.ServerConfigs.CDC
		}
		if componentServerConfigs == nil {
			componentServerConfigs = make(map[string]interface{})
		}
		for k, v := range globalComponentConfig.ConfigMap {
			componentServerConfigs[k] = v
		}
		switch globalComponentConfig.TiDBClusterComponent {
		case spec.TiDBClusterComponent_TiDB:
			topo.ServerConfigs.TiDB = componentServerConfigs
		case spec.TiDBClusterComponent_TiKV:
			topo.ServerConfigs.TiKV = componentServerConfigs
		case spec.TiDBClusterComponent_PD:
			topo.ServerConfigs.PD = componentServerConfigs
		case spec.TiDBClusterComponent_TiFlash:
			topo.ServerConfigs.TiFlash = componentServerConfigs
		case spec.TiDBClusterComponent_TiFlashLearner:
			topo.ServerConfigs.TiFlashLearner = componentServerConfigs
		case spec.TiDBClusterComponent_Pump:
			topo.ServerConfigs.Pump = componentServerConfigs
		case spec.TiDBClusterComponent_Drainer:
			topo.ServerConfigs.Drainer = componentServerConfigs
		case spec.TiDBClusterComponent_CDC:
			topo.ServerConfigs.CDC = componentServerConfigs
		}
	}

	cmdEditConfigReq := CmdEditConfigReq{
		TiUPComponent: req.TiUPComponent,
		InstanceName:  req.InstanceName,
		NewTopo:       topo,
		TiUPHome:      GetTiUPHomeForComponent(ctx, req.TiUPComponent),
		TimeoutS:      req.TimeoutS,
		Flags:         req.Flags,
	}
	manager.startTiUPEditConfigOperation(ctx, cmdEditConfigReq, operationID)
}

func (manager *SecondPartyManager) ClusterEditInstanceConfig(ctx context.Context,
	cmdEditInstanceConfigReq CmdEditInstanceConfigReq, workFlowNodeID string) (string, error) {
	framework.LogWithContext(ctx).Infof("clustereditinstanceConfig cmdeditinstanceconfigreq: %v, "+
		"workflownodeid: %s", cmdEditInstanceConfigReq, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterEditInstanceConfig, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	}

	cmdShowConfigReq := CmdShowConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName:  cmdEditInstanceConfigReq.InstanceName,
	}
	cmdShowConfigResp, err := manager.ClusterShowConfig(ctx, &cmdShowConfigReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("check orignal config error: %+v", err)
		manager.operationStatusCh <- OperationStatusMember{
			OperationID: secondPartyOperation.ID,
			Status:      secondparty.OperationStatus_Error,
			Result:      "",
			ErrorStr:    fmt.Sprintln(err),
		}
		return secondPartyOperation.ID, err
	}
	topo := cmdShowConfigResp.TiDBClusterTopo

	manager.startTiUPEditInstanceConfigOperation(ctx, secondPartyOperation.ID, &cmdEditInstanceConfigReq, topo)
	return secondPartyOperation.ID, nil
}

func (manager *SecondPartyManager) startTiUPEditInstanceConfigOperation(ctx context.Context, operationID string,
	req *CmdEditInstanceConfigReq, topo *spec2.Specification) {
	switch req.TiDBClusterComponent {
	case spec.TiDBClusterComponent_TiDB:
		for idx, tiDBServer := range topo.TiDBServers {
			if tiDBServer.Host == req.Host && tiDBServer.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, tiDBServer, FieldKey_Yaml, k, v)
				}
				topo.TiDBServers[idx] = tiDBServer
				break
			}
		}
	case spec.TiDBClusterComponent_TiKV:
		for idx, tiKVServer := range topo.TiKVServers {
			if tiKVServer.Host == req.Host && tiKVServer.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, tiKVServer, FieldKey_Yaml, k, v)
				}
				topo.TiKVServers[idx] = tiKVServer
				break
			}
		}
	case spec.TiDBClusterComponent_TiFlash:
		for idx, tiFlashServer := range topo.TiFlashServers {
			if tiFlashServer.Host == req.Host && tiFlashServer.FlashServicePort == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, tiFlashServer, FieldKey_Yaml, k, v)
				}
				topo.TiFlashServers[idx] = tiFlashServer
				break
			}
		}
	case spec.TiDBClusterComponent_PD:
		for idx, pdServer := range topo.PDServers {
			if pdServer.Host == req.Host && pdServer.ClientPort == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, pdServer, FieldKey_Yaml, k, v)
				}
				topo.PDServers[idx] = pdServer
				break
			}
		}
	case spec.TiDBClusterComponent_Pump:
		for idx, pumpServer := range topo.PumpServers {
			if pumpServer.Host == req.Host && pumpServer.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, pumpServer, FieldKey_Yaml, k, v)
				}
				topo.PumpServers[idx] = pumpServer
				break
			}
		}
	case spec.TiDBClusterComponent_Drainer:
		for idx, drainer := range topo.Drainers {
			if drainer.Host == req.Host && drainer.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, drainer, FieldKey_Yaml, k, v)
				}
				topo.Drainers[idx] = drainer
				break
			}
		}
	case spec.TiDBClusterComponent_CDC:
		for idx, cdcServer := range topo.CDCServers {
			if cdcServer.Host == req.Host && cdcServer.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, cdcServer, FieldKey_Yaml, k, v)
				}
				topo.CDCServers[idx] = cdcServer
				break
			}
		}
	case spec.TiDBClusterComponent_TiSparkMasters:
		for idx, tiSparkMaster := range topo.TiSparkMasters {
			if tiSparkMaster.Host == req.Host && tiSparkMaster.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, tiSparkMaster, FieldKey_Yaml, k, v)
				}
				topo.TiSparkMasters[idx] = tiSparkMaster
				break
			}
		}
	case spec.TiDBClusterComponent_TiSparkWorkers:
		for idx, tiSparkWorker := range topo.TiSparkWorkers {
			if tiSparkWorker.Host == req.Host && tiSparkWorker.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, tiSparkWorker, FieldKey_Yaml, k, v)
				}
				topo.TiSparkWorkers[idx] = tiSparkWorker
				break
			}
		}
	case spec.TiDBClusterComponent_Prometheus:
		for idx, monitor := range topo.Monitors {
			if monitor.Host == req.Host && monitor.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, monitor, FieldKey_Yaml, k, v)
				}
				topo.Monitors[idx] = monitor
				break
			}
		}
	case spec.TiDBClusterComponent_Grafana:
		for idx, grafana := range topo.Grafanas {
			if grafana.Host == req.Host && grafana.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, grafana, FieldKey_Yaml, k, v)
				}
				topo.Grafanas[idx] = grafana
				break
			}
		}
	case spec.TiDBClusterComponent_Alertmanager:
		for idx, alertManager := range topo.Alertmanagers {
			if alertManager.Host == req.Host && alertManager.WebPort == req.Port {
				for k, v := range req.ConfigMap {
					SetField(ctx, alertManager, FieldKey_Yaml, k, v)
				}
				topo.Alertmanagers[idx] = alertManager
				break
			}
		}
	}

	cmdEditConfigReq := CmdEditConfigReq{
		TiUPComponent: req.TiUPComponent,
		InstanceName:  req.InstanceName,
		NewTopo:       topo,
		TiUPHome:      GetTiUPHomeForComponent(ctx, req.TiUPComponent),
		TimeoutS:      req.TimeoutS,
		Flags:         req.Flags,
	}
	manager.startTiUPEditConfigOperation(ctx, cmdEditConfigReq, operationID)
}

func (manager *SecondPartyManager) startTiUPEditConfigOperation(ctx context.Context, req CmdEditConfigReq,
	operationID string) {
	newData, err := yaml.Marshal(req.NewTopo)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("startTiUPeditconfigoperation marshal new config(%+v) error: %+v",
			req.NewTopo, err)
		manager.operationStatusCh <- OperationStatusMember{
			OperationID: operationID,
			Status:      secondparty.OperationStatus_Error,
			Result:      "",
			ErrorStr:    fmt.Sprintf("startTiUPeditconfigoperation marshal new config(%+v) error: %+v", req.NewTopo, err),
		}
		return
	}

	topologyTmpFilePath, err := newTmpFileWithContent("tidb-cluster-topology", newData)
	if err != nil {
		manager.operationStatusCh <- OperationStatusMember{
			OperationID: operationID,
			Status:      secondparty.OperationStatus_Error,
			Result:      "",
			ErrorStr:    fmt.Sprintln(err),
		}
		return
	}
	go func() {
		//defer os.Remove(topologyTmpFilePath)
		var args []string
		args = append(args, string(req.TiUPComponent), "edit-config", req.InstanceName, "--topology-file", topologyTmpFilePath)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, manager.TiUPBinPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) ClusterReload(ctx context.Context, cmdReloadConfigReq CmdReloadConfigReq,
	workFlowNodeID string) (operationID string, err error) {
	framework.LogWithContext(ctx).Infof("clusterreload cmdreloadconfigreq: %v, workflownodeid: %s",
		cmdReloadConfigReq, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterReload, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	}
	cmdReloadConfigReq.TiUPHome = GetTiUPHomeForComponent(ctx, cmdReloadConfigReq.TiUPComponent)
	manager.startTiUPReloadOperation(ctx, secondPartyOperation.ID, &cmdReloadConfigReq)
	return secondPartyOperation.ID, nil
}

func (manager *SecondPartyManager) startTiUPReloadOperation(ctx context.Context, operationID string, req *CmdReloadConfigReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "reload", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, manager.TiUPBinPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) ClusterExec(ctx context.Context, cmdClusterExecReq CmdClusterExecReq,
	workFlowNodeID string) (operationID string, err error) {
	framework.LogWithContext(ctx).Infof("clusterexec cmdclusterexecreq: %v, workflownodeid: %s",
		cmdClusterExecReq, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_ClusterExec, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	}
	cmdClusterExecReq.TiUPHome = GetTiUPHomeForComponent(ctx, cmdClusterExecReq.TiUPComponent)
	manager.startTiUPExecOperation(ctx, secondPartyOperation.ID, &cmdClusterExecReq)
	return secondPartyOperation.ID, nil
}

func (manager *SecondPartyManager) startTiUPExecOperation(ctx context.Context, operationID string, req *CmdClusterExecReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "exec", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, manager.TiUPBinPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) Dumpling(ctx context.Context, timeoutS int, flags []string, workFlowNodeID string) (
	operationID string, err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("dumpling, timeouts: %d, "+
		"flags: %v, workflownodeid: %s", timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_Dumpling, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var dumplingReq CmdDumplingReq
		dumplingReq.TimeoutS = timeoutS
		dumplingReq.TiUPPath = manager.TiUPBinPath
		dumplingReq.Flags = flags
		dumplingReq.TiUPHome = GetTiUPHomeForComponent(ctx, DefaultComponentTypeStr)
		manager.startTiUPDumplingOperation(ctx, secondPartyOperation.ID, &dumplingReq)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPDumplingOperation(ctx context.Context, operationID string, req *CmdDumplingReq) {
	go func() {
		var args []string
		args = append(args, "dumpling")
		args = append(args, req.Flags...)
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) Lightning(ctx context.Context, timeoutS int, flags []string, workFlowNodeID string) (
	operationID string, err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("lightning, timeouts: %d, "+
		"flags: %v, workflownodeid: %s", timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_Lightning, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var lightningReq CmdLightningReq
		lightningReq.TimeoutS = timeoutS
		lightningReq.TiUPPath = manager.TiUPBinPath
		lightningReq.Flags = flags
		lightningReq.TiUPHome = GetTiUPHomeForComponent(ctx, DefaultComponentTypeStr)
		manager.startTiUPLightningOperation(ctx, secondPartyOperation.ID, &lightningReq)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPLightningOperation(ctx context.Context, operationID string, req *CmdLightningReq) {
	go func() {
		var args []string
		args = append(args, "tidb-lightning")
		args = append(args, req.Flags...)
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) Transfer(ctx context.Context, tiUPComponent TiUPComponentTypeStr,
	instanceName string, collectorYaml string, remotePath string, timeoutS int, flags []string, workFlowNodeID string) (
	operationID string, err error) {
	framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID).Infof("transfer tiupcomponent: %s, "+
		"instancename: %s, collectoryaml: %s, remotepath: %s, timeouts: %d, flags: %v, workflownodeid: %s",
		string(tiUPComponent), instanceName, collectorYaml, remotePath, timeoutS, flags, workFlowNodeID)
	secondPartyOperation, err := models.GetSecondPartyOperationReaderWriter().Create(ctx,
		secondparty.OperationType_Transfer, workFlowNodeID)
	if secondPartyOperation == nil || err != nil {
		err = fmt.Errorf("secondpartyoperation:%v, err:%v", secondPartyOperation, err)
		return "", err
	} else {
		var req CmdTransferReq
		req.TiUPComponent = tiUPComponent
		req.InstanceName = instanceName
		req.CollectorYaml = collectorYaml
		req.RemotePath = remotePath
		req.TiUPHome = GetTiUPHomeForComponent(ctx, tiUPComponent)
		req.TimeoutS = timeoutS
		req.Flags = flags
		req.TiUPPath = manager.TiUPBinPath
		manager.startTiUPTransferOperation(ctx, secondPartyOperation.ID, &req)
		return secondPartyOperation.ID, nil
	}
}

func (manager *SecondPartyManager) startTiUPTransferOperation(ctx context.Context, operationID string, req *CmdTransferReq) {
	collectorTmpFilePath, err := newTmpFileWithContent(collectorTmpFilePrefix, []byte(req.CollectorYaml))
	if err != nil {
		manager.operationStatusCh <- OperationStatusMember{
			OperationID: operationID,
			Status:      secondparty.OperationStatus_Error,
			Result:      "",
			ErrorStr:    fmt.Sprintln(err),
		}
		return
	}
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "push", req.InstanceName, collectorTmpFilePath, req.RemotePath)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-manager.startTiUPOperation(ctx, operationID, req.TiUPPath, args, req.TimeoutS, req.TiUPHome, "")
	}()
}

func (manager *SecondPartyManager) ClusterComponentCtl(ctx context.Context, str TiUPComponentTypeStr,
	clusterVersion string, component spec.TiDBClusterComponent, flags []string, timeoutS int) (string, error) {
	framework.LogWithContext(ctx).Infof("clustercomponentctl tiupcomponent: %s,  clusterversion: %s. "+
		"component: %s, flags: %v", string(str), clusterVersion, string(component), flags)
	var args []string
	args = append(args, fmt.Sprintf("%s:%s", string(str), clusterVersion))
	args = append(args, string(component))
	args = append(args, flags...)
	tiUPHome := GetTiUPHomeForComponent(ctx, str)
	return manager.startSyncTiUPOperation(ctx, args, timeoutS, tiUPHome)
}

func (manager *SecondPartyManager) startSyncTiUPOperation(ctx context.Context, args []string,
	timeoutS int, tiUPHome string) (result string, err error) {
	logInFunc := framework.LogWithContext(ctx)
	logInFunc.Info("operation starts processing:", fmt.Sprintf("TIUP_HOME=%s tiuppath:%s tiupargs:%v timeouts:%d",
		tiUPHome, manager.TiUPBinPath, args, timeoutS))
	var cmd *exec.Cmd
	var cancelFp context.CancelFunc
	if timeoutS != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutS)*time.Second)
		cancelFp = cancel
		cmd = exec.CommandContext(ctx, manager.TiUPBinPath, args...)
	} else {
		cmd = exec.Command(manager.TiUPBinPath, args...)
		cancelFp = func() {}
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("TIUP_HOME=%s", tiUPHome))
	defer cancelFp()
	cmd.SysProcAttr = genSysProcAttr()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		logInFunc.Errorf("cmd start err: %+v, errstr: %s", err, stderr.String())
		err = fmt.Errorf("cmd start err: %+v, errstr: %s", err, stderr.String())
		return "", err
	}
	return string(data), nil
}

func (manager *SecondPartyManager) startTiUPOperation(ctx context.Context, operationID string, tiUPPath string,
	tiUPArgs []string, TimeoutS int, tiUPHome string, password string) (exitCh chan struct{}) {
	exitCh = make(chan struct{})
	logInFunc := framework.LogWithContext(ctx).WithField("operation", operationID)
	logInFunc.Info("operation starts processing:", fmt.Sprintf("TIUP_HOME=%s tiuppath:%s tiupargs:%v timeouts:%d",
		tiUPHome, tiUPPath, tiUPArgs, TimeoutS))
	manager.operationStatusCh <- OperationStatusMember{
		OperationID: operationID,
		Status:      secondparty.OperationStatus_Processing,
		Result:      "",
		ErrorStr:    "",
	}
	go func() {
		defer close(exitCh)
		timeout := time.Duration(TimeoutS) * time.Second
		t0 := time.Now()
		successFp := func() {
			logInFunc.Info("operation finished, time cost", time.Since(t0))
			manager.operationStatusCh <- OperationStatusMember{
				OperationID: operationID,
				Status:      secondparty.OperationStatus_Finished,
				Result:      "",
				ErrorStr:    "",
			}
		}

		if password == "" {
			var cmd *exec.Cmd
			var cancelFp context.CancelFunc
			if TimeoutS != 0 {
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				cancelFp = cancel
				cmd = exec.CommandContext(ctx, tiUPPath, tiUPArgs...)
			} else {
				cmd = exec.Command(tiUPPath, tiUPArgs...)
				cancelFp = func() {}
			}
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, fmt.Sprintf("TIUP_HOME=%s", tiUPHome))
			defer cancelFp()
			cmd.SysProcAttr = genSysProcAttr()
			var out, stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr

			if err := cmd.Start(); err != nil {
				logInFunc.Errorf("cmd start err: %+v, errStr: %s", err, stderr.String())
				manager.operationStatusCh <- OperationStatusMember{
					OperationID: operationID,
					Status:      secondparty.OperationStatus_Error,
					Result:      "",
					ErrorStr:    fmt.Sprintf("cmd start err: %+v, errStr: %s", err, stderr.String()),
				}
				return
			}
			logInFunc.Info("cmd started and waiting")
			err := cmd.Wait()
			if err != nil {
				logInFunc.Errorf("cmd wait return with err: %+v, errStr: %s", err, stderr.String())
				if exiterr, ok := err.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						if status.ExitStatus() == 0 {
							successFp()
							return
						}
					}
				}
				logInFunc.Errorf("cmd wait return with err: %+v, errstr: %s, time cost: %v", err, stderr.String(),
					time.Since(t0))
				manager.operationStatusCh <- OperationStatusMember{
					OperationID: operationID,
					Status:      secondparty.OperationStatus_Error,
					Result:      "",
					ErrorStr: fmt.Sprintf("cmd wait return with err: %+v, errStr: %s, time cost: %v", err,
						stderr.String(), time.Since(t0)),
				}
				return
			} else {
				logInFunc.Info("cmd wait return successfully")
				successFp()
				return
			}
		} else {

			cmd := fmt.Sprintf("%s %s", tiUPPath, strings.Join(tiUPArgs, " "))
			e, _, err := expect.Spawn(cmd, timeout, expect.Verbose(true), expect.SetEnv([]string{fmt.Sprintf("TIUP_HOME=%s", tiUPHome)}))
			if err != nil {
				logInFunc.Errorf("cmd(TIUP_HOME=%s %s) spawned return with err: %+v, time cost: %v", tiUPHome, cmd, err, time.Since(t0))
				manager.operationStatusCh <- OperationStatusMember{
					OperationID: operationID,
					Status:      secondparty.OperationStatus_Error,
					Result:      "",
					ErrorStr:    fmt.Sprintf("cmd(TIUP_HOME=%s %s) spawned return with err: %+v, time cost: %v", tiUPHome, cmd, err, time.Since(t0)),
				}
				return
			}
			defer e.Close()

			_, _, err = e.Expect(regexp.MustCompile(".*Input SSH password:"), timeout)
			if err != nil {
				logInFunc.Errorf("cmd(TIUP_HOME=%s %s) expect(Input SSH password:) return with err: %+v, time cost: %v", tiUPHome, cmd, err, time.Since(t0))
				manager.operationStatusCh <- OperationStatusMember{
					OperationID: operationID,
					Status:      secondparty.OperationStatus_Error,
					Result:      "",
					ErrorStr:    fmt.Sprintf("cmd(TIUP_HOME=%s %s) expect(Input SSH password:) return with err: %+v, time cost: %v", tiUPHome, cmd, err, time.Since(t0)),
				}
				return
			}
			err = e.Send(fmt.Sprintf("%s\n", password))
			if err != nil {
				logInFunc.Errorf("cmd(TIUP_HOME=%s %s) sent password(%s) return with err: %+v, time cost: %v", tiUPHome, cmd, password, err, time.Since(t0))
				manager.operationStatusCh <- OperationStatusMember{
					OperationID: operationID,
					Status:      secondparty.OperationStatus_Error,
					Result:      "",
					ErrorStr:    fmt.Sprintf("cmd(TIUP_HOME=%s %s) sent password(%s) return with err: %+v, time cost: %v", tiUPHome, cmd, password, err, time.Since(t0)),
				}
				return
			}
			_, _, err = e.Expect(regexp.MustCompile(".*successfully.*"), timeout)
			if err != nil {
				logInFunc.Errorf("cmd(TIUP_HOME=%s %s) expect(successfully) return with err: %+v, time cost: %v", tiUPHome, cmd, err, time.Since(t0))
				manager.operationStatusCh <- OperationStatusMember{
					OperationID: operationID,
					Status:      secondparty.OperationStatus_Error,
					Result:      "",
					ErrorStr:    fmt.Sprintf("cmd(TIUP_HOME=%s %s) expect(successfully) return with err: %+v, time cost: %v", tiUPHome, cmd, err, time.Since(t0)),
				}
				return
			}

			logInFunc.Infof("cmd(%s) wait return successfully", cmd)
			successFp()
			return
		}
	}()
	return exitCh
}

func GetTiUPHomeForComponent(ctx context.Context, tiUPComponent TiUPComponentTypeStr) string {
	var component string
	switch tiUPComponent {
	case TiEMComponentTypeStr:
		component = string(TiEMComponentTypeStr)
	default:
		component = string(DefaultComponentTypeStr)
	}
	tiUPConfig, err := models.GetTiUPConfigReaderWriter().QueryByComponentType(context.Background(), component)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("fail get tiup_home for %s: %s", component, err.Error())
		return ""
	} else {
		return tiUPConfig.TiupHome
	}
}

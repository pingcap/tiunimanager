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

package secondparty

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/spec"
	spec2 "github.com/pingcap/tiup/pkg/cluster/spec"
	"gopkg.in/yaml.v2"
	"os/exec"
	"syscall"
	"time"

	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
)

type TiUPComponentTypeStr string

const (
	ClusterComponentTypeStr TiUPComponentTypeStr = "cluster"
	DMComponentTypeStr      TiUPComponentTypeStr = "dm"
	TiEMComponentTypeStr	TiUPComponentTypeStr = "tiem"
)

func (secondMicro *SecondMicro) MicroSrvTiupDeploy(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, version string, configStrYaml string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvtiupdeploy tiupcomponent: %s, instancename: %s, version: %s, configstryaml: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, version, configStrYaml, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var deployReq CmdDeployReq
		deployReq.TiUPComponent = tiupComponent
		deployReq.InstanceName = instanceName
		deployReq.Version = version
		deployReq.ConfigStrYaml = configStrYaml
		deployReq.TimeoutS = timeoutS
		deployReq.Flags = flags
		deployReq.TiupPath = secondMicro.TiupBinPath
		deployReq.TaskID = rsp.Id
		secondMicro.startNewTiupDeployTask(ctx, deployReq.TaskID, &deployReq)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupDeployTask(ctx context.Context, taskID uint64, req *CmdDeployReq) {
	topologyTmpFilePath, err := newTmpFileWithContent("tiem-topology", []byte(req.ConfigStrYaml))
	if err != nil {
		secondMicro.taskStatusCh <- TaskStatusMember{
			TaskID:   taskID,
			Status:   TaskStatusError,
			ErrorStr: fmt.Sprintln(err),
		}
		return
	}
	go func() {
		//defer os.Remove(topologyTmpFilePath)
		var args []string
		args = append(args, string(req.TiUPComponent), "deploy", req.InstanceName, req.Version, topologyTmpFilePath)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupScaleOut(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, configStrYaml string, timeoutS int, flags []string, bizID uint64)(taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvtiupscaleout tiupcomponent: %s, instancename: %s, configstryaml: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, configStrYaml, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_ScaleOut
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", rsp, err)
		return 0, err
	} else {
		var scaleOutReq CmdScaleOutReq
		scaleOutReq.TiUPComponent = tiupComponent
		scaleOutReq.InstanceName = instanceName
		scaleOutReq.ConfigStrYaml = configStrYaml
		scaleOutReq.TimeoutS = timeoutS
		scaleOutReq.Flags = flags
		scaleOutReq.TiupPath = secondMicro.TiupBinPath
		scaleOutReq.TaskID = rsp.Id
		secondMicro.startNewTiupScaleOutTask(ctx, scaleOutReq.TaskID, &scaleOutReq)
		return rsp.Id, nil
	}
	return
}

func (secondMicro *SecondMicro) startNewTiupScaleOutTask(ctx context.Context, taskID uint64, req *CmdScaleOutReq) {
	topologyTmpFilePath, err := newTmpFileWithContent("tiem-topology", []byte(req.ConfigStrYaml))
	if err != nil {
		secondMicro.taskStatusCh <- TaskStatusMember{
			TaskID:   taskID,
			Status:   TaskStatusError,
			ErrorStr: fmt.Sprintln(err),
		}
		return
	}
	go func() {
		//defer os.Remove(topologyTmpFilePath)
		var args []string
		args = append(args, string(req.TiUPComponent), "scale-out", req.InstanceName, topologyTmpFilePath)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupScaleIn(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, nodeId string, timeoutS int, flags []string, bizID uint64) (taskId uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvtiupscalein tiupcomponent: %s, instancename: %s, nodeId: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, nodeId, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_ScaleIn
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", rsp, err)
		return 0, err
	} else {
		var scaleInReq CmdScaleInReq
		scaleInReq.TiUPComponent = tiupComponent
		scaleInReq.InstanceName = instanceName
		scaleInReq.NodeId = nodeId
		scaleInReq.TimeoutS = timeoutS
		scaleInReq.Flags = flags
		scaleInReq.TiupPath = secondMicro.TiupBinPath
		scaleInReq.TaskID = rsp.Id
		secondMicro.startNewTiupScaleInTask(ctx, scaleInReq.TaskID, &scaleInReq)
		return rsp.Id, nil
	}
	return
}

func (secondMicro *SecondMicro) startNewTiupScaleInTask(ctx context.Context, taskID uint64, req *CmdScaleInReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "scale-in", req.InstanceName, "--node", req.NodeId)
		args = append(args, req.Flags...)
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupStart(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvtiupstart tiupComponent: %s, instancename: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var req CmdStartReq
		req.TiUPComponent = tiupComponent
		req.TaskID = rsp.Id
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiupPath = secondMicro.TiupBinPath
		req.Flags = flags
		secondMicro.startNewTiupStartTask(ctx, req.TaskID, &req)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupStartTask(ctx context.Context, taskID uint64, req *CmdStartReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "start", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupRestart(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvtiuprestart tiupcomponent: %s, instancename: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restart
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var req CmdStartReq
		req.TiUPComponent = tiupComponent
		req.TaskID = rsp.Id
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiupPath = secondMicro.TiupBinPath
		req.Flags = flags
		secondMicro.startNewTiupRestartTask(ctx, req.TaskID, &req)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupRestartTask(ctx context.Context, taskID uint64, req *CmdStartReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "restart", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupStop(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvtiupstop tiupComponent: %s, instancename: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Stop
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var req CmdStartReq
		req.TiUPComponent = tiupComponent
		req.TaskID = rsp.Id
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiupPath = secondMicro.TiupBinPath
		req.Flags = flags
		secondMicro.startNewTiupStopTask(ctx, req.TaskID, &req)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupStopTask(ctx context.Context, taskID uint64, req *CmdStartReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "stop", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupList(ctx context.Context, tiupComponent TiUPComponentTypeStr, timeoutS int, flags []string) (resp *CmdListResp, err error) {
	framework.LogWithContext(ctx).Infof("microsrvtiuplist tiupComponent: %s, timeout: %d, flags: %v", string(tiupComponent), timeoutS, flags)
	var req CmdListReq
	req.TiUPComponent = tiupComponent
	req.TimeoutS = timeoutS
	req.TiupPath = secondMicro.TiupBinPath
	req.Flags = flags
	cmdListResp, err := secondMicro.startNewTiupListTask(ctx, &req)
	return &cmdListResp, err
}

func (secondMicro *SecondMicro) startNewTiupListTask(ctx context.Context, req *CmdListReq) (resp CmdListResp, err error) {
	var args []string
	args = append(args, string(req.TiUPComponent), "list")
	args = append(args, req.Flags...)
	args = append(args, "--yes")

	logInFunc := framework.LogWithContext(ctx)
	logInFunc.Info("task start processing:", fmt.Sprintf("tiupPath:%s tiupArgs:%v timeouts:%d", req.TiupPath, args, req.TimeoutS))
	var cmd *exec.Cmd
	var cancelFp context.CancelFunc
	if req.TimeoutS != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.TimeoutS)*time.Second)
		cancelFp = cancel
		cmd = exec.CommandContext(ctx, req.TiupPath, args...)
	} else {
		cmd = exec.Command(req.TiupPath, args...)
		cancelFp = func() {}
	}
	defer cancelFp()
	cmd.SysProcAttr = genSysProcAttr()
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	var data []byte
	if data, err = cmd.Output(); err != nil {
		logInFunc.Errorf("cmd start err: %+v, errStr: %s", err, stderr.String())
		err = fmt.Errorf("cmd start err: %+v, errStr: %s", err, stderr.String())
		return
	}
	resp.ListRespStr = string(data)
	return
}

func (secondMicro *SecondMicro) MicroSrvTiupDestroy(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvtiupstop tiupComponent: %s, instancename: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var req CmdDestroyReq
		req.TiUPComponent = tiupComponent
		req.TaskID = rsp.Id
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiupPath = secondMicro.TiupBinPath
		req.Flags = flags
		secondMicro.startNewTiupDestroyTask(ctx, req.TaskID, &req)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupDestroyTask(ctx context.Context, taskID uint64, req *CmdDestroyReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "destroy", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvDumpling(ctx context.Context, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvdumpling, timeouts: %d, flags: %v, bizid: %d", timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Dumpling
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var dumplingReq CmdDumplingReq
		dumplingReq.TaskID = rsp.Id
		dumplingReq.TimeoutS = timeoutS
		dumplingReq.TiupPath = secondMicro.TiupBinPath
		dumplingReq.Flags = flags
		secondMicro.startNewTiupDumplingTask(ctx, dumplingReq.TaskID, &dumplingReq)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupDumplingTask(ctx context.Context, taskID uint64, req *CmdDumplingReq) {
	go func() {
		var args []string
		args = append(args, "dumpling")
		args = append(args, req.Flags...)
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvLightning(ctx context.Context, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvlightning, timeouts: %d, flags: %v, bizid: %d", timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Lightning
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var lightningReq CmdLightningReq
		lightningReq.TaskID = rsp.Id
		lightningReq.TimeoutS = timeoutS
		lightningReq.TiupPath = secondMicro.TiupBinPath
		lightningReq.Flags = flags
		secondMicro.startNewTiupLightningTask(ctx, lightningReq.TaskID, &lightningReq)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupLightningTask(ctx context.Context, taskID uint64, req *CmdLightningReq) {
	go func() {
		var args []string
		args = append(args, "tidb-lightning")
		args = append(args, req.Flags...)
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupDisplay(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string) (resp *CmdDisplayResp, err error) {
	framework.LogWithContext(ctx).Infof("microsrvtiupclusterdisplay tiupcomponent: %s,  instanceName: %s, timeouts: %d, flags: %v", string(tiupComponent), instanceName, timeoutS, flags)
	var req CmdDisplayReq
	req.TiUPComponent = tiupComponent
	req.InstanceName = instanceName
	req.TimeoutS = timeoutS
	req.TiupPath = secondMicro.TiupBinPath
	req.Flags = flags
	cmdDisplayResp, err := secondMicro.startNewTiupDisplayTask(ctx, &req)
	return &cmdDisplayResp, err
}

func (secondMicro *SecondMicro) startNewTiupDisplayTask(ctx context.Context, req *CmdDisplayReq) (resp CmdDisplayResp, err error) {
	var args []string
	args = append(args, string(req.TiUPComponent), "display")
	args = append(args, req.InstanceName)
	args = append(args, req.Flags...)

	logInFunc := framework.LogWithContext(ctx)
	logInFunc.Info("task start processing:", fmt.Sprintf("tiupPath:%s tiupArgs:%v timeouts:%d", req.TiupPath, args, req.TimeoutS))
	var cmd *exec.Cmd
	var cancelFp context.CancelFunc
	if req.TimeoutS != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.TimeoutS)*time.Second)
		cancelFp = cancel
		cmd = exec.CommandContext(ctx, req.TiupPath, args...)
	} else {
		cmd = exec.Command(req.TiupPath, args...)
		cancelFp = func() {}
	}
	defer cancelFp()
	cmd.SysProcAttr = genSysProcAttr()
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	var data []byte
	if data, err = cmd.Output(); err != nil {
		logInFunc.Errorf("cmd start err: %+v, errStr: %s", err, stderr.String())
		err = fmt.Errorf("cmd start err: %+v, errStr: %s", err, stderr.String())
		return
	}
	resp.DisplayRespString = string(data)
	return
}

func (secondMicro *SecondMicro) MicroSrvTiupTransfer(ctx context.Context, tiupComponent TiUPComponentTypeStr, instanceName string, collectorYaml string, remotePath string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvtiuptransfer tiupcomponent: %s, instancename: %s, collectoryaml: %s, remotepath: %s, timeouts: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, collectorYaml, remotePath, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Transfer
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var req CmdTransferReq
		req.TiUPComponent = tiupComponent
		req.InstanceName = instanceName
		req.CollectorYaml = collectorYaml
		req.RemotePath = remotePath
		req.TimeoutS = timeoutS
		req.Flags = flags
		req.TiupPath = secondMicro.TiupBinPath
		req.TaskID = rsp.Id
		secondMicro.startNewTiupTransferTask(ctx, req.TaskID, &req)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupTransferTask(ctx context.Context, taskID uint64, req *CmdTransferReq) {
	collectorTmpFilePath, err := newTmpFileWithContent("tiem-collector", []byte(req.CollectorYaml))
	if err != nil {
		secondMicro.taskStatusCh <- TaskStatusMember{
			TaskID:   taskID,
			Status:   TaskStatusError,
			ErrorStr: fmt.Sprintln(err),
		}
		return
	}
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "push", req.InstanceName, collectorTmpFilePath, req.RemotePath)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupUpgrade(ctx context.Context, tiupComponent TiUPComponentTypeStr,
	instanceName string, version string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvtiupupgrade tiupcomponent: %s" +
		", instancename: %s, version: %s, timeouts: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName,
		version, timeoutS, flags, bizID)
	req := dbPb.CreateTiupTaskRequest{
		Type : dbPb.TiupTaskType_Upgrade,
		BizID : bizID,
	}
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var req CmdUpgradeReq
		req.TiUPComponent = tiupComponent
		req.InstanceName = instanceName
		req.Version = version
		req.TimeoutS = timeoutS
		req.Flags = flags
		req.TiupPath = secondMicro.TiupBinPath
		req.TaskID = rsp.Id
		secondMicro.startNewTiupUpgradeTask(ctx, req.TaskID, &req)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupUpgradeTask(ctx context.Context, taskID uint64, req *CmdUpgradeReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "upgrade", req.InstanceName, req.Version)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(ctx, taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupShowConfig(ctx context.Context, req *CmdShowConfigReq) (resp *CmdShowConfigResp, err error) {
	framework.LogWithContext(ctx).Infof("microsrvtiupshowconfig cmdshowconfigreq: %v", req)
	var args []string
	args = append(args, string(req.TiUPComponent), "show-config")
	args = append(args, req.InstanceName)
	args = append(args, req.Flags...)

	logInFunc := framework.LogWithContext(ctx)
	logInFunc.Info("task start processing:", fmt.Sprintf("tiupPath:%s tiupArgs:%v timeouts:%d", secondMicro.TiupBinPath, args, req.TimeoutS))
	var cmd *exec.Cmd
	var cancelFp context.CancelFunc
	if req.TimeoutS != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.TimeoutS)*time.Second)
		cancelFp = cancel
		cmd = exec.CommandContext(ctx, secondMicro.TiupBinPath, args...)
	} else {
		cmd = exec.Command(secondMicro.TiupBinPath, args...)
		cancelFp = func() {}
	}
	defer cancelFp()
	cmd.SysProcAttr = genSysProcAttr()
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	var data []byte
	if data, err = cmd.Output(); err != nil {
		logInFunc.Errorf("cmd start err: %+v, errStr: %s", err, stderr.String())
		err = fmt.Errorf("cmd start err: %+v, errStr: %s", err, stderr.String())
		return
	}

	topoStr := string(data)
	topo := spec2.Specification{}
	if err = yaml.UnmarshalStrict([]byte(topoStr), topo); err != nil {
		logInFunc.Errorf("parse original config(%s) error: %+v", topoStr, err)
		return
	}

	resp.TiDBClusterTopo = topo
	return
}

func (secondMicro *SecondMicro) MicroSrvTiupEditGlobalConfig(ctx context.Context,
	cmdEditGlobalConfigReq CmdEditGlobalConfigReq, bizID uint64) (uint64, error) {
	framework.LogWithContext(ctx).Infof("microsrvtiupeditglobalconfig cmdeditglobalconfigreq: %v, bizid: %d", cmdEditGlobalConfigReq, bizID)
	req := dbPb.CreateTiupTaskRequest{
		Type : dbPb.TiupTaskType_EditGlobalConfig,
		BizID : bizID,
	}
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	}

	cmdShowConfigReq := CmdShowConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName: cmdEditGlobalConfigReq.InstanceName,
	}
	cmdShowConfigResp, err := secondMicro.MicroSrvTiupShowConfig(ctx, &cmdShowConfigReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("check orignal config error: %+v", err)
		secondMicro.taskStatusCh <- TaskStatusMember{
			TaskID:   rsp.Id,
			Status:   TaskStatusError,
			ErrorStr: fmt.Sprintln(err),
		}
		return rsp.Id, err
	}
	topo := cmdShowConfigResp.TiDBClusterTopo

	secondMicro.startTiupEditGlobalConfigTask(ctx, rsp.Id, &cmdEditGlobalConfigReq, &topo)
	return rsp.Id, nil
}

func (secondMicro *SecondMicro) startTiupEditGlobalConfigTask(ctx context.Context, taskID uint64, req *CmdEditGlobalConfigReq, topo *spec2.Specification) {
	var componentServerConfigs map[string]interface{}
	switch req.TiDBClusterComponent {
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
	for k, v := range req.ConfigMap {
		componentServerConfigs[k] = v
	}
	switch req.TiDBClusterComponent {
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

	cmdEditConfigReq := CmdEditConfigReq{
		TiUPComponent: req.TiUPComponent,
		InstanceName: req.InstanceName,
		NewTopo: topo,
		TimeoutS: req.TimeoutS,
		Flags: req.Flags,
	}
	secondMicro.startNewTiupEditConfigTask(ctx, cmdEditConfigReq, taskID)
}

func (secondMicro *SecondMicro) MicroSrvTiupEditInstanceConfig(ctx context.Context,
	cmdEditInstanceConfigReq CmdEditInstanceConfigReq, bizID uint64) (uint64, error) {
	framework.LogWithContext(ctx).Infof("microsrvtiupeditinstanceconfig cmdeditinstanceconfigreq: %v, bizid: %d", cmdEditInstanceConfigReq, bizID)
	req := dbPb.CreateTiupTaskRequest{
		Type : dbPb.TiupTaskType_EditInstanceConfig,
		BizID : bizID,
	}
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	}

	cmdShowConfigReq := CmdShowConfigReq{
		TiUPComponent: ClusterComponentTypeStr,
		InstanceName: cmdEditInstanceConfigReq.InstanceName,
	}
	cmdShowConfigResp, err := secondMicro.MicroSrvTiupShowConfig(ctx, &cmdShowConfigReq)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("check orignal config error: %+v", err)
		secondMicro.taskStatusCh <- TaskStatusMember{
			TaskID:   rsp.Id,
			Status:   TaskStatusError,
			ErrorStr: fmt.Sprintln(err),
		}
		return rsp.Id, err
	}
	topo := cmdShowConfigResp.TiDBClusterTopo

	secondMicro.startTiupEditInstanceConfigTask(ctx, rsp.Id, &cmdEditInstanceConfigReq, &topo)
	return rsp.Id, nil
}

func (secondMicro *SecondMicro) startTiupEditInstanceConfigTask(ctx context.Context, taskID uint64, req *CmdEditInstanceConfigReq, topo *spec2.Specification) {
	switch req.TiDBClusterComponent {
	case spec.TiDBClusterComponent_TiDB:
		for idx, tiDBServer := range topo.TiDBServers {
			if tiDBServer.Host == req.Host && tiDBServer.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(tiDBServer, FieldKey_Yaml, k, v)
				}
				topo.TiDBServers[idx] = tiDBServer
				break
			}
		}
	case spec.TiDBClusterComponent_TiKV:
		for idx, tiKVServer := range topo.TiKVServers {
			if tiKVServer.Host == req.Host && tiKVServer.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(tiKVServer, FieldKey_Yaml, k, v)
				}
				topo.TiKVServers[idx] = tiKVServer
				break
			}
		}
	case spec.TiDBClusterComponent_TiFlash:
		for idx, tiFlashServer := range topo.TiFlashServers {
			if tiFlashServer.Host == req.Host && tiFlashServer.FlashServicePort == req.Port {
				for k, v := range req.ConfigMap {
					SetField(tiFlashServer, FieldKey_Yaml, k, v)
				}
				topo.TiFlashServers[idx] = tiFlashServer
				break
			}
		}
	case spec.TiDBClusterComponent_PD:
		for idx, pdServer := range topo.PDServers {
			if pdServer.Host == req.Host && pdServer.ClientPort == req.Port {
				for k, v := range req.ConfigMap {
					SetField(pdServer, FieldKey_Yaml, k, v)
				}
				topo.PDServers[idx] = pdServer
				break
			}
		}
	case spec.TiDBClusterComponent_Pump:
		for idx, pumpServer := range topo.PumpServers {
			if pumpServer.Host == req.Host && pumpServer.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(pumpServer, FieldKey_Yaml, k, v)
				}
				topo.PumpServers[idx] = pumpServer
				break
			}
		}
	case spec.TiDBClusterComponent_Drainer:
		for idx, drainer := range topo.Drainers {
			if drainer.Host == req.Host && drainer.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(drainer, FieldKey_Yaml, k, v)
				}
				topo.Drainers[idx] = drainer
				break
			}
		}
	case spec.TiDBClusterComponent_CDC:
		for idx, cdcServer := range topo.CDCServers {
			if cdcServer.Host == req.Host && cdcServer.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(cdcServer, FieldKey_Yaml, k, v)
				}
				topo.CDCServers[idx] = cdcServer
				break
			}
		}
	case spec.TiDBClusterComponent_TiSparkMasters:
		for idx, tiSparkMaster := range topo.TiSparkMasters {
			if tiSparkMaster.Host == req.Host && tiSparkMaster.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(tiSparkMaster, FieldKey_Yaml, k, v)
				}
				topo.TiSparkMasters[idx] = tiSparkMaster
				break
			}
		}
	case spec.TiDBClusterComponent_TiSparkWorkers:
		for idx, tiSparkWorker := range topo.TiSparkWorkers {
			if tiSparkWorker.Host == req.Host && tiSparkWorker.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(tiSparkWorker, FieldKey_Yaml, k, v)
				}
				topo.TiSparkWorkers[idx] = tiSparkWorker
				break
			}
		}
	case spec.TiDBClusterComponent_Prometheus:
		for idx, monitor := range topo.Monitors {
			if monitor.Host == req.Host && monitor.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(monitor, FieldKey_Yaml, k, v)
				}
				topo.Monitors[idx] = monitor
				break
			}
		}
	case spec.TiDBClusterComponent_Grafana:
		for idx, grafana := range topo.Grafanas {
			if grafana.Host == req.Host && grafana.Port == req.Port {
				for k, v := range req.ConfigMap {
					SetField(grafana, FieldKey_Yaml, k, v)
				}
				topo.Grafanas[idx] = grafana
				break
			}
		}
	case spec.TiDBClusterComponent_Alertmanager:
		for idx, alertManager := range topo.Alertmanagers {
			if alertManager.Host == req.Host && alertManager.WebPort == req.Port {
				for k, v := range req.ConfigMap {
					SetField(alertManager, FieldKey_Yaml, k, v)
				}
				topo.Alertmanagers[idx] = alertManager
				break
			}
		}
	}

	cmdEditConfigReq := CmdEditConfigReq{
		TiUPComponent: req.TiUPComponent,
		InstanceName: req.InstanceName,
		NewTopo: topo,
		TimeoutS: req.TimeoutS,
		Flags: req.Flags,
	}
	secondMicro.startNewTiupEditConfigTask(ctx, cmdEditConfigReq, taskID)
}

func (secondMicro *SecondMicro) startNewTiupEditConfigTask(ctx context.Context, req CmdEditConfigReq,
	taskID uint64) {
	newData, err := yaml.Marshal(req.NewTopo)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("startnewtiupeditconfigtask marshal new config(%+v) error: %+v", req.NewTopo, err)
		secondMicro.taskStatusCh <- TaskStatusMember{
			TaskID:   taskID,
			Status:   TaskStatusError,
			ErrorStr: fmt.Sprintf("startnewtiupeditconfigtask marshal new config(%+v) error: %+v", req.NewTopo, err),
		}
		return
	}

	topologyTmpFilePath, err := newTmpFileWithContent("tidb-cluster-topology", newData)
	if err != nil {
		secondMicro.taskStatusCh <- TaskStatusMember{
			TaskID:   taskID,
			Status:   TaskStatusError,
			ErrorStr: fmt.Sprintln(err),
		}
		return
	}
	go func() {
		//defer os.Remove(topologyTmpFilePath)
		var args []string
		args = append(args, string(req.TiUPComponent), "edit-config", req.InstanceName, "-f", topologyTmpFilePath)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(ctx, taskID, secondMicro.TiupBinPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupReload(ctx context.Context, cmdReloadConfigReq CmdReloadConfigReq,
	bizID uint64) (taskID uint64, err error) {
	framework.LogWithContext(ctx).Infof("microsrvtiupreload cmdreloadconfigreq: %v, bizid: %d", cmdReloadConfigReq, bizID)
	req := dbPb.CreateTiupTaskRequest{
		Type : dbPb.TiupTaskType_Reload,
		BizID : bizID,
	}
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	}
	secondMicro.startNewTiupReloadTask(ctx, rsp.Id, &cmdReloadConfigReq)
	return rsp.Id, nil
}

func (secondMicro *SecondMicro) startNewTiupReloadTask(ctx context.Context, taskID uint64, req *CmdReloadConfigReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "reload", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(ctx, taskID, secondMicro.TiupBinPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) startNewTiupTask(ctx context.Context, taskID uint64, tiupPath string, tiupArgs []string, TimeoutS int) (exitCh chan struct{}) {
	exitCh = make(chan struct{})
	logInFunc := framework.LogWithContext(ctx).WithField("task", taskID)
	logInFunc.Info("task start processing:", fmt.Sprintf("tiupPath:%s tiupArgs:%v timeouts:%d", tiupPath, tiupArgs, TimeoutS))
	secondMicro.taskStatusCh <- TaskStatusMember{
		TaskID:   taskID,
		Status:   TaskStatusProcessing,
		ErrorStr: "",
	}
	go func() {
		defer close(exitCh)
		var cmd *exec.Cmd
		var cancelFp context.CancelFunc
		if TimeoutS != 0 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(TimeoutS)*time.Second)
			cancelFp = cancel
			cmd = exec.CommandContext(ctx, tiupPath, tiupArgs...)
		} else {
			cmd = exec.Command(tiupPath, tiupArgs...)
			cancelFp = func() {}
		}
		defer cancelFp()
		cmd.SysProcAttr = genSysProcAttr()
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		t0 := time.Now()
		if err := cmd.Start(); err != nil {
			logInFunc.Errorf("cmd start err: %+v, errStr: %s", err, stderr.String())
			secondMicro.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintf("cmd start err: %+v, errStr: %s", err, stderr.String()),
			}
			return
		}
		logInFunc.Info("cmd started")
		successFp := func() {
			logInFunc.Info("task finished, time cost", time.Since(t0))
			secondMicro.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusFinished,
				ErrorStr: "",
			}
		}
		logInFunc.Info("cmd wait")
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
			logInFunc.Errorf("cmd wait return with err: %+v, errStr: %s, time cost: %v", err, stderr.String(), time.Since(t0))
			secondMicro.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintf("cmd wait return with err: %+v, errStr: %s, time cost: %v", err, stderr.String(), time.Since(t0)),
			}
			return
		} else {
			logInFunc.Info("cmd wait return successfully")
			successFp()
			return
		}
	}()
	return exitCh
}

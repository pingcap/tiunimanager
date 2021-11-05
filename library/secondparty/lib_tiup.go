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
	"context"
	"fmt"
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
)

func (secondMicro *SecondMicro) MicroSrvTiupDeploy(tiupComponent TiUPComponentTypeStr, instanceName string, version string, configStrYaml string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	logger.Infof("microsrvtiupdeploy tiupcomponent: %s, instancename: %s, version: %s, configstryaml: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, version, configStrYaml, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
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
		secondMicro.startNewTiupDeployTask(deployReq.TaskID, &deployReq)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupDeployTask(taskID uint64, req *CmdDeployReq) {
	topologyTmpFilePath, err := newTmpFileWithContent([]byte(req.ConfigStrYaml))
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
		<-secondMicro.startNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupStart(tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	logger.Infof("microsrvtiupstart tiupComponent: %s, instancename: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var req CmdStartReq
		req.TiUPComponent = tiupComponent
		req.TaskID = rsp.Id
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiupPath = secondMicro.TiupBinPath
		req.Flags = flags
		secondMicro.startNewTiupStartTask(req.TaskID, &req)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupStartTask(taskID uint64, req *CmdStartReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "start", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupRestart(tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	logger.Infof("microsrvtiuprestart tiupcomponent: %s, instancename: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restart
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var req CmdStartReq
		req.TiUPComponent = tiupComponent
		req.TaskID = rsp.Id
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiupPath = secondMicro.TiupBinPath
		req.Flags = flags
		secondMicro.startNewTiupRestartTask(req.TaskID, &req)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupRestartTask(taskID uint64, req *CmdStartReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "restart", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupStop(tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	logger.Infof("microsrvtiupstop tiupComponent: %s, instancename: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Stop
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var req CmdStartReq
		req.TiUPComponent = tiupComponent
		req.TaskID = rsp.Id
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiupPath = secondMicro.TiupBinPath
		req.Flags = flags
		secondMicro.startNewTiupStopTask(req.TaskID, &req)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupStopTask(taskID uint64, req *CmdStartReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "stop", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupList(tiupComponent TiUPComponentTypeStr, timeoutS int, flags []string) (resp *CmdListResp, err error) {
	logger.Infof("microsrvtiuplist tiupComponent: %s, timeout: %d, flags: %v", string(tiupComponent), timeoutS, flags)
	var req CmdListReq
	req.TiUPComponent = tiupComponent
	req.TimeoutS = timeoutS
	req.TiupPath = secondMicro.TiupBinPath
	req.Flags = flags
	cmdListResp, err := secondMicro.startNewTiupListTask(&req)
	return &cmdListResp, err
}

func (secondMicro *SecondMicro) startNewTiupListTask(req *CmdListReq) (resp CmdListResp, err error) {
	var args []string
	args = append(args, string(req.TiUPComponent), "list")
	args = append(args, req.Flags...)
	args = append(args, "--yes")

	logger.Info("task start processing:", fmt.Sprintf("tiupPath:%s tiupArgs:%v timeouts:%d", req.TiupPath, args, req.TimeoutS))
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
	var data []byte
	if data, err = cmd.Output(); err != nil {
		logger.Error("cmd start err", err)
		return
	}
	resp.ListRespStr = string(data)
	return
}

func (secondMicro *SecondMicro) MicroSrvTiupDestroy(tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	logger.Infof("microsrvtiupstop tiupComponent: %s, instancename: %s, timeout: %d, flags: %v, bizid: %d", string(tiupComponent), instanceName, timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var req CmdDestroyReq
		req.TiUPComponent = tiupComponent
		req.TaskID = rsp.Id
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiupPath = secondMicro.TiupBinPath
		req.Flags = flags
		secondMicro.startNewTiupDestroyTask(req.TaskID, &req)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupDestroyTask(taskID uint64, req *CmdDestroyReq) {
	go func() {
		var args []string
		args = append(args, string(req.TiUPComponent), "destroy", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvDumpling(timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	logger.Infof("microsrvdumpling, timeouts: %d, flags: %v, bizid: %d", timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Dumpling
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var dumplingReq CmdDumplingReq
		dumplingReq.TaskID = rsp.Id
		dumplingReq.TimeoutS = timeoutS
		dumplingReq.TiupPath = secondMicro.TiupBinPath
		dumplingReq.Flags = flags
		secondMicro.startNewTiupDumplingTask(dumplingReq.TaskID, &dumplingReq)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupDumplingTask(taskID uint64, req *CmdDumplingReq) {
	go func() {
		var args []string
		args = append(args, "dumpling")
		args = append(args, req.Flags...)
		<-secondMicro.startNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvLightning(timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	logger.Infof("microsrvlightning, timeouts: %d, flags: %v, bizid: %d", timeoutS, flags, bizID)
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Lightning
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var lightningReq CmdLightningReq
		lightningReq.TaskID = rsp.Id
		lightningReq.TimeoutS = timeoutS
		lightningReq.TiupPath = secondMicro.TiupBinPath
		lightningReq.Flags = flags
		secondMicro.startNewTiupLightningTask(lightningReq.TaskID, &lightningReq)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewTiupLightningTask(taskID uint64, req *CmdLightningReq) {
	go func() {
		var args []string
		args = append(args, "tidb-lightning")
		args = append(args, req.Flags...)
		<-secondMicro.startNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupDisplay(tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string) (resp *CmdDisplayResp, err error) {
	logger.Infof("microsrvtiupclusterdisplay tiupcomponent: %s,  instanceName: %s, timeouts: %d, flags: %v", string(tiupComponent), instanceName, timeoutS, flags)
	var req CmdDisplayReq
	req.TiUPComponent = tiupComponent
	req.InstanceName = instanceName
	req.TimeoutS = timeoutS
	req.TiupPath = secondMicro.TiupBinPath
	req.Flags = flags
	cmdDisplayResp, err := secondMicro.startNewTiupDisplayTask(&req)
	return &cmdDisplayResp, err
}

func (secondMicro *SecondMicro) startNewTiupDisplayTask(req *CmdDisplayReq) (resp CmdDisplayResp, err error) {
	var args []string
	args = append(args, string(req.TiUPComponent), "display")
	args = append(args, req.InstanceName)
	args = append(args, req.Flags...)

	logger.Info("task start processing:", fmt.Sprintf("tiupPath:%s tiupArgs:%v timeouts:%d", req.TiupPath, args, req.TimeoutS))
	//fmt.Println("task start processing:", fmt.Sprintf("tiupPath:%s tiupArgs:%v timeouts:%d", tiupPath, tiupArgs, TimeoutS))
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
	var data []byte
	if data, err = cmd.Output(); err != nil {
		logger.Error("cmd start err", err)
		return
	}
	resp.DisplayRespString = string(data)
	return
}

func (secondMicro *SecondMicro) startNewTiupTask(taskID uint64, tiupPath string, tiupArgs []string, TimeoutS int) (exitCh chan struct{}) {
	exitCh = make(chan struct{})
	logInFunc := logger.WithField("task", taskID)
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
		t0 := time.Now()
		if err := cmd.Start(); err != nil {
			logInFunc.Error("cmd start err", err)
			secondMicro.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintf("%+v, errStr: %s", err, err.Error()),
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
			logInFunc.Error("cmd wait return with err", err)
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					if status.ExitStatus() == 0 {
						successFp()
						return
					}
				}
			}
			logInFunc.Error("task err:", err, "time cost", time.Since(t0))
			secondMicro.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintf("%+v, errStr: %s", err, err.Error()),
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

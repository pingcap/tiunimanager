package secondparty

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
)

func (secondMicro *SecondMicro) MicroSrvTiupDeploy(instanceName string, version string, configStrYaml string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var deployReq CmdDeployReq
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
		defer os.Remove(topologyTmpFilePath)
		var args []string
		args = append(args, "cluster", "deploy", req.InstanceName, req.Version, topologyTmpFilePath)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupStart(instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var req CmdStartReq
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
		args = append(args, "cluster", "start", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupList(timeoutS int, flags []string) (resp *CmdListResp, err error) {
	var req CmdListReq
	req.TimeoutS = timeoutS
	req.TiupPath = secondMicro.TiupBinPath
	req.Flags = flags
	cmdListResp, err := secondMicro.startNewTiupListTask(&req)
	return &cmdListResp, err
}

func (secondMicro *SecondMicro) startNewTiupListTask(req *CmdListReq) (resp CmdListResp, err error) {
	var args []string
	args = append(args, "cluster", "list")
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

func (secondMicro *SecondMicro) MicroSrvTiupDestroy(instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var req CmdDestroyReq
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
		args = append(args, "cluster", "destroy", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-secondMicro.startNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func (secondMicro *SecondMicro) MicroSrvTiupDumpling(timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
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

func (secondMicro *SecondMicro) MicroSrvTiupLightning(timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
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

func (secondMicro *SecondMicro) MicroSrvTiupClusterDisplay(clusterName string, timeoutS int, flags []string) (resp *CmdClusterDisplayResp, err error) {
	var clusterDisplayReq CmdClusterDisplayReq
	clusterDisplayReq.ClusterName = clusterName
	clusterDisplayReq.TimeoutS = timeoutS
	clusterDisplayReq.TiupPath = secondMicro.TiupBinPath
	clusterDisplayReq.Flags = flags
	cmdClusterDisplayResp, err := secondMicro.startNewTiupClusterDisplayTask(&clusterDisplayReq)
	return &cmdClusterDisplayResp, err
}

func (secondMicro *SecondMicro) startNewTiupClusterDisplayTask(req *CmdClusterDisplayReq) (resp CmdClusterDisplayResp, err error) {
	var args []string
	args = append(args, "cluster", "display")
	args = append(args, req.ClusterName)
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
				ErrorStr: fmt.Sprintln(err),
			}
			return
		}
		logInFunc.Info("cmd started")
		successFp := func() {
			logInFunc.Info("task finished, time cost", time.Now().Sub(t0))
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
			logInFunc.Error("task err:", err, "time cost", time.Now().Sub(t0))
			secondMicro.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintln(err),
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
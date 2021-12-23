/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
 ******************************************************************************/

/**
 * @Author: guobob
 * @Description:
 * @File:  dbreplay.go
 * @Version: 1.0.0
 * @Date: 2021/12/14 15:29
 */

package dbreplay

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/common/cmd"
	"github.com/pingcap-inc/tiem/common/structs"
	wsecondparty "github.com/pingcap-inc/tiem/models/workflow/secondparty"
	//dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/errors"
	"strconv"
	"strings"

	"github.com/pingcap-inc/tiem/library/framework"
	"time"
)

type TrafficReplayRun struct {
	Deploy *TrafficReplayDeploy
	Dns    []string
	//uint second
	DBReplayRuntime int64
	TcpDumpStartTS  time.Time
	DBReplicateTS   time.Time
	ProcessID       [][]int64
}

//NewTrafficReplayRun : new traffic replay run
func NewTrafficReplayRun(task *cluster.TrafficReplayTask) *TrafficReplayRun {
	return &TrafficReplayRun{
		Deploy: &TrafficReplayDeploy{
			Basic: &TrafficReplayBasic{
				ClusterIDs: []string{task.ProductClusterID, task.SimulationClusterID},
				RootPath:   make(map[string]string),
				ClusterAggregations : make([]*domain.ClusterAggregation,0),
				Hosts: make(map[string]*structs.HostInfo),
				HostKyePairs: make([][]*HostKey,2,2),
				Status:     "INIT",
			},
			//TODO User specified directory
			//need to specify the directory I need to run the program?
			//Also confirm if each server needs to be specified?
			//DeployPath: []string{task.DeployPath},
			//TcpdumpTempPath: []string{task.TcpdumpTempPath},
			//PcapTempPath: []string{task.PcapTempPath},
			//PcapFileStorePath: []string{task.PcapFileStorePath},
			//ResultFileStorePath: []string{task.ResultFileStorePath},
		},
		Dns : make([]string,0),
		DBReplayRuntime: task.RunTime,
		ProcessID: make([][]int64,0),
	}
}

//GenerateDns : Generate simulation environment tidb server connection  dns
//demo : root:******@tcp(172.16.5.189:4002)/test
func (r *TrafficReplayRun) GenerateDns(ctx context.Context) {

	framework.LogWithContext(ctx).Infof("begin generate simulation environment tidb server connection  dns ")
	defer framework.LogWithContext(ctx).Infof("end generate simulation environment tidb server connection  dns ")

	simulationClusterAggregations := r.Deploy.Basic.ClusterAggregations[1]
	cs := simulationClusterAggregations.Cluster
	config := simulationClusterAggregations.CurrentTopologyConfigRecord.ConfigModel.TiDBServers

	hostKeys:=r.Deploy.Basic.HostKyePairs[1]

	//Current version DNS default link test database
	for k,v := range  hostKeys{
		for _,c := range config{
			if v.Host == c.ListenHost && v.Port==c.Port{
				r.Dns[k]=fmt.Sprintf("root:%v@tcp(%v:%v)/test",
					cs.DbPassword, v.Host, v.Port)
			}
		}
	}

	return
}


// WaitTiUPTaskEnd : Wait for the TiUP task to finish running,
//if  wait more than 300s, the task is considered to have timed out.
func (r *TrafficReplayRun)WaitTiUPTaskEnd(ctx context.Context,taskID string) error{

	var taskFinish  = false
	for i := 0; i < 30; i++ {
		time.Sleep(10 * time.Second)
		//rsp, err := client.DBClient.FindTiupTaskByID(ctx, &req)
		rsp,err := secondparty.Manager.GetOperationStatus(ctx,taskID)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get deploy task err = %s", err.Error())
			return err
		}
		if rsp.Status == wsecondparty.OperationStatus_Error{
			framework.LogWithContext(ctx).Errorf("operate cluster error, %s", rsp.ErrorStr)
			return errors.New(rsp.ErrorStr)
		}
		if rsp.Status == wsecondparty.OperationStatus_Finished {
			taskFinish =true
			break
		}
	}
	if !taskFinish{
		return errors.New("wait TiUP task finished timeout ,wait time 300 second")
	}
	return nil
}

//StopTidbServer : Call the TiUP interface to stop the specified tidb server
func (r *TrafficReplayRun) StopTidbServer(ctx context.Context,  h *HostKey,clusterName string  ) error {

	framework.LogWithContext(ctx).Infof("begin stop tidbserver %v ", h.Key)
	defer framework.LogWithContext(ctx).Infof("end stop tidbserver %v ", h.Key)



	flag := []string{"-y " , " -N ", h.Key}


	//Stopping the tidb server via TiUP
	stopTiDBTaskID,err := secondparty.Manager.ClusterStop(
		ctx, secondparty.ClusterComponentTypeStr, clusterName, 0, flag, r.Deploy.Basic.ID)

	/*taskID, _ := strconv.Atoi(r.Deploy.Basic.ID)
	stopTiDBTaskID, err := secondparty.SecondParty.MicroSrvTiupStop(
		ctx, secondparty.ClusterComponentTypeStr, clusterName, 0, flag, uint64(taskID))
	 */
	if err != nil {
		framework.LogWithContext(ctx).Errorf("call tiup api destroy cluster err = %s", err.Error())
		return err
	}
	defer framework.LogWithContext(ctx).Infof("got stopTiDBTaskId %v", stopTiDBTaskID)

	//wait stop tidb server task end
	err = r.WaitTiUPTaskEnd(ctx,stopTiDBTaskID)
	if err !=nil{
		framework.LogWithContext(ctx).Errorf("stop tidb server fail, %v",err)
		return err
	}

	framework.LogWithContext(ctx).Infof("stop tidb server %v sucess ",h.Key)
	return nil
}

//GenerateTcpdumpCommand : generate tcpdump command
//demo : tcpdump  -w data.%s.pcap -G 300 -Z $USER -z $(pwd)/rotate-hook.sh  host 172.16.4.155 and  tcp   port 4001
func (r *TrafficReplayRun) GenerateTcpdumpCommand(ctx context.Context, h *HostKey ,pos int) string {
	var command string

	framework.LogWithContext(ctx).Infof("before generate tcpdump command ")
	defer framework.LogWithContext(ctx).Infof("after generate tcpdump command %s ", command)

	tcpdumpTempDir := r.Deploy.TcpdumpTempPath[pos]

	command = "tcpdump -w data.%s.pcap -G 300 -Z $USER -z" +
		fmt.Sprintf("%v/rotate-hook.sh  host %v and  tcp   port %v &",
			tcpdumpTempDir, h.Host, h.Port)
	return command
}

//GenerateDBReplayCommand : Generate the start db-replay command
func (r *TrafficReplayRun) GenerateDBReplayCommand(ctx context.Context, pos int) string {
	framework.LogWithContext(ctx).Infof("begin generate DBReplay command ")
	defer framework.LogWithContext(ctx).Infof("end generate DBReplay command ")
	deploy := r.Deploy
	//db-replay dir replay -D "/home/replay/" -d"root:glb34007@tcp(172.16.6.191:4050)/test"  -o "/home/output/"
	//--log-output="/home/log/db-replay.log" --filesize 100 -t 60 &
	command := fmt.Sprintf("%s/db-replay dir replay -D %s -d %s  -o %s --log-output=%s --filesize 100 -t %v &",
		deploy.DeployPath[pos], deploy.PcapFileStorePath[pos], r.Dns[pos], deploy.ResultFileStorePath[pos], deploy.DeployPath[pos],
		r.DBReplayRuntime)

	return command
}

//ExecCommandOnRemoteServer : Commands for starting tcpdump and db-replay
func (r *TrafficReplayRun) ExecCommandOnRemoteServer(ctx context.Context, host *structs.HostInfo, command string) (string, error) {
	framework.LogWithContext(ctx).Infof("begin start command on server [%v]", host.IP)
	defer framework.LogWithContext(ctx).Infof("end start command on server [%v]", host.IP)

	config := cmd.NewSessionConfig(host.UserName, host.Passwd,
		host.IP, 22, 0)

	remoteSess, err := cmd.NewRemoteSession(config)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("connect remote server fail ,%v", err)
		return "", err
	}

	err = remoteSess.Exec(command, cmd.NoWaitCmdEnd)

	output := remoteSess.Output()

	errClose := remoteSess.Close()
	if errClose != nil {
		framework.LogWithContext(ctx).Errorf("close ssh connect  fail , host [%v],error message [%v]",
			host.IP, errClose)
	}

	if err != nil {
		framework.LogWithContext(ctx).Errorf("exec command on remote server [%v] fail , [%v]", host.IP, err)
		return "", err
	}

	//example: [1] 27866
	pids := strings.Split(output, " ")
	if len(pids) < 2 {
		return "0", nil
	}
	return pids[1], nil
}

//SetTcpDumpStartTS : Collect the tcpdump start time of all tidb server nodes and
//take the maximum value as the db-repay replay start time stamp
func (r *TrafficReplayRun) SetTcpDumpStartTS(ts time.Time) {
	if r.TcpDumpStartTS.Before(ts) {
		r.TcpDumpStartTS = ts
	}
}

//SetDBReplicateTS : Current version sync target time set to 10s after tcpdump maximum time
func (r *TrafficReplayRun) SetDBReplicateTS() {
	r.DBReplicateTS = r.TcpDumpStartTS.Add(10 * time.Second)
}


func (r *TrafficReplayRun) SetCDCReplicateTargetTS(ctx context.Context) error {
	//TODO set CDC replicate target ts

	return nil
}


func (r *TrafficReplayRun) WaitReplicateEnd(ctx context.Context) error {
	//TODO wait replicate end

	//failed query retries time
	var retryTimes = 2
	fmt.Println(retryTimes)
	return nil
}

func (r *TrafficReplayRun) AddProcessIDCompare(ctx context.Context, tcpdumpPid, dbreplayPid string) error {
	framework.LogWithContext(ctx).Infof("begin save process %v-%v ", tcpdumpPid, dbreplayPid)
	defer framework.LogWithContext(ctx).Infof("end save process %v-%v ", tcpdumpPid, dbreplayPid)
	tpID, err := strconv.ParseInt(tcpdumpPid, 10, 64)
	if err != nil {
		return err
	}
	dbpID, err := strconv.ParseInt(dbreplayPid, 10, 64)
	if err != nil {
		return err
	}
	pid := []int64{tpID, dbpID}
	r.ProcessID = append(r.ProcessID, pid)
	return nil
}

func (r *TrafficReplayRun) CheckReachRunTime(ctx context.Context) bool {
	framework.LogWithContext(ctx).Infof("begin check that the task has reached the set running time ,%v-%v-%v",
		r.DBReplicateTS, time.Now().String(), r.DBReplayRuntime)
	defer framework.LogWithContext(ctx).Infof("begin check that the task has reached the set running time ,%v-%v-%v",
		r.DBReplicateTS, time.Now().String(), r.DBReplayRuntime)
	if int64(time.Since(r.DBReplicateTS).Seconds())> r.DBReplayRuntime  {
		return true
	}
	return false
}

func (r *TrafficReplayRun) CheckTaskStatus(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin check task status ")
	defer framework.LogWithContext(ctx).Infof("begin check task status ")
	//TODO
	return nil
}

func (r *TrafficReplayRun) AnalysisDBReplayOutput(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin analysis db-replay output")
	defer framework.LogWithContext(ctx).Infof("end analysis db-replay output")
	//TODO
	return nil
}

//TaskTrack : Track the status of task execution at regular intervals and analyse the output
func (r *TrafficReplayRun) TaskTrack(ctx context.Context) error {
	timer := time.NewTicker(5 * time.Minute)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			//TODO: Check that the running time has been reached

			if r.CheckReachRunTime(ctx) {
				framework.LogWithContext(ctx).Infof("Task run time reaches specified time %v-%v",
					r.DBReplayRuntime,r.DBReplicateTS)
				return nil
			}

			//TODO: Check tcpdump and db-replay run status
			err := r.CheckTaskStatus(ctx)
			if err != nil {
				return err
			}

			//TODO: Analyse the db-replay output file
			err = r.AnalysisDBReplayOutput(ctx)
			if err != nil {
				return err
			}
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
	//return nil
}

func (r *TrafficReplayRun) StopProcessByPID(ctx context.Context, host *structs.HostInfo, pID int64) error {
	//If pid is 0, the process is not considered to have started,
	//so there is no need to stop it and returns success directly
	if pID == 0 {
		return nil
	}
	framework.LogWithContext(ctx).Infof("begin stop process [%v] on server [%v]", pID, host.IP)
	defer framework.LogWithContext(ctx).Infof("sotp process [%v] on server [%v] end ", pID, host.IP)

	config := cmd.NewSessionConfig(host.UserName, host.Passwd,
		host.IP, 22, 0)

	remoteSess, err := cmd.NewRemoteSession(config)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("connect remote server fail ,%v", err)
		return err
	}
	command := fmt.Sprintf("kill -9 %v", pID)
	err = remoteSess.Exec(command, cmd.WaitCmdEnd)
	errClose := remoteSess.Close()
	if errClose != nil {
		framework.LogWithContext(ctx).Errorf("close connect session on  server [%s]", host.IP)
	}

	if err != nil {
		framework.LogWithContext(ctx).Errorf("exec command [%s] on server [%s] fail [%v]", command, host.IP, err)
		return err
	}

	return nil
}

func (r *TrafficReplayRun) StopPrograms(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin stop program")
	defer framework.LogWithContext(ctx).Info("end stop program")
	productHostKeys := r.Deploy.Basic.HostKyePairs[0]
	//simulationHost := r.Deploy.Basic.Hosts[1]

	//var err error
	for k := range productHostKeys {
		res ,err := r.Deploy.Basic.GetHostPair(ctx,k,AllHostInfo)
		if err !=nil{
			return err
		}
		//There may be processes in some nodes that have not yet started,
		//so the slice storing the process id may be shorter than
		//the slice with the coarse host information
		if k < len(r.ProcessID) {
			//stop tcpdump
			err = r.StopProcessByPID(ctx, res[0], r.ProcessID[k][0])
			if err != nil {
				framework.LogWithContext(ctx).Errorf("stop tcpdump %v on server %s fail ,%v ",
					r.ProcessID[k][0], res[0].IP, err)
			}
			err = r.StopProcessByPID(ctx, res[1], r.ProcessID[k][1])
			if err != nil {
				framework.LogWithContext(ctx).Errorf("stop db-replay %v on server %s fail ,%v ",
					r.ProcessID[k][1], res[1].IP, err)
			}
		}
	}
	return nil
}

func (r *TrafficReplayRun) ClearRemoteDir(ctx context.Context, host *structs.HostInfo, dir string) error {
	framework.LogWithContext(ctx).Infof("begin clear dir [%v] on remote server [%v]", dir, host.IP)
	defer framework.LogWithContext(ctx).Infof("end clear dir [%v] on remote server [%v]", dir, host.IP)
	config := cmd.NewSessionConfig(host.UserName, host.Passwd,
		host.IP, 22, 0)

	remoteSess, err := cmd.NewRemoteSession(config)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("connect remote server fail ,%v", err)
		return err
	}
	command := fmt.Sprintf("rm -rf %v", dir)
	err = remoteSess.Exec(command, cmd.WaitCmdEnd)
	errClose := remoteSess.Close()
	if errClose != nil {
		framework.LogWithContext(ctx).Errorf("close ssh connect session on remote server [%v] fail,[%v]",
			host.IP, err)
	}
	if err != nil {
		framework.LogWithContext(ctx).Errorf("exec command [%v] on remote server [%v] fail,[%v]",
			command, host.IP, err)
	}
	return err
}

func (r *TrafficReplayRun) ClearData(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin clear data for task %v ", r.Deploy.Basic.ID)
	defer framework.LogWithContext(ctx).Infof("end clear data for task %v", r.Deploy.Basic.ID)
	productHostKeys := r.Deploy.Basic.HostKyePairs[0]

	var err error
	for k  := range productHostKeys {
		//TODO Failed to remove directory, task continues
		//clear tcpdump temp path
		res ,err := r.Deploy.Basic.GetHostPair(ctx,k,AllHostInfo)
		if err !=nil{
			return nil
		}
		err = r.ClearRemoteDir(ctx, res[0], r.Deploy.TcpdumpTempPath[k])
	/*
		//clear pcap temp path
		err = r.ClearRemoteDir(ctx, res[1], r.Deploy.PcapTempPath[k])
	*/
		//clear pcap store path
		err = r.ClearRemoteDir(ctx, res[1], r.Deploy.PcapFileStorePath[k])
		//clear result file store path
		err = r.ClearRemoteDir(ctx, res[1], r.Deploy.ResultFileStorePath[k])
		//clear deploy path
		err = r.ClearRemoteDir(ctx, res[1], r.Deploy.DeployPath[k])
	}
	return err
}

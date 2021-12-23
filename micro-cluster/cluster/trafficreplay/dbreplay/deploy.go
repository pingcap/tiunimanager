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
 * @File:  deploy.go
 * @Version: 1.0.0
 * @Date: 2021/12/13 15:28
 */

package dbreplay

import (
	"context"
	"fmt"
	"github.com/go-basic/uuid"
	"github.com/pingcap-inc/tiem/common/cmd"
	"github.com/pingcap-inc/tiem/common/mutualtrust"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"strings"
)

type TrafficReplayDeploy struct {
	Basic               *TrafficReplayBasic
	TcpdumpExist        []bool
	DeployPath          []string
	TcpdumpTempPath     []string
	//PcapTempPath        []string
	PcapFileStorePath   []string
	ResultFileStorePath []string
}

//CreateMutualTrust : Creation of mutual trust between the tidb server servers
//of the production and simulation environments
func (d *TrafficReplayDeploy) CreateMutualTrust(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin create mutual trust")
	defer framework.LogWithContext(ctx).Infof("end create mutual trust")
	productHostKey := d.Basic.HostKyePairs[0]

	for k := range productHostKey {
		res,err := d.Basic.GetHostPair(ctx,k,AllHostInfo)
		srcHost := &mutualtrust.Host{
			IP:       res[0].IP,
			Port:     22,
			User:     res[0].UserName,
			Password: res[0].Passwd,
		}
		dstHost := &mutualtrust.Host{
			IP:       res[1].IP,
			Port:     22,
			User:     res[1].UserName,
			Password: res[1].Passwd,
		}
		key, err := srcHost.SelectSecret()
		if err != nil {
			framework.LogWithContext(ctx).Errorf("read host[%s] secret key fail , %v", srcHost.IP, err)
			return err
		} else {
			err = dstHost.SecretWrite(key)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("write secret key to host[%s] fail , %v", dstHost.IP, err)
				return err
			}
		}
	}
	return nil
}

//GetBinPath : Get the server path variable
func (d *TrafficReplayDeploy) GetBinPath(ctx context.Context, config cmd.SessionConfig) (string, error) {
	framework.LogWithContext(ctx).Infof("begin get env bin path  on host %s", config.Host)
	defer framework.LogWithContext(ctx).Infof("end get env  bin path  on host %s", config.Host)
	r, err := cmd.NewRemoteSession(config)
	if err != nil {
		return "", err
	}
	err = r.Exec("echo $PATH", cmd.WaitCmdEnd)
	binPath := r.Output()

	errClose := r.Close()
	if errClose != nil {
		framework.LogWithContext(ctx).Warn("close ssh connect fail , " + errClose.Error())
	}

	if err != nil {
		return "", err
	}
	return binPath, nil
}

//CheckBinExist : Check if the binary program exists on the remote server
func (d *TrafficReplayDeploy) CheckBinExist(ctx context.Context, config cmd.SessionConfig,
	path, binName string) (bool, int, error) {

	framework.LogWithContext(ctx).Infof("begin check %s in dir %s on host %s", binName, path, config.Host)
	defer framework.LogWithContext(ctx).Infof("end check %s in dir %s on host %s", binName, path, config.Host)
	r, err := cmd.NewRemoteSession(config)
	if err != nil {
		return false, 0, err
	}

	err = r.Exec("ls "+path+"/"+binName, cmd.WaitCmdEnd)

	errClose := r.Close()
	if errClose != nil {
		framework.LogWithContext(ctx).Warn("close ssh connect fail , " + err.Error())
	}

	if err != nil {
		return false, 1, err
	}

	return true, 1, nil
}

//CheckProductTidbServerBinExist : Check if the corresponding binary file exists for the production environment
func (d *TrafficReplayDeploy) CheckProductTidbServerBinExist(ctx context.Context, binName string) error {
	productHostKeys := d.Basic.HostKyePairs[0]
	var exist bool
	var step int
	var binPaths string

	for k := range productHostKeys {
		//Check for the presence of binary programs in the PATH environment
		//variable to determine if the program is installed
		res ,err := d.Basic.GetHostPair(ctx,k,ProductHostInfo)
		if err !=nil{
			return err
		}
		config := cmd.NewSessionConfig(res[0].UserName, res[0].Passwd,
			res[0].IP, 22, 0)

		binPaths, err = d.GetBinPath(ctx, config)
		if err != nil {
			return err
		}

		binPath := strings.Split(binPaths, ":")
		for _, v := range binPath {
			exist, step, err = d.CheckBinExist(ctx, config, strings.Replace(v, "\n", "", -1), binName)
			if err != nil && step == 0 {
				//if connect to server fail ,will  return err
				return err
			} else if exist == true {
				d.TcpdumpExist[k] = true
			} else {
				//if the binary file does not exist,
				//continue to find another directory
				continue
			}
		}
	}

	return nil
}


//DeployBinOnRemoteServer : Invoke the yum command to install the program on a remote server
func (d *TrafficReplayDeploy) DeployBinOnRemoteServer(ctx context.Context,
	config cmd.SessionConfig, binName string) error {

	framework.LogWithContext(ctx).Infof("begin deploy %s on host %s", binName, config.Host)
	defer framework.LogWithContext(ctx).Infof("end deploy %s on host %s", binName, config.Host)

	r, err := cmd.NewRemoteSession(config)
	if err != nil {
		return err
	}

	//TODO :The current version installs the binary via yum
	err = r.Exec("yum install -y "+binName, cmd.WaitCmdEnd)
	errClose := r.Close()
	if errClose != nil {
		framework.LogWithContext(ctx).Errorf("close ssh connect fail , " + err.Error())
	}
	if err != nil {
		return err
	}
	return nil
}

//DeployTcpdump : Installing tcpdump on a tidb server without tcpdump software
func (d *TrafficReplayDeploy) DeployTcpdump(ctx context.Context) error {
	productHostKeys := d.Basic.HostKyePairs[0]
	tcpdumpExist := d.TcpdumpExist

	for k := range productHostKeys {
		if tcpdumpExist[k] == true {
			continue
		}
		res,err:=d.Basic.GetHostPair(ctx,k,ProductHostInfo)
		if err !=nil{
			return err
		}
		config := cmd.NewSessionConfig(res[0].UserName, res[0].Passwd,
			res[0].IP, 22, 0)

		err = d.DeployBinOnRemoteServer(ctx, config, "tcpdump")
		if err != nil {
			return err
		}
	}

	return nil
}

//DeployDBReplay : Deploy db-replay application via scp
func (d *TrafficReplayDeploy) DeployDBReplay(ctx context.Context) error {
	simulationHostKeys := d.Basic.HostKyePairs[1]

	for k := range simulationHostKeys {
		res,err:=d.Basic.GetHostPair(ctx,k,SimulationHostInfo)
		if err !=nil{
			return err
		}
		config := &cmd.ScpSessionConfig{
			Sess: cmd.NewSessionConfig(res[0].UserName, res[0].Passwd,
				res[0].IP, 22, 0),
			//TODO  to be confirmed
			SrcFileName: "",
			DstFileName: d.DeployPath[k],
			Permissions: "0665",
		}
		//db-replay deployment via remote copy of files
		err = config.ScpFile(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

//CreateDirOnRemoteServer : create dir on remote server
func (d *TrafficReplayDeploy) CreateDirOnRemoteServer(ctx context.Context,
	host *structs.HostInfo, path string) error {

	config := cmd.NewSessionConfig(host.UserName, host.Passwd, host.IP, 22, 0)

	r, err := cmd.NewRemoteSession(config)
	if err != nil {
		return err
	}


	//Create the directory by mkdir -p and set 754 permissions
	cmdStr := fmt.Sprintf("mkdir -p %s ; chmod -p %s 754", path, path)
	err = r.Exec(cmdStr, cmd.WaitCmdEnd)

	errClose := r.Close()
	if errClose != nil {
		framework.LogWithContext(ctx).Errorf("close connect fail ,%v", errClose)
		return errClose
	}
	if err != nil {
		return err
	}

	return nil
}

//GenerateDir : generate dir for traffic replay
func (d *TrafficReplayDeploy) GenerateDir(ctx context.Context) error {

	framework.LogWithContext(ctx).Infof("begin generate dir")
	defer framework.LogWithContext(ctx).Infof("end generate dir")
	//RootPath[0]: Production environment
	//RootPath[1]: Simulation environment
	//path : root-path + uuid
	p := d.Basic.HostKyePairs[0]
	s := d.Basic.HostKyePairs[1]

	//Production environment creates tcpdump hook temporary file storage directory
	for k, v := range p {
		rootPath ,ok := d.Basic.RootPath[v.Key]
		if !ok{
			return fmt.Errorf("can not find rootpath %s",v.Key)
		}
		d.TcpdumpTempPath[k] = rootPath + uuid.New()
	}
	//The simulation environment creates a temporary storage directory for pcap files,
	//a storage directory for pcap files, a storage directory for db-replay results files,
	//and a deployment directory for binary files
	for k, v := range s {
		rootPath ,ok := d.Basic.RootPath[v.Key]
		if !ok{
			return fmt.Errorf("can not find rootpath %s",v.Key)
		}
		d.DeployPath[k] = rootPath + uuid.New()
		//d.PcapTempPath[k] = rootPath + uuid.New()
		d.PcapFileStorePath[k] = rootPath + uuid.New()
		d.ResultFileStorePath[k] = rootPath + uuid.New()
	}
	return nil
}

//CreateDirForTcpdump : create dir for tcpdump
func (d *TrafficReplayDeploy) CreateDirForTcpdump(ctx context.Context) error {
	productHostKeys := d.Basic.HostKyePairs[0]
	for k := range productHostKeys {
		res,err := d.Basic.GetHostPair(ctx,k,ProductHostInfo)
		if err!=nil{
			return err
		}
		err = d.CreateDirOnRemoteServer(ctx, res[0], d.TcpdumpTempPath[k])
		if err != nil {
			return err
		}
	}
	return nil
}

//CreateDirForDBReplay : create dir for DB-Replay
func (d *TrafficReplayDeploy) CreateDirForDBReplay(ctx context.Context) error {
	simulationHostKeys := d.Basic.HostKyePairs[1]
	for k := range simulationHostKeys {
		res ,err := d.Basic.GetHostPair(ctx,k,SimulationHostInfo)
		if err!=nil{
			return err
		}
		err=d.CreateDirOnRemoteServer(ctx,res[0],d.DeployPath[k])
		if err !=nil{
			return err
		}
/*
		err = d.CreateDirOnRemoteServer(ctx, res[0], d.PcapTempPath[k])
		if err != nil {
			return err
		}
*/
		err = d.CreateDirOnRemoteServer(ctx, res[0], d.PcapFileStorePath[k])
		if err != nil {
			return err
		}

		err = d.CreateDirOnRemoteServer(ctx, res[0], d.ResultFileStorePath[k])
		if err != nil {
			return err
		}

	}

	return nil
}

//CreateDirForTrafficReplay : create dir for traffic replay
func (d *TrafficReplayDeploy) CreateDirForTrafficReplay(ctx context.Context) error {
	err := d.CreateDirForTcpdump(ctx)
	if err != nil {
		return err
	}

	err = d.CreateDirForDBReplay(ctx)
	if err != nil {
		return err
	}

	return nil
}

//GenerateHookText : Generate tcpdump file transfer hook command
//The current scheme is to transfer a pcap file name plus an .ok
//flag file after transferring the pcap file, and the flag to determine
//if the pcap file transfer is complete during db-replay playback is whether
//the ok file exists in the current directory
//TODO: db-replay needs to be modified to support this solution
func (d *TrafficReplayDeploy) GenerateHookText(host *structs.HostInfo, path string) string {
	txt := fmt.Sprintf("#!/usr/bin/bash \n "+"set -xe \n"+
		"scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $1 root@%s:%s \n"+
		"touch $1.ok \n"+
		"scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $1.ok root@%s:%s \n"+
		"rm -rf $1.ok \n"+
		"rm -rf $1 \n", host.IP, path,host.IP, path)

	return txt
}

//CreateHookScript : Create the tcpdump hock script with command
func (d *TrafficReplayDeploy) CreateHookScript(ctx context.Context) error {
	productHostKeys := d.Basic.HostKyePairs[0]


	for k := range productHostKeys {
		res,err:=d.Basic.GetHostPair(ctx,k,AllHostInfo)
		if err !=nil{
			return err
		}
		command := d.GenerateHookText(res[1], d.PcapFileStorePath[k])
		config := cmd.NewSessionConfig(res[0].UserName, res[0].Passwd,
			res[0].IP, 22, 0)

		r, err := cmd.NewRemoteSession(config)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("connect remote host %s fail ,%v ",
				config.Host, err)
			return err
		}
		err = r.Exec(fmt.Sprintf("echo %s > %s/rotate_hook.sh;chmod +x %s/rotate_hook.sh",
			command, d.TcpdumpTempPath[k], d.TcpdumpTempPath[k]), cmd.WaitCmdEnd)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("write rotate_hook.sh on host %s fail ,%v",
				config.Host, err)
			return err
		}
	}

	return nil
}

//DeployFileBeat : deploy file beat on simulation tidb server
func (d *TrafficReplayDeploy) DeployFileBeat(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin deploy file beat")
	defer framework.LogWithContext(ctx).Infof("end deploy file beat")
	//TODO : Install filebeat or configure filebeat.yml
	return nil
}

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

package hostinitiator

import (
	"context"
	"fmt"

	"github.com/pingcap-inc/tiem/util/scp"
	sshclient "github.com/pingcap-inc/tiem/util/ssh"

	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
)

type FileHostInitiator struct {
	sshClient       sshclient.SSHClientExecutor
	secondPartyServ secondparty.SecondPartyService
}

func NewFileHostInitiator() *FileHostInitiator {
	hostInitiator := new(FileHostInitiator)
	hostInitiator.sshClient = nil
	hostInitiator.secondPartyServ = secondparty.Manager
	return hostInitiator
}

func (p *FileHostInitiator) SetSSHClient(c sshclient.SSHClientExecutor) {
	p.sshClient = c
}

func (p *FileHostInitiator) SetSecondPartyServ(s secondparty.SecondPartyService) {
	p.secondPartyServ = s
}

func (p *FileHostInitiator) CopySSHID(ctx context.Context, h *structs.HostInfo) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("copy ssh id to host %s %s@%s", h.HostName, h.UserName, h.IP)

	err = scp.CopySSHID(ctx, h.IP, h.UserName, h.Passwd, rp_consts.DefaultCopySshIDTimeOut)
	if err != nil {
		log.Errorf("copy ssh id to host %s %s@%s failed, %v", h.HostName, h.UserName, h.IP, err)
		return err
	}

	return nil
}

func (p *FileHostInitiator) Prepare(ctx context.Context, h *structs.HostInfo) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("prepare for host %s %s begins", h.HostName, h.IP)

	tempateInfo := templateCheckHost{}
	tempateInfo.buildCheckHostTemplateItems(h)

	templateStr, err := tempateInfo.generateTopologyConfig(ctx)
	if err != nil {
		return err
	}

	resultStr, err := p.secondPartyServ.CheckTopo(ctx, secondparty.ClusterComponentTypeStr, templateStr, rp_consts.DefaultTiupTimeOut,
		[]string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa", "--apply", "--format", "json"})
	if err != nil {
		errMsg := fmt.Sprintf("call second serv to apply host %s %s [%v] failed, %v", h.HostName, h.IP, templateStr, err)
		return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
	}
	log.Infof("apply host %s %s done, %s", h.HostName, h.IP, resultStr)

	var results checkHostResults
	(&results).buildFromJson(resultStr)
	sortedResult := results.analyzeCheckResults()
	log.Infof("build from result json get sorted result: %v", sortedResult)

	fails, hasFails := sortedResult["Fail"]
	var needInstallNumaCtl = false
	var needSetSwap = false
	if hasFails {
		for _, fail := range *fails {
			if fail.Name == "command" {
				needInstallNumaCtl = true
			}
			if fail.Name == "swap" {
				needSetSwap = true
			}
		}
		log.Infof("after apply host %s %s, needInstallNumaCtl: %v, needSetSwap: %v", h.HostName, h.IP, needInstallNumaCtl, needSetSwap)
	}

	if !needInstallNumaCtl && !needSetSwap {
		log.Infof("no need to install numactl and set swap for host %s %s", h.HostName, h.IP)
		return nil
	}

	err = p.connectToHost(ctx, h)
	if err != nil {
		log.Errorf("connect to host %s %s failed, %v", h.HostName, h.IP, err)
		return err
	}
	defer p.closeConnect()

	if needInstallNumaCtl {
		if err = p.installNumaCtl(ctx, h); err != nil {
			errMsg := fmt.Sprintf("install numactl on host %s %s failed, %v", h.HostName, h.IP, err)
			return errors.NewError(errors.TIEM_RESOURCE_PREPARE_HOST_ERROR, errMsg)
		}
	}

	if needSetSwap {
		if err = p.setOffSwap(ctx, h); err != nil {
			errMsg := fmt.Sprintf("set off swap on host %s %s failed, %v", h.HostName, h.IP, err)
			return errors.NewError(errors.TIEM_RESOURCE_PREPARE_HOST_ERROR, errMsg)
		}
	}

	return nil
}

func (p *FileHostInitiator) Verify(ctx context.Context, h *structs.HostInfo) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("verify host %v begins", *h)
	tempateInfo := templateCheckHost{}
	tempateInfo.buildCheckHostTemplateItems(h)

	templateStr, err := tempateInfo.generateTopologyConfig(ctx)
	if err != nil {
		return err
	}
	ignoreWarnings, ok := ctx.Value(rp_consts.ContextIgnoreWarnings).(bool)
	if !ok {
		return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, "get ignore warning flag from context failed")
	}
	log.Infof("verify host %s %s ignore warning (%t)", h.HostName, h.IP, ignoreWarnings)

	resultStr, err := p.secondPartyServ.CheckTopo(ctx, secondparty.ClusterComponentTypeStr, templateStr, rp_consts.DefaultTiupTimeOut,
		[]string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa", "--format", "json"})
	if err != nil {
		errMsg := fmt.Sprintf("call second serv to check host %s %s [%v] failed, %v", h.HostName, h.IP, templateStr, err)
		return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
	}
	log.Infof("verify host %s %s for %v done", h.HostName, h.IP, tempateInfo)

	// deal with the result
	var results checkHostResults
	(&results).buildFromJson(resultStr)
	sortedResult := results.analyzeCheckResults()
	log.Infof("build from json %s, get sorted result: %v", resultStr, sortedResult)

	fails, hasFails := sortedResult["Fail"]
	if hasFails {
		errMsg := fmt.Sprintf("check host %s %s has %d fails, %v", h.HostName, h.IP, len(*fails), *fails)
		return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
	}

	warnings, hasWarns := sortedResult["Warn"]
	if hasWarns {
		errMsg := fmt.Sprintf("check host %s %s has %d warnings, %v", h.HostName, h.IP, len(*warnings), *warnings)
		log.Warnln(errMsg)
		if !ignoreWarnings {
			return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
		}
	}

	pass, hasPasses := sortedResult["Pass"]
	if hasPasses {
		log.Infof("check host %s %s succeed, %v", h.HostName, h.IP, *pass)
	} else {
		log.Warnf("check host %s %s no pass", h.HostName, h.IP)
	}

	return nil
}

func (p *FileHostInitiator) InstallSoftware(ctx context.Context, hosts []structs.HostInfo) (err error) {
	if err = p.installTcpDump(ctx, hosts); err != nil {
		return err
	}
	return nil
}

func (p *FileHostInitiator) JoinEMCluster(ctx context.Context, hosts []structs.HostInfo) (err error) {
	tempateInfo := templateScaleOut{}
	for _, host := range hosts {
		tempateInfo.HostIPs = append(tempateInfo.HostIPs, host.IP)
	}

	templateStr, err := tempateInfo.generateTopologyConfig(ctx)
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Infof("join em cluster on %s", templateStr)

	workFlowNodeID, ok := ctx.Value(rp_consts.ContextWorkFlowNodeIDKey).(string)
	if !ok || workFlowNodeID == "" {
		return errors.NewErrorf(errors.TIEM_RESOURCE_INIT_FILEBEAT_ERROR, "get work flow node from context failed, %s, %v", workFlowNodeID, ok)
	}

	emClusterName := framework.Current.GetClientArgs().EMClusterName
	framework.LogWithContext(ctx).Infof("join em cluster %s with work flow id %s", emClusterName, workFlowNodeID)
	operationId, err := p.secondPartyServ.ClusterScaleOut(ctx, secondparty.TiEMComponentTypeStr, emClusterName, templateStr, rp_consts.DefaultTiupTimeOut,
		[]string{"--user", "root", "-i", "/home/tiem/.ssh/tiup_rsa"}, workFlowNodeID, "")
	if err != nil {
		return errors.NewErrorf(errors.TIEM_RESOURCE_INIT_FILEBEAT_ERROR, "join em cluster %s [%v] failed, %v", emClusterName, templateStr, err)
	}
	framework.LogWithContext(ctx).Infof("join em cluster %s for %v in operationId %s", emClusterName, tempateInfo, operationId)

	return nil
}

func (p *FileHostInitiator) LeaveEMCluster(ctx context.Context, nodeId string) (err error) {
	framework.LogWithContext(ctx).Infof("host %s leave em cluster", nodeId)

	workFlowNodeID, ok := ctx.Value(rp_consts.ContextWorkFlowNodeIDKey).(string)
	if !ok || workFlowNodeID == "" {
		return errors.NewErrorf(errors.TIEM_RESOURCE_UNINSTALL_FILEBEAT_ERROR, "get work flow node from context failed, %s, %v", workFlowNodeID, ok)
	}

	emClusterName := framework.Current.GetClientArgs().EMClusterName
	framework.LogWithContext(ctx).Infof("leave em cluster %s with work flow id %s", emClusterName, workFlowNodeID)
	operationId, err := p.secondPartyServ.ClusterScaleIn(ctx, secondparty.TiEMComponentTypeStr, emClusterName, nodeId, rp_consts.DefaultTiupTimeOut,
		[]string{"--yes"}, workFlowNodeID)
	if err != nil {
		return errors.NewErrorf(errors.TIEM_RESOURCE_UNINSTALL_FILEBEAT_ERROR, "leave em cluster %s [%s] failed, %v", emClusterName, nodeId, err)
	}
	framework.LogWithContext(ctx).Infof("leave em cluster %s for %s in operationId %s", emClusterName, nodeId, operationId)

	return nil
}

func (p *FileHostInitiator) connectToHost(ctx context.Context, h *structs.HostInfo) (err error) {
	if p.sshClient != nil {
		errMsg := fmt.Sprintf("connect to %s %s failed, host initiator already has unclosed connect", h.HostName, h.IP)
		return errors.NewError(errors.TIEM_RESOURCE_CONNECT_TO_HOST_ERROR, errMsg)
	}
	p.sshClient = sshclient.NewSSHClient(h.IP, rp_consts.HostSSHPort, sshclient.Passwd, h.UserName, h.Passwd)
	if err = p.sshClient.Connect(); err != nil {
		return err
	}

	return nil
}

func (p *FileHostInitiator) closeConnect() {
	if p.sshClient != nil {
		p.sshClient.Close()
		p.sshClient = nil
	}
}

func (p *FileHostInitiator) setOffSwap(ctx context.Context, h *structs.HostInfo) (err error) {
	checkExisted := "count=`cat /etc/sysctl.conf | grep -E '^vm.swappiness = 0$' | wc -l`"
	changeConf := "if [ $count -eq 0 ] ; then echo 'vm.swappiness = 0'>> /etc/sysctl.conf; fi"
	flushCmd := "swapoff -a"
	updateCmd := "sysctl -p"
	fstabCmd := "sed -i '/swap/s/^\\(.*\\)$/#\\1/g' /etc/fstab"
	result, err := p.sshClient.RunCommandsInSession([]string{checkExisted, changeConf, flushCmd, updateCmd, fstabCmd})
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Infof("host %s [%s] set off swap, %v", h.HostName, h.IP, result)
	return nil
}

func (p *FileHostInitiator) installNumaCtl(ctx context.Context, h *structs.HostInfo) (err error) {
	installNumaCtrlCmd := "yum install -y numactl"
	result, err := p.sshClient.RunCommandsInSession([]string{installNumaCtrlCmd})
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Infof("host %s [%s] install numactl, %v", h.HostName, h.IP, result)
	return nil

}

func (p *FileHostInitiator) installTcpDump(ctx context.Context, hosts []structs.HostInfo) (err error) {
	return nil
}

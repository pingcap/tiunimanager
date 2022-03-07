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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pingcap-inc/tiem/deployment"

	sshclient "github.com/pingcap-inc/tiem/util/ssh"

	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
)

type FileHostInitiator struct {
	sshClient      sshclient.SSHClientExecutor
	deploymentServ deployment.Interface
}

func NewFileHostInitiator() *FileHostInitiator {
	hostInitiator := new(FileHostInitiator)
	hostInitiator.sshClient = sshclient.SSHExecutor{}
	hostInitiator.deploymentServ = deployment.M
	return hostInitiator
}

func (p *FileHostInitiator) SetSSHClient(c sshclient.SSHClientExecutor) {
	p.sshClient = c
}

func (p *FileHostInitiator) SetDeploymentServ(d deployment.Interface) {
	p.deploymentServ = d
}

func (p *FileHostInitiator) AuthHost(ctx context.Context, deployUser, userGroup string, h *structs.HostInfo) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("begin to auth host %s %s with %s:%s", h.HostName, h.IP, deployUser, userGroup)
	err = p.createDeployUser(ctx, deployUser, userGroup, h)
	if err != nil {
		log.Errorf("auth host failed, %v", err)
		return err
	}

	err = p.buildAuth(ctx, deployUser, h)
	if err != nil {
		log.Errorf("auth host failed after user created, %v", err)
		return err
	}

	log.Infof("auth host %s %s succeed", h.HostName, h.IP)
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

	// tiup args should be: []string{"--user", "xxx", "-i", "/home/tidb/.ssh/tiup_rsa", "--apply", "--format", "json"}
	args := framework.GetTiupAuthorizaitonFlag()
	args = append(args, "--apply")
	args = append(args, "--format")
	args = append(args, "json")
	tiupHomeForTidb := framework.GetTiupHomePathForTidb()
	resultStr, err := deployment.M.CheckConfig(ctx, deployment.TiUPComponentTypeCluster, templateStr, tiupHomeForTidb,
		args, rp_consts.DefaultTiupTimeOut)
	if err != nil {
		errMsg := fmt.Sprintf("call deployment serv to apply host %s %s [%v] failed, %v", h.HostName, h.IP, templateStr, err)
		return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
	}
	log.Infof("apply host %s %s done, %s", h.HostName, h.IP, resultStr)

	var results checkHostResults
	(&results).buildFromJson(resultStr)
	sortedResult := results.analyzeCheckResults()
	log.Infof("build from result json get sorted result: %v", sortedResult)

	err = p.autoFix(ctx, h, sortedResult)
	if err != nil {
		log.Errorf("auto fix host %s %s failed, %v", h.HostName, h.IP, err)
		return err
	}

	log.Infof("prepare for host %s %s succeed", h.HostName, h.IP)

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

	// tiup args should be: []string{"--user", "xxx", "-i", "/home/tidb/.ssh/tiup_rsa", "--format", "json"}
	args := framework.GetTiupAuthorizaitonFlag()
	args = append(args, "--format")
	args = append(args, "json")
	tiupHomeForTidb := framework.GetTiupHomePathForTidb()
	resultStr, err := deployment.M.CheckConfig(ctx, deployment.TiUPComponentTypeCluster, templateStr, tiupHomeForTidb,
		args, rp_consts.DefaultTiupTimeOut)
	if err != nil {
		errMsg := fmt.Sprintf("call deployment serv to check host %s %s [%v] failed, %v", h.HostName, h.IP, templateStr, err)
		return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
	}
	log.Infof("verify host %s %s for %v done", h.HostName, h.IP, tempateInfo)

	// deal with the result
	var results checkHostResults
	(&results).buildFromJson(resultStr)
	sortedResult := results.analyzeCheckResults()
	log.Infof("build from json %s, get sorted result: %v", resultStr, sortedResult)

	fails, hasFails := sortedResult[string(Fail)]
	if hasFails {
		errMsg := fmt.Sprintf("check host %s %s has %d fails, %v", h.HostName, h.IP, len(*fails), *fails)
		return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
	}

	warnings, hasWarns := sortedResult[string(Warn)]
	if hasWarns {
		errMsg := fmt.Sprintf("check host %s %s has %d warnings, %v", h.HostName, h.IP, len(*warnings), *warnings)
		log.Warnln(errMsg)
		if !ignoreWarnings {
			ignoreCpuGovWarn, err := p.passCpuGovernorWarn(ctx, h, warnings)
			if err == nil && ignoreCpuGovWarn {
				log.Infof("ignore cpu governor warning for vm %s %s", h.HostName, h.IP)
			} else {
				return errors.NewError(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, errMsg)
			}
		}
	}

	pass, hasPasses := sortedResult[string(Pass)]
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

func (p *FileHostInitiator) PreCheckHostInstallFilebeat(ctx context.Context, hosts []structs.HostInfo) (installed bool, err error) {
	log := framework.LogWithContext(ctx)
	log.Infoln("begin precheck before join em cluster")
	emClusterName := framework.Current.GetClientArgs().EMClusterName

	// Parse EM topology structure to check whether filebeat has been installed already
	tiupHomeForTiem := framework.GetTiupHomePathForTiem()
	result, err := p.deploymentServ.Display(ctx, deployment.TiUPComponentTypeEM, emClusterName, tiupHomeForTiem, []string{"--json"}, rp_consts.DefaultTiupTimeOut)
	if err != nil {
		log.Errorf("precheck before join em cluster failed, %v", err)
		return false, errors.NewErrorf(errors.TIEM_RESOURCE_INIT_FILEBEAT_ERROR, "precheck join em cluster %s failed, %v", emClusterName, err)
	}
	emTopo := new(structs.EMMetaTopo)
	err = json.Unmarshal([]byte(result), emTopo)
	if err != nil {
		return false, errors.NewErrorf(errors.TIEM_RESOURCE_INIT_FILEBEAT_ERROR, "precheck join em cluster %s failed on umarshal, %v", emClusterName, err)
	}
	installed = false
LOOP:
	for _, instance := range emTopo.Instances {
		if instance.Role == "filebeat" {
			for _, host := range hosts {
				if instance.Host == host.IP {
					installed = true
					log.Infof("host %s %s has been install filebeat", host.HostName, host.IP)
					break LOOP
				}
			}
		}
	}
	return installed, nil
}

func (p *FileHostInitiator) JoinEMCluster(ctx context.Context, hosts []structs.HostInfo) (operationID string, err error) {
	tempateInfo := templateScaleOut{}
	for _, host := range hosts {
		tempateInfo.HostAddrs = append(tempateInfo.HostAddrs, HostAddr{
			HostIP:  host.IP,
			SSHPort: int(host.SSHPort),
		})
	}

	templateStr, err := tempateInfo.generateTopologyConfig(ctx)
	if err != nil {
		return "", err
	}
	framework.LogWithContext(ctx).Infof("join em cluster on %s", templateStr)

	workFlowID, ok := ctx.Value(rp_consts.ContextWorkFlowIDKey).(string)
	if !ok || workFlowID == "" {
		return "", errors.NewErrorf(errors.TIEM_RESOURCE_INIT_FILEBEAT_ERROR, "get work flow from context failed, %s, %v", workFlowID, ok)
	}

	emClusterName := framework.Current.GetClientArgs().EMClusterName
	framework.LogWithContext(ctx).Infof("join em cluster %s with work flow id %s", emClusterName, workFlowID)
	args := framework.GetTiupAuthorizaitonFlag()
	tiupHomeForTiem := framework.GetTiupHomePathForTiem()
	operationID, err = deployment.M.ScaleOut(ctx, deployment.TiUPComponentTypeEM, emClusterName, templateStr,
		tiupHomeForTiem, workFlowID, args, rp_consts.DefaultTiupTimeOut)
	if err != nil {
		return "", errors.NewErrorf(errors.TIEM_RESOURCE_INIT_FILEBEAT_ERROR, "join em cluster %s [%v] failed, %v", emClusterName, templateStr, err)
	}
	framework.LogWithContext(ctx).Infof("join em cluster %s for %v in operationID %s", emClusterName, tempateInfo, operationID)

	return operationID, nil
}

func (p *FileHostInitiator) LeaveEMCluster(ctx context.Context, nodeId string) (operationID string, err error) {
	framework.LogWithContext(ctx).Infof("host %s leave em cluster", nodeId)

	workFlowID, ok := ctx.Value(rp_consts.ContextWorkFlowIDKey).(string)
	if !ok || workFlowID == "" {
		return "", errors.NewErrorf(errors.TIEM_RESOURCE_UNINSTALL_FILEBEAT_ERROR, "get work flow from context failed, %s, %v", workFlowID, ok)
	}

	emClusterName := framework.Current.GetClientArgs().EMClusterName
	tiupHomeForTiem := framework.GetTiupHomePathForTiem()
	framework.LogWithContext(ctx).Infof("leave em cluster %s with work flow id %s", emClusterName, workFlowID)
	operationID, err = deployment.M.ScaleIn(ctx, deployment.TiUPComponentTypeEM, emClusterName,
		nodeId, tiupHomeForTiem,
		workFlowID, []string{}, rp_consts.DefaultTiupTimeOut)
	if err != nil {
		return "", errors.NewErrorf(errors.TIEM_RESOURCE_UNINSTALL_FILEBEAT_ERROR, "leave em cluster %s [%s] failed, %v", emClusterName, nodeId, err)
	}
	framework.LogWithContext(ctx).Infof("leave em cluster %s for %s in operationId %s", emClusterName, nodeId, operationID)

	return operationID, nil
}

func (p *FileHostInitiator) installTcpDump(ctx context.Context, hosts []structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) passCpuGovernorWarn(ctx context.Context, h *structs.HostInfo, warnings *[]checkHostResult) (ok bool, err error) {
	log := framework.LogWithContext(ctx)
	var needCheckVM = false
	if len(*warnings) == 1 {
		// xx.xx.xx.xx  cpu-governor    Warn    Unable to determine current CPU frequency governor policy
		if (*warnings)[0].Name == "cpu-governor" && strings.HasPrefix((*warnings)[0].Message, "Unable to determine") {
			needCheckVM = true
		}
	}
	log.Infof("need check vm (%v) for host %s %s", needCheckVM, h.HostName, h.IP)
	if needCheckVM {
		isVm, err := p.isVirtualMachine(ctx, h)
		if err == nil && isVm {
			return true, nil
		} else {
			return false, err
		}
	}
	return false, nil
}

func (p *FileHostInitiator) isVirtualMachine(ctx context.Context, h *structs.HostInfo) (isVM bool, err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("begin to check host manufacturer on host %s %s", h.HostName, h.IP)
	vmManufacturer := []string{"QEMU", "XEN", "KVM", "VMWARE", "VIRTUALBOX", "VBOX", "ORACLE", "MICROSOFT", "ZVM", "BOCHS", "PARALLELS", "UML"}
	dmidecodeCmd := "dmidecode -s system-manufacturer | tr -d '\n'"
	result, err := p.sshClient.RunCommandsInRemoteHost(h.IP, int(h.SSHPort), sshclient.Passwd, h.UserName, h.Passwd, true, rp_consts.DefaultCopySshIDTimeOut, []string{dmidecodeCmd})
	if err != nil {
		log.Errorf("execute %s on host %s %s failed, %v", dmidecodeCmd, h.HostName, h.IP, err)
		return false, err
	}
	isVM = false
	for _, vm := range vmManufacturer {
		if strings.EqualFold(result, vm) {
			isVM = true
			break
		}
	}
	log.Infof("host %s [%s] manufacturer is %s, should be VM (%v)", h.HostName, h.IP, result, isVM)
	return isVM, nil
}

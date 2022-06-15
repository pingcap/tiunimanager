/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

	"github.com/pingcap/tiunimanager/deployment"
	"github.com/pingcap/tiunimanager/models"

	sshclient "github.com/pingcap/tiunimanager/util/ssh"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	rp_consts "github.com/pingcap/tiunimanager/micro-cluster/resourcemanager/resourcepool/constants"
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

func (p *FileHostInitiator) skipAuthHost(ctx context.Context, deployUser string, h *structs.HostInfo) bool {
	log := framework.LogWithContext(ctx)
	specifiedUser := framework.GetCurrentSpecifiedUser()
	specifiedPrivateKey := framework.GetCurrentSpecifiedPrivateKeyPath()
	specifiedPublicKey := framework.GetCurrentSpecifiedPublicKeyPath()
	log.Infof("begin to test whether skip auth host %s %s with deploy user %s, with key pair <%s, %s> and specified user %s",
		h.HostName, h.IP, deployUser, specifiedPublicKey, specifiedPrivateKey, specifiedUser)
	// try to test connection if user specified loginUserName == deployUser
	if specifiedUser == deployUser {
		defaultPrivateKey := framework.GetPrivateKeyFilePath(deployUser)
		defaultPublicKey := framework.GetPublicKeyFilePath(deployUser)
		if specifiedPrivateKey == "" {
			specifiedPrivateKey = defaultPrivateKey
		}
		if specifiedPublicKey == "" {
			specifiedPublicKey = defaultPublicKey
		}
		// do the auth check only in the case when using the default deployUser key pair, because the only way we get the private key
		// for calling tiup commands is to call `GetPrivateKeyFilePath()`, which returns back the deployUser's default private key.
		if specifiedPrivateKey == defaultPrivateKey && specifiedPublicKey == defaultPublicKey {
			lsCmd := "ls -l"
			authenticate := sshclient.HostAuthenticate{SshType: sshclient.Key, AuthenticatedUser: deployUser, AuthenticateContent: specifiedPrivateKey}
			_, err := p.sshClient.RunCommandsInRemoteHost(h.IP, int(h.SSHPort), authenticate, true, rp_consts.DefaultCopySshIDTimeOut, []string{lsCmd})
			if err != nil {
				log.Errorf("check connection to host %s %s@%s:%d by execute \"%s\" failed, %v", h.HostName, deployUser, h.IP, h.SSHPort, lsCmd, err)
				return false
			}
			log.Infof("skip auth host %s %s succeed", h.HostName, h.IP)
			return true
		}
		log.Infof("can not skip auth host %s %s because either public key %s or private key %s is not specified as default %s, %s",
			h.HostName, h.IP, specifiedPublicKey, specifiedPrivateKey, defaultPublicKey, defaultPrivateKey)
		return false
	}
	log.Infof("specified user %s is different with deploy user %s", specifiedUser, deployUser)
	return false
}

// Create deployUser on target host and set up auth key
func (p *FileHostInitiator) AuthHost(ctx context.Context, deployUser, userGroup string, h *structs.HostInfo) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("begin to auth host %s %s with %s:%s", h.HostName, h.IP, deployUser, userGroup)

	if p.skipAuthHost(ctx, deployUser, h) {
		log.Infof("skip auth host %s %s@%s:%d with specified private key %s", h.HostName, deployUser, h.IP, h.SSHPort, framework.GetCurrentSpecifiedPrivateKeyPath())
		return nil
	}

	authenticate, err := p.getUserSpecifiedAuthenticateToHost(ctx, h)
	if err != nil {
		log.Errorf("get host %s %s authenticate content failed, %v", h.HostName, h.IP, err)
		return err
	}

	// create deploy user on target host and set up auth key
	err = p.createDeployUser(ctx, deployUser, userGroup, h, authenticate)
	if err != nil {
		log.Errorf("auth host failed, %v", err)
		return err
	}

	err = p.buildAuth(ctx, deployUser, h, authenticate)
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
		return errors.NewError(errors.TIUNIMANAGER_RESOURCE_HOST_NOT_EXPECTED, errMsg)
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
	log.Infof("verify host %s %s begins", h.HostName, h.IP)
	tempateInfo := templateCheckHost{}
	tempateInfo.buildCheckHostTemplateItems(h)

	templateStr, err := tempateInfo.generateTopologyConfig(ctx)
	if err != nil {
		return err
	}
	ignoreWarnings, ok := ctx.Value(rp_consts.ContextIgnoreWarnings).(bool)
	if !ok {
		return errors.NewError(errors.TIUNIMANAGER_RESOURCE_HOST_NOT_EXPECTED, "get ignore warning flag from context failed")
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
		return errors.NewError(errors.TIUNIMANAGER_RESOURCE_HOST_NOT_EXPECTED, errMsg)
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
		return errors.NewError(errors.TIUNIMANAGER_RESOURCE_HOST_NOT_EXPECTED, errMsg)
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
				return errors.NewError(errors.TIUNIMANAGER_RESOURCE_HOST_NOT_EXPECTED, errMsg)
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
	tiupHomeForEm := framework.GetTiupHomePathForEm()
	result, err := p.deploymentServ.Display(ctx, deployment.TiUPComponentTypeEM, emClusterName, tiupHomeForEm, []string{"--json"}, rp_consts.DefaultTiupTimeOut)
	if err != nil {
		log.Errorf("precheck before join em cluster failed, %v", err)
		return false, errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_INIT_FILEBEAT_ERROR, "precheck join em cluster %s failed, %v", emClusterName, err)
	}
	emTopo := new(structs.EMMetaTopo)
	err = json.Unmarshal([]byte(result), emTopo)
	if err != nil {
		return false, errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_INIT_FILEBEAT_ERROR, "precheck join em cluster %s failed on umarshal, %v", emClusterName, err)
	}
	installed = false
	// []hosts only contains one host since importing each host in a async workflow
	for _, host := range hosts {
		installed, err = p.checkInstanceInstalled(ctx, emTopo, host.IP, constants.EMInstanceNameOfFileBeat)
		if err != nil {
			return false, err
		}
	}

	return installed, nil
}

func (p *FileHostInitiator) checkInstanceInstalled(ctx context.Context, emTopo *structs.EMMetaTopo, hostIp string, instance string) (installed bool, err error) {
	log := framework.LogWithContext(ctx)
	for _, emInstance := range emTopo.Instances {
		if emInstance.Role == instance && emInstance.Host == hostIp {
			log.Infof("%s has been installed on host %s in status %s", instance, hostIp, emInstance.Status)
			if emInstance.Status == string(constants.EMInstanceUP) {
				return true, nil
			} else {
				return false, errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_BAD_INSTANCE_EXIST, "%s has been installed on host %s but in %s status", instance, hostIp, emInstance.Status)
			}
		}
	}
	return false, nil
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
		return "", errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_INIT_FILEBEAT_ERROR, "get work flow from context failed, %s, %v", workFlowID, ok)
	}

	emClusterName := framework.Current.GetClientArgs().EMClusterName
	framework.LogWithContext(ctx).Infof("join em cluster %s with work flow id %s", emClusterName, workFlowID)
	args := framework.GetTiupAuthorizaitonFlag()
	tiupHomeForEm := framework.GetTiupHomePathForEm()
	operationID, err = deployment.M.ScaleOut(ctx, deployment.TiUPComponentTypeEM, emClusterName, templateStr,
		tiupHomeForEm, workFlowID, args, rp_consts.DefaultTiupTimeOut)
	if err != nil {
		return "", errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_INIT_FILEBEAT_ERROR, "join em cluster %s [%v] failed, %v", emClusterName, templateStr, err)
	}
	framework.LogWithContext(ctx).Infof("join em cluster %s for %v in operationID %s", emClusterName, tempateInfo, operationID)

	return operationID, nil
}

func (p *FileHostInitiator) LeaveEMCluster(ctx context.Context, nodeId string) (operationID string, err error) {
	framework.LogWithContext(ctx).Infof("host %s leave em cluster", nodeId)

	workFlowID, ok := ctx.Value(rp_consts.ContextWorkFlowIDKey).(string)
	if !ok || workFlowID == "" {
		return "", errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_UNINSTALL_FILEBEAT_ERROR, "get work flow from context failed, %s, %v", workFlowID, ok)
	}

	emClusterName := framework.Current.GetClientArgs().EMClusterName
	tiupHomeForEm := framework.GetTiupHomePathForEm()
	framework.LogWithContext(ctx).Infof("leave em cluster %s with work flow id %s", emClusterName, workFlowID)
	operationID, err = deployment.M.ScaleIn(ctx, deployment.TiUPComponentTypeEM, emClusterName,
		nodeId, tiupHomeForEm,
		workFlowID, []string{}, rp_consts.DefaultTiupTimeOut)
	if err != nil {
		return "", errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_UNINSTALL_FILEBEAT_ERROR, "leave em cluster %s [%s] failed, %v", emClusterName, nodeId, err)
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
	// TODO: Add extraVMManufacturer []string read from system config table for user specified env.
	vmManufacturer := []string{"QEMU", "XEN", "KVM", "VMWARE", "VIRTUALBOX", "ALIBABA", "VBOX", "ORACLE", "MICROSOFT", "ZVM", "BOCHS", "PARALLELS", "UML"}
	dmidecodeCmd := "dmidecode -s system-manufacturer | tr -d '\n'"
	authenticate := p.getEMAuthenticateToHost(ctx)
	result, err := p.sshClient.RunCommandsInRemoteHost(h.IP, int(h.SSHPort), *authenticate, true, rp_consts.DefaultCopySshIDTimeOut, []string{dmidecodeCmd})
	if err != nil {
		log.Errorf("execute %s on host %s %s failed, %v", dmidecodeCmd, h.HostName, h.IP, err)
		return false, err
	}
	isVM = false
	for _, vm := range vmManufacturer {
		if strings.Contains(strings.ToLower(result), strings.ToLower(vm)) {
			isVM = true
			break
		}
	}

	if !isVM {
		isVM = p.extraVMManufacturerCheck(ctx, result)
	}

	log.Infof("host %s [%s] manufacturer is %s, should be VM (%v)", h.HostName, h.IP, result, isVM)
	return isVM, nil
}

// Try to match the user specified vm facturer, give a warn if failed
func (p *FileHostInitiator) extraVMManufacturerCheck(ctx context.Context, facturer string) bool {
	extraConfig, err := models.GetConfigReaderWriter().GetConfig(ctx, constants.ConfigKeyExtraVMFacturer)
	if err != nil {
		framework.LogWithContext(ctx).Warnf("get config ConfigKeyExtraVMFacturer failed, %v", err)
		return false
	}

	extraConfigVMFacturer := extraConfig.ConfigValue
	if extraConfigVMFacturer == "" {
		framework.LogWithContext(ctx).Infoln("extra vm facturer is not specified")
		return false
	}

	if strings.Contains(strings.ToLower(facturer), strings.ToLower(extraConfigVMFacturer)) {
		return true
	}

	return false
}

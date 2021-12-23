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
	"bytes"
	"context"
	"strconv"
	"strings"
	"text/template"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	sshclient "github.com/pingcap-inc/tiem/library/util/ssh"
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

func (p *FileHostInitiator) Verify(ctx context.Context, h *structs.HostInfo) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("verify host %v begins", *h)
	defer log.Infof("verify host %s %s end, %v", h.HostName, h.IP, err)

	err = p.verifyConnect(ctx, h)
	if err != nil {
		return err
	}
	defer p.closeSSHConnect()

	if err = p.verifyCpuMem(ctx, h); err != nil {
		return err
	}

	if err = p.verifyDisks(ctx, h); err != nil {
		return err
	}

	if err = p.verifyFS(ctx, h); err != nil {
		return err
	}

	if err = p.verifySwap(ctx, h); err != nil {
		return err
	}

	if err = p.verifyEnv(ctx, h); err != nil {
		return err
	}

	if err = p.verifyOSEnv(ctx, h); err != nil {
		return err
	}

	return nil
}

func (p *FileHostInitiator) SetConfig(ctx context.Context, h *structs.HostInfo) (err error) {
	log := framework.LogWithContext(ctx)
	log.Infof("set host config %v begins", *h)
	defer log.Infof("set host %s %s config end, %v", h.HostName, h.IP, err)

	return nil
}

func (p *FileHostInitiator) InstallSoftware(ctx context.Context, hosts []structs.HostInfo) (err error) {
	if err = p.installFileBeat(ctx, hosts); err != nil {
		return err
	}
	if err = p.installTcpDump(ctx, hosts); err != nil {
		return err
	}
	return nil
}

func (p *FileHostInitiator) verifyConnect(ctx context.Context, h *structs.HostInfo) (err error) {
	p.sshClient = sshclient.NewSSHClient(h.IP, hostSSHPort, sshclient.Passwd, h.UserName, h.Passwd)
	if err = p.sshClient.Connect(); err != nil {
		return err
	}
	return nil
}

func (p *FileHostInitiator) closeSSHConnect() {
	if p.sshClient != nil {
		p.sshClient.Close()
	}
}

func (p *FileHostInitiator) verifyCpuMem(ctx context.Context, h *structs.HostInfo) (err error) {
	getArchCmd := "lscpu | grep 'Architecture:' | awk '{print $2}'"
	arch, err := p.sshClient.RunCommandsInSession([]string{getArchCmd})
	if err != nil {
		return err
	}
	if !strings.EqualFold(arch, h.Arch) {
		return errors.NewEMErrorf(errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, "Host %s [%s] arch %s is not as import %s", h.HostName, h.IP, arch, h.Arch)
	}

	getCpuCoresCmd := "lscpu | grep 'CPU(s):' | awk '{print $2}'"
	cpuCoreStr, err := p.sshClient.RunCommandsInSession([]string{getCpuCoresCmd})
	if err != nil {
		return err
	}
	cpuCores, err := strconv.Atoi(cpuCoreStr)
	if err != nil {
		return err
	}
	if cpuCores != int(h.CpuCores) {
		framework.LogWithContext(ctx).Warnf("host %s [%s] cpuCores %d is not as import %d", h.HostName, h.IP, cpuCores, h.CpuCores)
	}

	getMemCmd := "free -g | grep 'Mem:' | awk '{print $2}'"
	memStr, err := p.sshClient.RunCommandsInSession([]string{getMemCmd})
	if err != nil {
		return err
	}
	mem, err := strconv.Atoi(memStr)
	if err != nil {
		return err
	}
	if mem != int(h.Memory) {
		framework.LogWithContext(ctx).Warnf("host %s [%s] memory %d is not as import %d", h.HostName, h.IP, mem, h.Memory)
	}
	return nil
}

func (p *FileHostInitiator) verifyDisks(ctx context.Context, h *structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) verifyFS(ctx context.Context, h *structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) verifySwap(ctx context.Context, h *structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) verifyEnv(ctx context.Context, h *structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) verifyOSEnv(ctx context.Context, h *structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) setOffSwap(ctx context.Context, h *structs.HostInfo) (err error) {
	changeConf := "echo 'vm.swappiness = 0'>> /etc/sysctl.conf"
	flushCmd := "swapoff -a && swapon -a"
	updateCmd := "sysctl -p"
	result, err := p.sshClient.RunCommandsInSession([]string{changeConf, flushCmd, updateCmd})
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Infof("host %s [%s] set off swap, %v", h.HostName, h.IP, result)
	return nil
}

const hostSSHPort = 22
const fileBeatDataDir = "/tiem-data"
const fileBeatDeployDir = "/tiem-deploy"

type templateScaleOut struct {
	Arch      string
	DeployDir string
	DataDir   string
	HostIPs   []string
}

func (p *templateScaleOut) generateTopologyConfig(ctx context.Context) (string, error) {
	t, err := template.New("import_topology.yaml").ParseFiles("template/import_topology.yaml")
	if err != nil {
		return "", errors.NewError(errors.TIEM_PARAMETER_INVALID, err.Error())
	}

	topology := new(bytes.Buffer)
	if err = t.Execute(topology, p); err != nil {
		return "", errors.NewError(errors.TIEM_UNRECOGNIZED_ERROR, err.Error())
	}
	framework.LogWithContext(ctx).Infof("generate topology config: %s", topology.String())

	return topology.String(), nil
}

func (p *FileHostInitiator) installFileBeat(ctx context.Context, hosts []structs.HostInfo) (err error) {
	arch := constants.GetArchAlias(constants.ArchType(hosts[0].Arch))
	tempateInfo := templateScaleOut{
		Arch:      arch,
		DeployDir: fileBeatDeployDir,
		DataDir:   fileBeatDataDir,
	}
	for _, host := range hosts {
		tempateInfo.HostIPs = append(tempateInfo.HostIPs, host.IP)
	}
	templateStr, err := tempateInfo.generateTopologyConfig(ctx)
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Infof("install filebeat on %s", templateStr)

	operationId, err := p.secondPartyServ.ClusterScaleOut(ctx, secondparty.TiEMComponentTypeStr, "", templateStr, 0, nil, "")
	if err != nil {
		return errors.NewEMErrorf(errors.TIEM_RESOURCE_INIT_FILEBEAT_ERROR, "install filebeat [%v] failed, %v", templateStr, err)
	}
	framework.LogWithContext(ctx).Infof("installing filebeat for %v in operationId %s", tempateInfo, operationId)

	return nil
}

func (p *FileHostInitiator) installTcpDump(ctx context.Context, hosts []structs.HostInfo) (err error) {
	return nil
}

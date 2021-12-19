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
	"strconv"
	"strings"

	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	sshclient "github.com/pingcap-inc/tiem/library/util/ssh"
)

type FileHostInitiator struct {
}

func NewFileHostInitiator() *FileHostInitiator {
	hostInitiator := new(FileHostInitiator)
	return hostInitiator
}

func (p *FileHostInitiator) VerifyConnect(ctx context.Context, h *structs.HostInfo) (client *sshclient.SSHClient, err error) {
	client = sshclient.NewSSHClient(h.IP, 22, sshclient.Passwd, h.UserName, h.Passwd)
	if err = client.Connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func (p *FileHostInitiator) VerifyCpuMem(ctx context.Context, c *sshclient.SSHClient, h *structs.HostInfo) (err error) {
	getArchCmd := "lscpu | grep 'Architecture:' | awk '{print $2}'"
	arch, err := c.RunCommandsInSession([]string{getArchCmd})
	if err != nil {
		return err
	}
	if !strings.EqualFold(arch, h.Arch) {
		return framework.NewTiEMErrorf(common.TIEM_RESOURCE_HOST_NOT_EXPECTED, "Host %s [%s] arch %s is not as import %s", h.HostName, h.IP, arch, h.Arch)
	}

	getCpuCoresCmd := "lscpu | grep 'CPU(s):' | awk '{print $2}'"
	cpuCoreStr, err := c.RunCommandsInSession([]string{getCpuCoresCmd})
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
	memStr, err := c.RunCommandsInSession([]string{getMemCmd})
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

func (p *FileHostInitiator) VerifyDisks(ctx context.Context, c *sshclient.SSHClient, h *structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) VerifyFS(ctx context.Context, c *sshclient.SSHClient, h *structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) VerifySwap(ctx context.Context, c *sshclient.SSHClient, h *structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) VerifyEnv(ctx context.Context, c *sshclient.SSHClient, h *structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) VerifyOSEnv(ctx context.Context, c *sshclient.SSHClient, h *structs.HostInfo) (err error) {
	return nil
}

func (p *FileHostInitiator) SetOffSwap(ctx context.Context, c *sshclient.SSHClient, h *structs.HostInfo) (err error) {
	changeConf := "echo 'vm.swappiness = 0'>> /etc/sysctl.conf"
	flushCmd := "swapoff -a && swapon -a"
	updateCmd := "sysctl -p"
	result, err := c.RunCommandsInSession([]string{changeConf, flushCmd, updateCmd})
	if err != nil {
		return err
	}
	framework.LogWithContext(ctx).Infof("host %s [%s] set off swap, %v", h.HostName, h.IP, result)
	return nil
}

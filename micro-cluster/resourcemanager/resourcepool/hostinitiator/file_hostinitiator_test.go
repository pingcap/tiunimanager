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
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	mock_secp "github.com/pingcap-inc/tiem/test/mocksecondparty_v2"
	mock_ssh "github.com/pingcap-inc/tiem/test/mockutil/mocksshclientexecutor"
	"github.com/stretchr/testify/assert"
)

func Test_CopySSHID(t *testing.T) {
	fileInitiator := NewFileHostInitiator()

	framework.InitBaseFrameworkForUt(framework.ClusterService)
	err := fileInitiator.CopySSHID(context.TODO(), &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"})
	assert.NotNil(t, err)
}

func Test_VerifyCpuMem_Succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).DoAndReturn(func(commands []string) (string, error) {
		command := strings.Join(commands, ";")
		if strings.Contains(command, "Architecture:") {
			return "x86_64", nil
		}
		if strings.Contains(command, "CPU(s):") {
			return "32", nil
		}
		if strings.Contains(command, "Mem:") {
			return "8", nil
		}
		return "", errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "BadRequest")
	}).Times(3)

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.verifyCpuMem(context.TODO(), &structs.HostInfo{Arch: "x86_64", CpuCores: 32, Memory: 8})
	assert.Nil(t, err)
}

func Test_VerifyCpuMem_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).DoAndReturn(func(commands []string) (string, error) {
		command := strings.Join(commands, ";")
		if strings.Contains(command, "Architecture:") {
			return "ARM64", nil
		}
		return "", errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "BadRequest")
	}).Times(1)

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.verifyCpuMem(context.TODO(), &structs.HostInfo{Arch: "x86_64", CpuCores: 32, Memory: 8})
	assert.NotNil(t, err)
}

func Test_SetConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.SetConfig(context.TODO(), &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_VerifyDisks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.verifyDisks(context.TODO(), &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_VerifyFS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.verifyFS(context.TODO(), &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_VerifySwap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.verifySwap(context.TODO(), &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_VerifyEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.verifyEnv(context.TODO(), &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_VerifyOSEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.verifyOSEnv(context.TODO(), &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_SetOffSwap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.setOffSwap(context.TODO(), &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_GenerateTopologyConfig(t *testing.T) {
	template_struct := templateScaleOut{}
	template_struct.HostIPs = append(template_struct.HostIPs, "192.168.177.177")
	template_struct.HostIPs = append(template_struct.HostIPs, "192.168.177.178")
	template_struct.HostIPs = append(template_struct.HostIPs, "192.168.177.179")

	str, err := template_struct.generateTopologyConfig(context.TODO())
	assert.Nil(t, err)
	t.Log(str)
}

func Test_InstallSoftware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.InstallSoftware(context.TODO(), []structs.HostInfo{{Arch: "X86_64", IP: "192.168.177.180"}})
	assert.Nil(t, err)
}

func Test_JoinEMCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSec := mock_secp.NewMockSecondPartyService(ctrl)
	mockSec.EXPECT().ClusterScaleOut(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSecondPartyServ(mockSec)

	ctx := context.WithValue(context.TODO(), rp_consts.ContextWorkFlowNodeIDKey, "fake-node-id")
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	err := fileInitiator.JoinEMCluster(ctx, []structs.HostInfo{{Arch: "X86_64", IP: "192.168.177.180"}})
	assert.Nil(t, err)
}

func Test_LeaveEMCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSec := mock_secp.NewMockSecondPartyService(ctrl)
	mockSec.EXPECT().ClusterScaleIn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSecondPartyServ(mockSec)

	ctx := context.WithValue(context.TODO(), rp_consts.ContextWorkFlowNodeIDKey, "fake-node-id")
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	err := fileInitiator.LeaveEMCluster(ctx, "192.168.177.180:0")
	assert.Nil(t, err)
}

func Test_BuildHostCheckResulsFromJson(t *testing.T) {
	jsonStr := `{"result":[
		{"node":"172.16.6.252","name":"os-version","status":"Pass","message":"OS is CentOS Linux 7 (Core) 7.6.1810"},
		{"node":"172.16.6.252","name":"cpu-cores","status":"Pass","message":"number of CPU cores / threads: 4"},
		{"node":"172.16.6.252","name":"cpu-governor","status":"Warn","message":"Unable to determine current CPU frequency governor policy"},
		{"node":"172.16.6.252","name":"swap","status":"Fail","message":"swap is enabled, please disable it for best performance"},
		{"node":"172.16.6.252","name":"memory","status":"Pass","message":"memory size is 8192MB"},
		{"node":"172.16.6.252","name":"disk","status":"Warn","message":"mount point /home does not have 'noatime' option set"},
		{"node":"172.16.6.252","name":"limits","status":"Fail","message":"soft limit of 'nofile' for user 'tidb' is not set or too low"},
		{"node":"172.16.6.252","name":"limits","status":"Fail","message":"hard limit of 'nofile' for user 'tidb' is not set or too low"},
		{"node":"172.16.6.252","name":"limits","status":"Fail","message":"soft limit of 'stack' for user 'tidb' is not set or too low"},
		{"node":"172.16.6.252","name":"sysctl","status":"Fail","message":"fs.file-max = 790964, should be greater than 1000000"},
		{"node":"172.16.6.252","name":"sysctl","status":"Fail","message":"net.core.somaxconn = 128, should be greater than 32768"},
		{"node":"172.16.6.252","name":"sysctl","status":"Fail","message":"net.ipv4.tcp_syncookies = 1, should be 0"},
		{"node":"172.16.6.252","name":"sysctl","status":"Fail","message":"vm.swappiness = 30, should be 0"},
		{"node":"172.16.6.252","name":"selinux","status":"Pass","message":"SELinux is disabled"},
		{"node":"172.16.6.252","name":"thp","status":"Fail","message":"THP is enabled, please disable it for best performance"},
		{"node":"172.16.6.252","name":"command","status":"Fail","message":"numactl not usable, bash: numactl: command not found"},
		{"node":"172.16.5.168","name":"exist","status":"Fail","message":"/home/tiem already exists"},
		{"node":"172.16.5.168","name":"exist","status":"Fail","message":"/root already exists"},
		{"node":"172.16.5.168","name":"os-version","status":"Pass","message":"OS is CentOS Linux 7 (Core) 7.6.1810"},
		{"node":"172.16.5.168","name":"cpu-cores","status":"Pass","message":"number of CPU cores / threads: 8"},
		{"node":"172.16.5.168","name":"cpu-governor","status":"Warn","message":"Unable to determine current CPU frequency governor policy"},
		{"node":"172.16.5.168","name":"swap","status":"Fail","message":"swap is enabled, please disable it for best performance"},
		{"node":"172.16.5.168","name":"memory","status":"Pass","message":"memory size is 16384MB"},
		{"node":"172.16.5.168","name":"disk","status":"Warn","message":"mount point / does not have 'noatime' option set"},
		{"node":"172.16.5.168","name":"disk","status":"Warn","message":"mount point /home does not have 'noatime' option set"},
		{"node":"172.16.5.168","name":"selinux","status":"Pass","message":"SELinux is disabled"},
		{"node":"172.16.5.168","name":"thp","status":"Pass","message":"THP is disabled"},
		{"node":"172.16.5.168","name":"command","status":"Fail","message":"numactl not usable, bash: numactl: command not found"}]}`

	var results checkHostResults
	err := (&results).buildFromJson(jsonStr)
	assert.Nil(t, err)
	assert.Equal(t, 28, len(results.Result))

	sortedResult := results.analyzeCheckResults()
	assert.Equal(t, 5, len(*sortedResult["Warn"]))
	assert.Equal(t, 14, len(*sortedResult["Fail"]))
	assert.Equal(t, 9, len(*sortedResult["Pass"]))
}

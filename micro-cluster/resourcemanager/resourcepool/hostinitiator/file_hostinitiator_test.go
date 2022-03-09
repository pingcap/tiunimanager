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
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/deployment"
	"github.com/pingcap-inc/tiem/library/framework"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	mock_deployment "github.com/pingcap-inc/tiem/test/mockdeployment"
	mock_ssh "github.com/pingcap-inc/tiem/test/mockutil/mocksshclientexecutor"
	sshclient "github.com/pingcap-inc/tiem/util/ssh"
	"github.com/stretchr/testify/assert"
)

func genHostInfo(hostName string, purpose string) *structs.HostInfo {
	host := structs.HostInfo{
		IP:       "192.168.56.11",
		HostName: hostName,
		OS:       "Centos",
		Kernel:   "3.10",
		Region:   "TEST_REGION",
		AZ:       "TEST_AZ",
		Rack:     "TEST_RACK",
		Status:   string(constants.HostOnline),
		Nic:      "10GE",
		Purpose:  purpose,
	}
	host.Disks = append(host.Disks, structs.DiskInfo{
		Name:     "sda",
		Path:     "/",
		Status:   string(constants.DiskReserved),
		Capacity: 512,
	})
	host.Disks = append(host.Disks, structs.DiskInfo{
		Name:     "sdb",
		Path:     "/mnt/sdb",
		Status:   string(constants.DiskAvailable),
		Capacity: 1024,
	})
	host.Disks = append(host.Disks, structs.DiskInfo{
		Name:     "sdc",
		Path:     "/mnt/sdc",
		Status:   string(constants.DiskAvailable),
		Capacity: 1024,
	})
	return &host
}

func Test_AuthHost(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.AuthHost(context.TODO(), "tiem", "tiem", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"})
	// depend on whether user home dir has public key
	if err != nil {
		emErr, ok := err.(errors.EMError)
		assert.True(t, ok)
		assert.Equal(t, errors.TIEM_RESOURCE_INIT_HOST_AUTH_ERROR, emErr.GetCode())
	}
}

func Test_Verify_ignoreWarings(t *testing.T) {
	jsonStr := `{"result":[
		{"node":"172.16.6.252","name":"os-version","status":"Pass","message":"OS is CentOS Linux 7 (Core) 7.6.1810"},
		{"node":"172.16.6.252","name":"cpu-cores","status":"Pass","message":"number of CPU cores / threads: 4"},
		{"node":"172.16.6.252","name":"cpu-governor","status":"Warn","message":"Unable to determine current CPU frequency governor policy"},
		{"node":"172.16.6.252","name":"memory","status":"Pass","message":"memory size is 8192MB"},
		{"node":"172.16.6.252","name":"disk","status":"Warn","message":"mount point /home does not have 'noatime' option set"},
		{"node":"172.16.6.252","name":"selinux","status":"Pass","message":"SELinux is disabled"},
		{"node":"172.16.5.168","name":"os-version","status":"Pass","message":"OS is CentOS Linux 7 (Core) 7.6.1810"},
		{"node":"172.16.5.168","name":"cpu-cores","status":"Pass","message":"number of CPU cores / threads: 8"},
		{"node":"172.16.5.168","name":"cpu-governor","status":"Warn","message":"Unable to determine current CPU frequency governor policy"},
		{"node":"172.16.5.168","name":"memory","status":"Pass","message":"memory size is 16384MB"},
		{"node":"172.16.5.168","name":"disk","status":"Warn","message":"mount point / does not have 'noatime' option set"},
		{"node":"172.16.5.168","name":"disk","status":"Warn","message":"mount point /home does not have 'noatime' option set"},
		{"node":"172.16.5.168","name":"selinux","status":"Pass","message":"SELinux is disabled"},
		{"node":"172.16.5.168","name":"thp","status":"Pass","message":"THP is disabled"}]}`

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSec := mock_deployment.NewMockInterface(ctrl)
	mockSec.EXPECT().CheckConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(jsonStr, nil)
	deployment.M = mockSec

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetDeploymentServ(mockSec)

	ctx := context.WithValue(context.TODO(), rp_consts.ContextIgnoreWarnings, true)
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	err := fileInitiator.Verify(ctx, &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180"})
	assert.Nil(t, err)
}

func Test_Verify_Warings(t *testing.T) {
	jsonStr := `{"result":[
		{"node":"172.16.6.252","name":"os-version","status":"Pass","message":"OS is CentOS Linux 7 (Core) 7.6.1810"},
		{"node":"172.16.6.252","name":"cpu-cores","status":"Pass","message":"number of CPU cores / threads: 4"},
		{"node":"172.16.6.252","name":"cpu-governor","status":"Warn","message":"Unable to determine current CPU frequency governor policy"},
		{"node":"172.16.6.252","name":"memory","status":"Pass","message":"memory size is 8192MB"},
		{"node":"172.16.6.252","name":"disk","status":"Warn","message":"mount point /home does not have 'noatime' option set"},
		{"node":"172.16.6.252","name":"selinux","status":"Pass","message":"SELinux is disabled"},
		{"node":"172.16.5.168","name":"os-version","status":"Pass","message":"OS is CentOS Linux 7 (Core) 7.6.1810"},
		{"node":"172.16.5.168","name":"cpu-cores","status":"Pass","message":"number of CPU cores / threads: 8"},
		{"node":"172.16.5.168","name":"cpu-governor","status":"Warn","message":"Unable to determine current CPU frequency governor policy"},
		{"node":"172.16.5.168","name":"memory","status":"Pass","message":"memory size is 16384MB"},
		{"node":"172.16.5.168","name":"disk","status":"Warn","message":"mount point / does not have 'noatime' option set"},
		{"node":"172.16.5.168","name":"disk","status":"Warn","message":"mount point /home does not have 'noatime' option set"},
		{"node":"172.16.5.168","name":"selinux","status":"Pass","message":"SELinux is disabled"},
		{"node":"172.16.5.168","name":"thp","status":"Pass","message":"THP is disabled"}]}`

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSec := mock_deployment.NewMockInterface(ctrl)
	mockSec.EXPECT().CheckConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(jsonStr, nil)
	deployment.M = mockSec

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetDeploymentServ(mockSec)

	ctx := context.WithValue(context.TODO(), rp_consts.ContextIgnoreWarnings, false)
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	err := fileInitiator.Verify(ctx, &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180"})
	assert.NotNil(t, err)
	emErr, ok := err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIEM_RESOURCE_HOST_NOT_EXPECTED, emErr.GetCode())
}

func Test_Prepare_NoError(t *testing.T) {
	jsonStr := `{"result":[
		{"node":"172.16.6.252","name":"os-version","status":"Pass","message":"OS is CentOS Linux 7 (Core) 7.6.1810"},
		{"node":"172.16.6.252","name":"cpu-cores","status":"Pass","message":"number of CPU cores / threads: 4"},
		{"node":"172.16.6.252","name":"cpu-governor","status":"Warn","message":"Unable to determine current CPU frequency governor policy"}]}
		`
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSec := mock_deployment.NewMockInterface(ctrl)
	mockSec.EXPECT().CheckConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(jsonStr, nil)
	deployment.M = mockSec

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetDeploymentServ(mockSec)

	err := fileInitiator.Prepare(context.TODO(), &structs.HostInfo{IP: "666.666.66.66", UserName: "r00t", Passwd: "fake"})
	assert.Nil(t, err)
}

func Test_Prepare_ConnectError(t *testing.T) {
	jsonStr := `{"result":[
		{"node":"172.16.6.252","name":"os-version","status":"Pass","message":"OS is CentOS Linux 7 (Core) 7.6.1810"},
		{"node":"172.16.6.252","name":"cpu-cores","status":"Pass","message":"number of CPU cores / threads: 4"},
		{"node":"172.16.6.252","name":"cpu-governor","status":"Warn","message":"Unable to determine current CPU frequency governor policy"},
		{"node":"172.16.6.252","name":"swap","status":"Fail","message":"swap is enabled, please disable it for best performance"}]}
		`
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSec := mock_deployment.NewMockInterface(ctrl)
	mockSec.EXPECT().CheckConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(jsonStr, nil)
	deployment.M = mockSec

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetDeploymentServ(mockSec)

	err := fileInitiator.Prepare(context.TODO(), &structs.HostInfo{IP: "666.666.66.66", UserName: "r00t", Passwd: "fake"})
	assert.NotNil(t, err)
	emErr, ok := err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIEM_RESOURCE_PREPARE_HOST_ERROR, emErr.GetCode())
}

func Test_SetOffSwap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.setOffSwap(context.TODO(), &structs.HostInfo{IP: "666.666.666.666", UserName: "r00t", Passwd: "fake"})
	assert.Nil(t, err)
}

func Test_installNumaCtl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.installNumaCtl(context.TODO(), &structs.HostInfo{IP: "666.666.666.666", UserName: "r00t", Passwd: "fake"})
	assert.Nil(t, err)
}

func Test_Remount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(host string, port int, authenticate sshclient.HostAuthenticate, sudo bool, timeoutS int, commands []string) (string, error) {
			if strings.HasPrefix(commands[0], "sed -n") {
				return "/dev/mapper/centos-root /data    xfs     defaults        0 0", nil
			} else if strings.HasPrefix(commands[0], "sed -i") {
				fields := strings.Split(commands[0], "#")
				assert.Equal(t, "defaults,nodelalloc,noatime", fields[4])
				return "", nil
			}
			return "", nil
		}).Times(2)

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.remountFS(context.TODO(), &structs.HostInfo{IP: "666.666.666.666", UserName: "r00t", Passwd: "fake"}, "/data", []string{"nodelalloc", "noatime"})
	assert.Nil(t, err)

}

func Test_GenerateTopologyConfig(t *testing.T) {
	template_struct := templateScaleOut{}
	template_struct.HostAddrs = append(template_struct.HostAddrs, HostAddr{"192.168.177.177", 10086})
	template_struct.HostAddrs = append(template_struct.HostAddrs, HostAddr{"192.168.177.178", 10087})
	template_struct.HostAddrs = append(template_struct.HostAddrs, HostAddr{"192.168.177.179", 10088})

	str, err := template_struct.generateTopologyConfig(context.TODO())
	assert.Nil(t, err)
	t.Log(str)
}

func Test_InstallSoftware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	err := fileInitiator.InstallSoftware(context.TODO(), []structs.HostInfo{{Arch: "X86_64", IP: "192.168.177.180"}})
	assert.Nil(t, err)
}

func Test_PreCheckHostInstallFilebeat(t *testing.T) {
	jsonStr := `
	{
		"cluster_meta": {
		  "cluster_type": "tiem",
		  "cluster_name": "tiem-test",
		  "cluster_version": "v1.0.0-beta.7",
		  "deploy_user": "tiem",
		  "ssh_type": "builtin"
		},
		"instances": [
		  {
			"id": "172.16.6.252:4112",
			"role": "alertmanager",
			"host": "172.16.6.252",
			"ports": "4112/4113",
			"os_arch": "linux/x86_64",
			"status": "Up",
			"since": "-",
			"data_dir": "/em-data/alertmanager-4112",
			"deploy_dir": "/em-deploy/alertmanager-4112",
			"ComponentName": "alertmanager",
			"Port": 4112
		  },
		  {
			"id": "172.16.6.252:0",
			"role": "filebeat",
			"host": "172.16.6.252",
			"ports": "",
			"os_arch": "linux/x86_64",
			"status": "Up",
			"since": "-",
			"data_dir": "/em-data/filebeat-0",
			"deploy_dir": "/em-deploy/filebeat-0",
			"ComponentName": "filebeat",
			"Port": 0
		  }
		]
	}
	`
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSec := mock_deployment.NewMockInterface(ctrl)
	mockSec.EXPECT().Display(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(jsonStr, nil).Times(2)
	framework.InitBaseFrameworkForUt(framework.ClusterService)

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetDeploymentServ(mockSec)

	installed, err := fileInitiator.PreCheckHostInstallFilebeat(context.TODO(), []structs.HostInfo{{Arch: "X86_64", IP: "172.16.6.252"}})
	assert.Nil(t, err)
	assert.True(t, installed)

	installed, err = fileInitiator.PreCheckHostInstallFilebeat(context.TODO(), []structs.HostInfo{{Arch: "X86_64", IP: "172.16.6.253"}})
	assert.Nil(t, err)
	assert.False(t, installed)
}

func Test_JoinEMCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSec := mock_deployment.NewMockInterface(ctrl)
	mockSec.EXPECT().ScaleOut(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	deployment.M = mockSec

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetDeploymentServ(mockSec)

	ctx := context.WithValue(context.TODO(), rp_consts.ContextWorkFlowIDKey, "fake-node-id")
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	_, err := fileInitiator.JoinEMCluster(ctx, []structs.HostInfo{{Arch: "X86_64", IP: "192.168.177.180"}})
	assert.Nil(t, err)
}

func Test_LeaveEMCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSec := mock_deployment.NewMockInterface(ctrl)
	mockSec.EXPECT().ScaleIn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	deployment.M = mockSec

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetDeploymentServ(mockSec)

	ctx := context.WithValue(context.TODO(), rp_consts.ContextWorkFlowIDKey, "fake-node-id")
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	_, err := fileInitiator.LeaveEMCluster(ctx, "192.168.177.180:0")
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
		{"node":"172.16.5.168","name":"exist","status":"Fail","message":"/home/tidb already exists"},
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

func Test_BuildCheckHostTemplateItems(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	host := genHostInfo("Test_Host1", "Compute,Storage,Schedule")
	templateInfo := templateCheckHost{}
	host.SSHPort = 10086
	(&templateInfo).buildCheckHostTemplateItems(host)

	assert.Equal(t, 2, len(templateInfo.TemplateItemsForCompute))
	assert.Equal(t, 2, len(templateInfo.TemplateItemsForStorage))
	assert.Equal(t, 2, len(templateInfo.TemplateItemsForSchedule))
	assert.Equal(t, 10086, templateInfo.GlobalSSHPort)

	t.Log(templateInfo)

	assert.Equal(t, 10000, templateInfo.TemplateItemsForCompute[0].Port1)
	assert.Equal(t, 11000, templateInfo.TemplateItemsForCompute[0].Port2)
	assert.Equal(t, 10001, templateInfo.TemplateItemsForCompute[1].Port1)
	assert.Equal(t, 11001, templateInfo.TemplateItemsForCompute[1].Port2)
	assert.Equal(t, "/mnt/sdb", templateInfo.TemplateItemsForCompute[0].DeployDir)
	assert.Equal(t, "", templateInfo.TemplateItemsForCompute[0].DataDir)
	assert.Equal(t, "/mnt/sdc", templateInfo.TemplateItemsForCompute[1].DeployDir)
	assert.Equal(t, "", templateInfo.TemplateItemsForCompute[1].DataDir)
	assert.Equal(t, "/mnt/sdc", templateInfo.TemplateItemsForStorage[1].DataDir)

	str, err := templateInfo.generateTopologyConfig(context.TODO())
	assert.Nil(t, err)
	t.Log(str)
}

func Test_GetRemountInfoFromMsg(t *testing.T) {
	fileInitiator := NewFileHostInitiator()
	message1 := "mount point /data does not have 'nodelalloc' option set, auto fixing not supported"
	mountPoint1, opt1, err := fileInitiator.getRemountInfoFromMsg(context.TODO(), message1)
	assert.Nil(t, err)
	assert.Equal(t, "/data", mountPoint1)
	assert.Equal(t, "nodelalloc", opt1)

	message2 := "mount point /data does not have 'noatime' option set, auto fixing not supported"
	mountPoint2, opt2, err := fileInitiator.getRemountInfoFromMsg(context.TODO(), message2)
	assert.Nil(t, err)
	assert.Equal(t, "/data", mountPoint2)
	assert.Equal(t, "noatime", opt2)

	message3 := "mount point /data does not have 'noatime', auto fixing not supported"
	_, _, err = fileInitiator.getRemountInfoFromMsg(context.TODO(), message3)
	assert.NotNil(t, err)
	emErr, ok := err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIEM_RESOURCE_PREPARE_HOST_ERROR, emErr.GetCode())

	message4 := "mount point /data does not have 'cached', auto fixing not supported"
	_, _, err = fileInitiator.getRemountInfoFromMsg(context.TODO(), message4)
	assert.NotNil(t, err)
	emErr, ok = err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIEM_RESOURCE_PREPARE_HOST_ERROR, emErr.GetCode())
}

func Test_AddRemountOpts(t *testing.T) {
	fileInitiator := NewFileHostInitiator()
	remount := map[string]map[string]struct{}{}
	fileInitiator.addRemountOpts(remount, "/data1", "noatime")
	fileInitiator.addRemountOpts(remount, "/data2", "nodelalloc")
	opts1, ok := remount["/data1"]
	assert.True(t, ok)
	_, ok = opts1["noatime"]
	assert.True(t, ok)
	opts2, ok := remount["/data2"]
	assert.True(t, ok)
	_, ok = opts2["nodelalloc"]
	assert.True(t, ok)
}

func Test_isVirtualMachine_True(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("QEMU", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	isVM, err := fileInitiator.isVirtualMachine(context.TODO(), &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180"})
	assert.Nil(t, err)
	assert.True(t, isVM)
}

func Test_isVirtualMachine_Contains(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("VMware, Inc.", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	isVM, err := fileInitiator.isVirtualMachine(context.TODO(), &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180"})
	assert.Nil(t, err)
	assert.True(t, isVM)
}

func Test_isVirtualMachine_False(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("Cisco Systems Inc", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	isVM, err := fileInitiator.isVirtualMachine(context.TODO(), &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180"})
	assert.Nil(t, err)
	assert.False(t, isVM)
}

func Test_passCpuGovernorWarn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("QEMU", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)
	// "node":"172.16.6.252","name":"cpu-governor","status":"Warn","message":""
	warnings := []checkHostResult{
		{Node: "172.16.6.252", Name: "cpu-governor", Status: "Warn", Message: "Unable to determine current CPU frequency governor policy"},
	}
	ok, err := fileInitiator.passCpuGovernorWarn(context.TODO(), &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180"}, &warnings)
	assert.Nil(t, err)
	assert.True(t, ok)
}

func Test_skipAuthHost_Skip(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	skip := fileInitiator.skipAuthHost(context.TODO(), "root", &structs.HostInfo{IP: "666.666.66.66"})
	assert.True(t, skip)
}

func Test_skipAuthHost_NotSkip(t *testing.T) {
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.NewErrorf(errors.TIEM_RESOURCE_CONNECT_TO_HOST_ERROR, "bad host"))

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	// specified user is 'root', deploy user is 'test-user'
	skip := fileInitiator.skipAuthHost(context.TODO(), "test-user", &structs.HostInfo{IP: "666.666.66.66"})
	assert.False(t, skip)

	skip = fileInitiator.skipAuthHost(context.TODO(), "root", &structs.HostInfo{IP: "666.666.66.66"})
	assert.False(t, skip)

	framework.Current.GetClientArgs().LoginPrivateKeyPath = ""
	framework.Current.GetClientArgs().LoginPublicKeyPath = ""
	skip = fileInitiator.skipAuthHost(context.TODO(), "root", &structs.HostInfo{IP: "666.666.66.66"})
	assert.False(t, skip)
}

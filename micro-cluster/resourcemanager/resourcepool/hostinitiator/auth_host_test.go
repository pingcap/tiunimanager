/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	mock_ssh "github.com/pingcap/tiunimanager/test/mockutil/mocksshclientexecutor"
	sshclient "github.com/pingcap/tiunimanager/util/ssh"
	"github.com/stretchr/testify/assert"
)

func Test_createDeployUser_succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.createDeployUser(context.TODO(), "tiunimanager", "tiunimanager", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"}, &sshclient.HostAuthenticate{})
	assert.Nil(t, err)
}

func Test_createDeployUser_args_error(t *testing.T) {
	fileInitiator := NewFileHostInitiator()

	err := fileInitiator.createDeployUser(context.TODO(), "", "", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"}, &sshclient.HostAuthenticate{})
	assert.NotNil(t, err)
	emErr, ok := err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIUNIMANAGER_RESOURCE_INIT_DEPLOY_USER_ERROR, emErr.GetCode())
	assert.Equal(t, "deployUser and group should not be null", emErr.GetMsg())

	err = fileInitiator.createDeployUser(context.TODO(), "tiunimanager", "tiunimanager", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"}, nil)
	assert.NotNil(t, err)
	emErr, ok = err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIUNIMANAGER_RESOURCE_INIT_DEPLOY_USER_ERROR, emErr.GetCode())
	assert.Equal(t, "authenticate info should not be nil while creating deploy user", emErr.GetMsg())
}

func Test_createDeployUser_failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", errors.NewError(errors.TIUNIMANAGER_RESOURCE_CONNECT_TO_HOST_ERROR, "")).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.createDeployUser(context.TODO(), "tiunimanager", "tiunimanager", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"}, &sshclient.HostAuthenticate{})
	assert.NotNil(t, err)
	emErr, ok := err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIUNIMANAGER_RESOURCE_INIT_DEPLOY_USER_ERROR, emErr.GetCode())

}

func Test_buildUserCommand_addUser(t *testing.T) {
	um := UserModuleConfig{
		Action: userActionAdd,
		Name:   "tiunimanager",
		Group:  "tiunimanager",
		Sudoer: true,
	}
	expectedCmd := "id -u tiunimanager > /dev/null 2>&1 || (/usr/sbin/groupadd -f tiunimanager && /usr/sbin/useradd -m -s /bin/bash -g tiunimanager tiunimanager) && echo 'tiunimanager ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/tiunimanager"
	createUserCmd := um.buildUserCommand()
	assert.Equal(t, expectedCmd, createUserCmd)
}

func Test_buildUserCommand_delUser(t *testing.T) {
	um := UserModuleConfig{
		Action: userActionDel,
		Name:   "tiunimanager",
		Group:  "tiunimanager",
		Sudoer: true,
	}
	expectedCmd := "/usr/sbin/userdel -r tiunimanager || [ $? -eq 6 ]"
	deleteUserCmd := um.buildUserCommand()
	assert.Equal(t, expectedCmd, deleteUserCmd)
}

func Test_appendAuthorizedKeysFile_succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.appendRemoteAuthorizedKeysFile(context.TODO(), nil, "tiunimanager", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"}, &sshclient.HostAuthenticate{})
	assert.Nil(t, err)
}

func Test_BuildAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.buildAuth(context.TODO(), "tiunimanager", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"}, &sshclient.HostAuthenticate{})
	// depend on whether user home dir has public key
	if err != nil {
		emErr, ok := err.(errors.EMError)
		assert.True(t, ok)
		assert.Equal(t, errors.TIUNIMANAGER_RESOURCE_INIT_HOST_AUTH_ERROR, emErr.GetCode())
	}
}

func Test_findSSHAuthorizedKeysFile(t *testing.T) {
	sshConfig := `
	HostKey /etc/ssh/ssh_host_rsa_key
	HostKey /etc/ssh/ssh_host_ecdsa_key
	HostKey /etc/ssh/ssh_host_ed25519_key
	SyslogFacility AUTHPRIV
	AuthorizedKeysFile	.ssh/authorized_keys_test
	PasswordAuthentication yes
	ChallengeResponseAuthentication no
	GSSAPIAuthentication yes
	GSSAPICleanupCredentials no
	UsePAM yes
	X11Forwarding yes
	UseDNS no
	AcceptEnv LANG LC_CTYPE LC_NUMERIC LC_TIME LC_COLLATE LC_MONETARY LC_MESSAGES
	AcceptEnv LC_PAPER LC_NAME LC_ADDRESS LC_TELEPHONE LC_MEASUREMENT
	AcceptEnv LC_IDENTIFICATION LC_ALL LANGUAGE
	AcceptEnv XMODIFIERS
	Subsystem	sftp	/usr/libexec/openssh/sftp-server`
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sshConfig, nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	keyFile := fileInitiator.findSSHAuthorizedKeysFile(context.TODO(), &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"}, &sshclient.HostAuthenticate{})
	assert.Equal(t, "~/.ssh/authorized_keys_test", keyFile)
}

func Test_getUserSpecifiedAuthenticateToHost(t *testing.T) {
	fileInitiator := NewFileHostInitiator()
	authenticate, err := fileInitiator.getUserSpecifiedAuthenticateToHost(context.TODO(), &structs.HostInfo{UserName: "test", Passwd: "password"})
	assert.Nil(t, err)
	assert.Equal(t, sshclient.Passwd, authenticate.SshType)
	assert.Equal(t, "test", authenticate.AuthenticatedUser)
	assert.Equal(t, "password", authenticate.AuthenticateContent)

	framework.InitBaseFrameworkForUt(framework.ClusterService)
	authenticate, err = fileInitiator.getUserSpecifiedAuthenticateToHost(context.TODO(), &structs.HostInfo{UserName: "test"})
	assert.Nil(t, err)
	assert.Equal(t, sshclient.Key, authenticate.SshType)
	assert.Equal(t, "root", authenticate.AuthenticatedUser)
	assert.Equal(t, "/fake/private/key/path", authenticate.AuthenticateContent)

	framework.Current.GetClientArgs().LoginHostUser = ""
	_, err = fileInitiator.getUserSpecifiedAuthenticateToHost(context.TODO(), &structs.HostInfo{UserName: "test"})
	assert.NotNil(t, err)

}

func Test_getEMAuthenticateToHost(t *testing.T) {
	fileInitiator := NewFileHostInitiator()
	framework.InitBaseFrameworkForUt(framework.ClusterService)

	authenticate := fileInitiator.getEMAuthenticateToHost(context.TODO())
	assert.Equal(t, sshclient.Key, authenticate.SshType)
	assert.Equal(t, "test-user", authenticate.AuthenticatedUser)
	assert.Equal(t, "/home/test-user/.ssh/tiup_rsa", authenticate.AuthenticateContent)
}

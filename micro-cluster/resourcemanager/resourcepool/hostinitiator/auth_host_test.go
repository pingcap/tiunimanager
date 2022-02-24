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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	mock_ssh "github.com/pingcap-inc/tiem/test/mockutil/mocksshclientexecutor"
	"github.com/stretchr/testify/assert"
)

func Test_createDeployUser_succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.createDeployUser(context.TODO(), "tiem", "tiem", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"})
	assert.Nil(t, err)
}

func Test_createDeployUser_no_user_failed(t *testing.T) {
	fileInitiator := NewFileHostInitiator()

	err := fileInitiator.createDeployUser(context.TODO(), "", "", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"})
	assert.NotNil(t, err)
	emErr, ok := err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIEM_RESOURCE_INIT_DEPLOY_USER_ERROR, emErr.GetCode())
	assert.Equal(t, "deployUser and group should not be null", emErr.GetMsg())
}

func Test_createDeployUser_failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", errors.NewError(errors.TIEM_RESOURCE_CONNECT_TO_HOST_ERROR, "")).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.createDeployUser(context.TODO(), "tiem", "tiem", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"})
	assert.NotNil(t, err)
	emErr, ok := err.(errors.EMError)
	assert.True(t, ok)
	assert.Equal(t, errors.TIEM_RESOURCE_INIT_DEPLOY_USER_ERROR, emErr.GetCode())

}

func Test_buildUserCommand_addUser(t *testing.T) {
	um := UserModuleConfig{
		Action: userActionAdd,
		Name:   "tiem",
		Group:  "tiem",
		Sudoer: true,
	}
	expectedCmd := "id -u tiem > /dev/null 2>&1 || (/usr/sbin/groupadd -f tiem && /usr/sbin/useradd -m -s /bin/bash -g tiem tiem) && echo 'tiem ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/tiem"
	createUserCmd := um.buildUserCommand()
	assert.Equal(t, expectedCmd, createUserCmd)
}

func Test_buildUserCommand_delUser(t *testing.T) {
	um := UserModuleConfig{
		Action: userActionDel,
		Name:   "tiem",
		Group:  "tiem",
		Sudoer: true,
	}
	expectedCmd := "/usr/sbin/userdel -r tiem || [ $? -eq 6 ]"
	deleteUserCmd := um.buildUserCommand()
	assert.Equal(t, expectedCmd, deleteUserCmd)
}

func Test_appendAuthorizedKeysFile_succeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.appendRemoteAuthorizedKeysFile(context.TODO(), nil, "tiem", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"})
	assert.Nil(t, err)
}

func Test_BuildAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	err := fileInitiator.buildAuth(context.TODO(), "tiem", &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"})
	// depend on whether user home dir has public key
	if err != nil {
		emErr, ok := err.(errors.EMError)
		assert.True(t, ok)
		assert.Equal(t, errors.TIEM_RESOURCE_INIT_HOST_AUTH_ERROR, emErr.GetCode())
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
	mockClient.EXPECT().RunCommandsInRemoteHost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sshConfig, nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	fileInitiator.SetSSHClient(mockClient)

	keyFile := fileInitiator.findSSHAuthorizedKeysFile(context.TODO(), &structs.HostInfo{Arch: "X86_64", IP: "192.168.177.180", UserName: "fakeUser", Passwd: "fakePasswd"})
	assert.Equal(t, "~/.ssh/authorized_keys_test", keyFile)
}

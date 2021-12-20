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
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	mock_ssh "github.com/pingcap-inc/tiem/test/mockutil/mocksshclientexecutor"
	"github.com/stretchr/testify/assert"
)

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
		return "", framework.NewTiEMErrorf(common.TIEM_PARAMETER_INVALID, "BadRequest")
	}).Times(3)

	fileInitiator := NewFileHostInitiator()
	err := fileInitiator.VerifyCpuMem(context.TODO(), mockClient, &structs.HostInfo{Arch: "x86_64", CpuCores: 32, Memory: 8})
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
		return "", framework.NewTiEMErrorf(common.TIEM_PARAMETER_INVALID, "BadRequest")
	}).Times(1)

	fileInitiator := NewFileHostInitiator()
	err := fileInitiator.VerifyCpuMem(context.TODO(), mockClient, &structs.HostInfo{Arch: "x86_64", CpuCores: 32, Memory: 8})
	assert.NotNil(t, err)
}

func Test_VerifyDisks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	err := fileInitiator.VerifyDisks(context.TODO(), mockClient, &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_VerifyFS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	err := fileInitiator.VerifyFS(context.TODO(), mockClient, &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_VerifySwap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	err := fileInitiator.VerifySwap(context.TODO(), mockClient, &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_VerifyEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	err := fileInitiator.VerifyEnv(context.TODO(), mockClient, &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_VerifyOSEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	err := fileInitiator.VerifyOSEnv(context.TODO(), mockClient, &structs.HostInfo{})
	assert.Nil(t, err)
}

func Test_SetOffSwap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_ssh.NewMockSSHClientExecutor(ctrl)
	mockClient.EXPECT().RunCommandsInSession(gomock.Any()).Return("", nil).AnyTimes()

	fileInitiator := NewFileHostInitiator()
	err := fileInitiator.SetOffSwap(context.TODO(), mockClient, &structs.HostInfo{})
	assert.Nil(t, err)
}

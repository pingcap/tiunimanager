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

package resourcepool

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	rp_consts "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/constants"
	"github.com/pingcap-inc/tiem/models"
	workflowModel "github.com/pingcap-inc/tiem/models/workflow"
	mock_initiator "github.com/pingcap-inc/tiem/test/mockresource/mockinitiator"
	mock_provider "github.com/pingcap-inc/tiem/test/mockresource/mockprovider"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
)

func Test_InstallSoftware(t *testing.T) {
	models.MockDB()
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	resourcePool := GetResourcePool()

	// Mock host initiator
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInitiator := mock_initiator.NewMockHostInitiator(ctrl)
	mockInitiator.EXPECT().InstallSoftware(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hosts []structs.HostInfo) error {
		return nil
	})

	resourcePool.SetHostInitiator(mockInitiator)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(rp_consts.ContextResourcePoolKey, resourcePool)
	flowContext.SetData(rp_consts.ContextImportHostInfoKey, []structs.HostInfo{{IP: "192.168.192.192"}})

	var node workflowModel.WorkFlowNode
	err := installSoftware(&node, flowContext)
	assert.Nil(t, err)
}

func Test_JoinEMCluster(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host initiator
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInitiator := mock_initiator.NewMockHostInitiator(ctrl)
	mockInitiator.EXPECT().JoinEMCluster(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hosts []structs.HostInfo) error {
		return nil
	})

	resourcePool.SetHostInitiator(mockInitiator)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(rp_consts.ContextResourcePoolKey, resourcePool)
	flowContext.SetData(rp_consts.ContextImportHostInfoKey, []structs.HostInfo{{IP: "192.168.192.192"}})

	var node workflowModel.WorkFlowNode
	err := joinEmCluster(&node, flowContext)
	assert.Nil(t, err)
}

func Test_Verify(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host initiator
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInitiator := mock_initiator.NewMockHostInitiator(ctrl)
	mockInitiator.EXPECT().Verify(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, h *structs.HostInfo) error {
		return nil
	})

	resourcePool.SetHostInitiator(mockInitiator)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(rp_consts.ContextResourcePoolKey, resourcePool)
	flowContext.SetData(rp_consts.ContextImportHostInfoKey, []structs.HostInfo{{IP: "192.168.192.192"}})

	var node workflowModel.WorkFlowNode
	err := verifyHosts(&node, flowContext)
	assert.Nil(t, err)
}

func Test_ImportHostSucceed(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host initiator
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl)
	mockProvider.EXPECT().UpdateHostStatus(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, hostId []string, status string) error {
		return nil
	})

	resourcePool.SetHostProvider(mockProvider)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(rp_consts.ContextResourcePoolKey, resourcePool)
	flowContext.SetData(rp_consts.ContextImportHostIDsKey, []string{"fake-host-id"})

	var node workflowModel.WorkFlowNode
	err := importHostSucceed(&node, flowContext)
	assert.Nil(t, err)
}

func Test_ImportHostFail(t *testing.T) {
	models.MockDB()
	resourcePool := GetResourcePool()

	// Mock host initiator
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProvider := mock_provider.NewMockHostProvider(ctrl)
	mockProvider.EXPECT().UpdateHostStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	resourcePool.SetHostProvider(mockProvider)

	flowContext := workflow.NewFlowContext(context.TODO())
	flowContext.SetData(rp_consts.ContextResourcePoolKey, resourcePool)
	flowContext.SetData(rp_consts.ContextImportHostIDsKey, []string{"fake-host-id"})

	var node workflowModel.WorkFlowNode
	err := importHostsFail(&node, flowContext)
	assert.Nil(t, err)
}

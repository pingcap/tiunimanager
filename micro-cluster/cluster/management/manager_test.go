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

package management

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockworkflow"
	"github.com/pingcap-inc/tiem/workflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

var emptyNode = func(task *wfModel.WorkFlowNode, context *workflow.FlowContext) error {
	return nil
}

func getEmptyFlow(name string) *workflow.WorkFlowDefine {
	return &workflow.WorkFlowDefine{
		FlowName: name,
		TaskNodes: map[string]*workflow.NodeDefine{
			"start": {"start", "done", "fail", workflow.SyncFuncNode, emptyNode},
			"done":  {"end", "", "", workflow.SyncFuncNode, emptyNode},
			"fail":  {"fail", "", "", workflow.SyncFuncNode, emptyNode},
		},
	}
}

func TestManager_CreateCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowCreateCluster, getEmptyFlow(constants.FlowCreateCluster))
	manager := Manager{}
	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil)

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		flowRW := mockworkflow.NewMockReaderWriter(ctrl)
		flowRW.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any()).Return(nil, nil)
		flowRW.EXPECT().CreateWorkFlowNode(gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)
		flowRW.EXPECT().UpdateWorkFlowDetail(gomock.Any(), gomock.Any(), gomock.Any()).Return( nil).Times(3)
		models.SetWorkFlowReaderWriter(flowRW)

		_, err := manager.CreateCluster(context.TODO(), cluster.CreateClusterReq {
			ResourceParameter: structs.ClusterResourceInfo{
				InstanceResource: []structs.ClusterResourceParameterCompute{
					{Type: "TiDB", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "TiKV", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
					{Type: "PD", Count: 1, Resource: []structs.ClusterResourceParameterComputeResource{
						{Zone: "Test_Zone1", DiskType: "SATA", DiskCapacity: 0, Spec: "4C8G", Count: 1},
					}},
				},
			},
		})
		assert.NoError(t, err)
	})
	t.Run("no computes", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil)

		_, err := manager.CreateCluster(context.TODO(), cluster.CreateClusterReq {
		})
		assert.Error(t, err)
	})
}

func TestManager_StopCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowStopCluster, getEmptyFlow(constants.FlowStopCluster))
	manager := Manager{}
	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, nil)

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		flowRW := mockworkflow.NewMockReaderWriter(ctrl)
		flowRW.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any()).Return(nil, nil)
		flowRW.EXPECT().CreateWorkFlowNode(gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)
		flowRW.EXPECT().UpdateWorkFlowDetail(gomock.Any(), gomock.Any(), gomock.Any()).Return( nil).Times(3)
		models.SetWorkFlowReaderWriter(flowRW)

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq {
			ClusterID: "111",
		})
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(nil, nil, errors.New(""))

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq {
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
	t.Run("status", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, nil)
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(""))

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq {
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
}

func TestManager_RestartCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowRestartCluster, getEmptyFlow(constants.FlowRestartCluster))
	manager := Manager{}
	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, nil)

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		flowRW := mockworkflow.NewMockReaderWriter(ctrl)
		flowRW.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any()).Return(nil, nil)
		flowRW.EXPECT().CreateWorkFlowNode(gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)
		flowRW.EXPECT().UpdateWorkFlowDetail(gomock.Any(), gomock.Any(), gomock.Any()).Return( nil).Times(3)
		models.SetWorkFlowReaderWriter(flowRW)

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq {
			ClusterID: "111",
		})
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(nil, nil, errors.New(""))

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq {
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
	t.Run("status", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, nil)
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(""))

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq {
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
}

func TestManager_DeleteCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workflow.GetWorkFlowService().RegisterWorkFlow(context.TODO(), constants.FlowDeleteCluster, getEmptyFlow(constants.FlowDeleteCluster))
	manager := Manager{}
	t.Run("normal", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)

		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, nil)

		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		flowRW := mockworkflow.NewMockReaderWriter(ctrl)
		flowRW.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any()).Return(nil, nil)
		flowRW.EXPECT().CreateWorkFlowNode(gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)
		flowRW.EXPECT().UpdateWorkFlowDetail(gomock.Any(), gomock.Any(), gomock.Any()).Return( nil).Times(3)
		models.SetWorkFlowReaderWriter(flowRW)

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq {
			ClusterID: "111",
		})
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(nil, nil, errors.New(""))

		_, err := manager.StopCluster(context.TODO(), cluster.StopClusterReq {
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
	t.Run("status", func(t *testing.T) {
		clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
		models.SetClusterReaderWriter(clusterRW)
		clusterRW.EXPECT().GetMeta(gomock.Any(), "111").Return(&management.Cluster{}, []*management.ClusterInstance{
			{},
			{},
		}, nil)
		clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(""))

		_, err := manager.DeleteCluster(context.TODO(), cluster.DeleteClusterReq {
			ClusterID: "111",
		})
		assert.Error(t, err)
	})
}

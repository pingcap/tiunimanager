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

/*******************************************************************************
 * @File: switchover_test.go
 * @Description: ut of switchover
 * @Author: hansen@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/15 11:30
*******************************************************************************/

package switchover

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	clusterMgr "github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/test/mockcdcmanager"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	mock_workflow_service "github.com/pingcap-inc/tiem/test/mockworkflow"
	workflow "github.com/pingcap-inc/tiem/workflow2"
	"github.com/stretchr/testify/assert"
)

func init() {
	models.MockDB()
}

func TestSwitchover(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(
		&management.Cluster{
			Entity: common.Entity{
				ID:       "id-xxxx",
				TenantId: "tid-xxx",
			},
		},
		[]*management.ClusterInstance{
			{
				Entity: common.Entity{
					ID:     "111111",
					Status: string(constants.ClusterRunning),
				},
				Type:   "TiDB",
				HostIP: []string{"127.0.0.1"},
				Ports:  []int32{111},
			},
			{
				Entity: common.Entity{
					ID:     "222222",
					Status: string(constants.ClusterRunning),
				},
				Type:   "TiDB",
				HostIP: []string{"127.0.0.1"},
				Ports:  []int32{111},
			},
			{
				Entity: common.Entity{
					ID:     "333333",
					Status: string(constants.ClusterRunning),
				},
				Type:   "TiKV",
				HostIP: []string{"127.0.0.1"},
				Ports:  []int32{111},
			},
			{
				Entity: common.Entity{
					ID:     "333333",
					Status: string(constants.ClusterRunning),
				},
				Type:   "CDC",
				HostIP: []string{"127.0.0.1"},
				Ports:  []int32{111},
			},
		}, make([]*management.DBUser, 0), nil).AnyTimes()
	clusterRW.EXPECT().GetRelations(gomock.Any(), gomock.Eq("2")).Return(
		[]*clusterMgr.ClusterRelation{
			{
				RelationType:         constants.ClusterRelationStandBy,
				SubjectClusterID:     "1",
				ObjectClusterID:      "2",
				SyncChangeFeedTaskID: "1",
			},
		}, nil,
	).AnyTimes()
	clusterRW.EXPECT().GetSlaves(gomock.Any(), gomock.Eq("1")).Return(
		[]*clusterMgr.ClusterRelation{
			{
				RelationType:         constants.ClusterRelationStandBy,
				SubjectClusterID:     "1",
				ObjectClusterID:      "2",
				SyncChangeFeedTaskID: "1",
			},
		}, nil,
	).AnyTimes()
	clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	workflowService := mock_workflow_service.NewMockWorkFlowService(ctrl)
	workflow.MockWorkFlowService(workflowService)
	defer workflow.MockWorkFlowService(workflow.NewWorkFlowManager())
	workflowService.EXPECT().RegisterWorkFlow(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("flow01", nil).AnyTimes()
	workflowService.EXPECT().InitContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().Start(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	cdcAPI := mockcdcmanager.NewMockCDCManagerAPI(ctrl)
	cdcAPI.EXPECT().Create(gomock.Any(), gomock.Any()).Return(cluster.CreateChangeFeedTaskResp{}, nil).AnyTimes()

	service := GetManager()
	resp, err := service.Switchover(context.TODO(), &cluster.MasterSlaveClusterSwitchoverReq{
		// old master
		SourceClusterID: "1",
		// old slave
		TargetClusterID: "2",
		Force:           false,
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp.WorkFlowID)
}

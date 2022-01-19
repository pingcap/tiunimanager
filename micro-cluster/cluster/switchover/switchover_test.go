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
	wfModel "github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/test/mockcdcmanager"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	mock_workflow_service "github.com/pingcap-inc/tiem/test/mockworkflow"
	"github.com/pingcap-inc/tiem/workflow"
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
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:       "id-xxxx",
			TenantId: "tid-xxx",
		},
	}, make([]*management.ClusterInstance, 0), nil).AnyTimes()
	clusterRW.EXPECT().GetRelations(gomock.Any(), gomock.Eq("2")).Return(
		[]*clusterMgr.ClusterRelation{
			{
				RelationType:         constants.ClusterRelationSlaveTo,
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
	workflowService.EXPECT().CreateWorkFlow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&workflow.WorkFlowAggregation{
		Flow:    &wfModel.WorkFlow{Entity: common.Entity{ID: "flow01"}},
		Context: workflow.FlowContext{Context: context.TODO(), FlowData: make(map[string]interface{})},
	}, nil).AnyTimes()
	workflowService.EXPECT().AddContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	workflowService.EXPECT().AsyncStart(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	cdcAPI := mockcdcmanager.NewMockCDCManagerAPI(ctrl)
	cdcAPI.EXPECT().Create(gomock.Any(), gomock.Any()).Return(cluster.CreateChangeFeedTaskResp{}, nil).AnyTimes()

	service := GetManager(cdcAPI)
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

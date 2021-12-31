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

package changefeed

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/changefeed"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockchangefeed"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	mock_secondparty_v2 "github.com/pingcap-inc/tiem/test/mocksecondparty_v2"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	knowledge.LoadKnowledge()
	models.MockDB()

	os.Exit(m.Run())
}

func TestManager_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), "clusterId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().Create(gomock.Any(),gomock.Any()).Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "11111",
		},
		Downstream: &changefeed.TiDBDownstream{
		},

	}, nil).AnyTimes()
	changefeedRW.EXPECT().UnlockStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().CreateChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed: true,
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		resp, err := GetManager().Create(context.TODO(), cluster.CreateChangeFeedTaskReq{
			Name: "aa",
			ClusterID: "clusterId",
			StartTS: "121212",
			FilterRules: []string{"*.*"},
			DownstreamType: "tidb",
			Downstream: changefeed.TiDBDownstream{
				Port: 11,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "11111", resp.ID)
		time.Sleep(time.Millisecond * 10)
	})
}

func TestManager_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), "clusterId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().Get(gomock.Any(),gomock.Any()).Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId: "clusterId",
		Downstream: &changefeed.TiDBDownstream{
		},

	}, nil).AnyTimes()
	changefeedRW.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().DeleteChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed: true,
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		resp, err := GetManager().Delete(context.TODO(), cluster.DeleteChangeFeedTaskReq{
			"taskId",
		})
		assert.NoError(t, err)
		assert.Equal(t, "taskId", resp.ID)
		time.Sleep(time.Millisecond * 10)
	})
}

func TestManager_Pause(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), "clusterId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().Get(gomock.Any(),gomock.Any()).Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId: "clusterId",
		Downstream: &changefeed.TiDBDownstream{
		},

	}, nil).AnyTimes()
	changefeedRW.EXPECT().LockStatus(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	changefeedRW.EXPECT().UnlockStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().PauseChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed: true,
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		_, err := GetManager().Pause(context.TODO(), cluster.PauseChangeFeedTaskReq{
			"taskId",
		})
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 10)
	})
}

func TestManager_Resume(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), "clusterId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().Get(gomock.Any(),gomock.Any()).Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId: "clusterId",
		Downstream: &changefeed.TiDBDownstream{
		},

	}, nil).AnyTimes()
	changefeedRW.EXPECT().LockStatus(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	changefeedRW.EXPECT().UnlockStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().ResumeChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed: true,
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		_, err := GetManager().Resume(context.TODO(), cluster.ResumeChangeFeedTaskReq{
			"taskId",
		})
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 10)
	})
	t.Run("without cdc", func(t *testing.T) {
		_, err := GetManager().Resume(context.TODO(), cluster.ResumeChangeFeedTaskReq {
			"taskId",
		})
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 10)
	})
}

func TestManager_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), "clusterId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().Get(gomock.Any(),gomock.Any()).Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			Status: string(constants.ChangeFeedStatusNormal),
			ID:     "taskId",
		},
		ClusterId: "clusterId",
		Downstream: &changefeed.TiDBDownstream{
		},

	}, nil).AnyTimes()
	changefeedRW.EXPECT().LockStatus(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	changefeedRW.EXPECT().UnlockStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	changefeedRW.EXPECT().UpdateConfig(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().UpdateChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed: true,
	}, nil).AnyTimes()

	mockSecond.EXPECT().PauseChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed: true,
	}, nil).AnyTimes()

	mockSecond.EXPECT().ResumeChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed: true,
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		_, err := GetManager().Update(context.TODO(), cluster.UpdateChangeFeedTaskReq{
			Name: "aa",
			FilterRules: []string{"*.*"},
			DownstreamType: "tidb",
			Downstream: changefeed.TiDBDownstream{
				Port: 11,
			},
		})
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 10)
	})
}

func TestManager_Query(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), "clusterId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().QueryByClusterId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*changefeed.ChangeFeedTask{
		{
			Entity: common.Entity{
				ID: "1111",
			},
			ClusterId: "clusterId",
			Downstream: &changefeed.TiDBDownstream{
			},
		},
		{
			Entity: common.Entity{
				ID: "2222",
			},
			ClusterId: "clusterId",
			Downstream: &changefeed.TiDBDownstream{
			},
		},
		{
			Entity: common.Entity{
				ID: "3333",
			},
			ClusterId: "clusterId",
			Downstream: &changefeed.TiDBDownstream{
			},
		},

	}, int64(3), nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().QueryChangeFeedTasks(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedQueryResp {
		Tasks: []secondparty.ChangeFeedInfo{
			{ChangeFeedID: "2222", CheckPointTSO: 9999},
			{ChangeFeedID: "3333", CheckPointTSO: 9999},
			{ChangeFeedID: "4444", CheckPointTSO: 9999},
			{ChangeFeedID: "5555", CheckPointTSO: 9999},
		},
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		resp, total, err := GetManager().Query(context.TODO(), cluster.QueryChangeFeedTaskReq{
			ClusterId: "clusterId",
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Contains(t, []string{resp[0].DownstreamSyncTS, resp[1].DownstreamSyncTS, resp[2].DownstreamSyncTS}, "9999" )
		assert.Contains(t, []string{resp[0].DownstreamSyncTS, resp[1].DownstreamSyncTS, resp[2].DownstreamSyncTS}, "0" )

	})
}

func TestManager_Detail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), "clusterId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().Get(gomock.Any(),gomock.Any()).Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId: "clusterId",
		Downstream: &changefeed.TiDBDownstream{
		},

	}, nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().DetailChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedDetailResp{
		ChangeFeedInfo: secondparty.ChangeFeedInfo {
			CheckPointTSO: 9999,
		},
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		resp, err := GetManager().Detail(context.TODO(), cluster.DetailChangeFeedTaskReq{
			"taskId",
		})
		assert.NoError(t, err)
		assert.Equal(t, "taskId", resp.ID)
		assert.Equal(t, "9999", resp.DownstreamSyncTS)

	})
}

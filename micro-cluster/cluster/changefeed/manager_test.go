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
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
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
	}, []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: "123455678", RoleType: string(constants.Root)},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "11111",
		},
		Downstream: &changefeed.TiDBDownstream{},
	}, nil).AnyTimes()
	changefeedRW.EXPECT().UnlockStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().CreateChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed:  true,
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		resp, err := GetManager().Create(context.TODO(), cluster.CreateChangeFeedTaskReq{
			Name:           "aa",
			ClusterID:      "clusterId",
			StartTS:        "121212",
			FilterRules:    []string{"*.*"},
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
	}, []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: "123455678", RoleType: string(constants.Root)},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId:  "clusterId",
		Downstream: &changefeed.TiDBDownstream{},
	}, nil).AnyTimes()
	changefeedRW.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().DeleteChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed:  true,
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
	}, []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: "123455678", RoleType: string(constants.Root)},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().UnlockStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().PauseChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed:  true,
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		changefeedRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&changefeed.ChangeFeedTask{
			Entity: common.Entity{
				ID: "taskId",
			},
			ClusterId:  "clusterId",
			Downstream: &changefeed.TiDBDownstream{},
		}, nil).Times(1)
		changefeedRW.EXPECT().LockStatus(gomock.Any(), gomock.Any()).Return(nil).Times(1)

		_, err := GetManager().Pause(context.TODO(), cluster.PauseChangeFeedTaskReq{
			"taskId",
		})
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 10)
	})

	t.Run("error1", func(t *testing.T) {
		changefeedRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&changefeed.ChangeFeedTask{
			Entity: common.Entity{
				ID: "taskId",
			},
			ClusterId:  "clusterId",
			Downstream: &changefeed.TiDBDownstream{},
		}, errors.Error(errors.TIEM_UNRECOGNIZED_ERROR)).Times(1)

		_, err := GetManager().Pause(context.TODO(), cluster.PauseChangeFeedTaskReq{
			"taskId",
		})
		assert.Error(t, err)
	})

	t.Run("error2", func(t *testing.T) {
		changefeedRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&changefeed.ChangeFeedTask{
			Entity: common.Entity{
				ID: "taskId",
			},
			ClusterId:  "clusterId",
			Downstream: &changefeed.TiDBDownstream{},
		}, nil).Times(1)
		changefeedRW.EXPECT().LockStatus(gomock.Any(), gomock.Any()).Return(errors.Error(errors.TIEM_PARAMETER_INVALID)).Times(1)

		_, err := GetManager().Pause(context.TODO(), cluster.PauseChangeFeedTaskReq{
			"taskId",
		})
		assert.Error(t, err)
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
	}, []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: "123455678", RoleType: string(constants.Root)},
	}, nil).AnyTimes()
	clusterRW.EXPECT().GetMeta(gomock.Any(), "errorId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: "123455678", RoleType: string(constants.Root)},
	}, errors.Error(errors.TIEM_UNMARSHAL_ERROR)).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().Get(gomock.Any(), "taskId").Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId:  "clusterId",
		Downstream: &changefeed.TiDBDownstream{},
	}, nil).AnyTimes()
	changefeedRW.EXPECT().Get(gomock.Any(), "errorId").Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId:  "clusterId",
		Downstream: &changefeed.TiDBDownstream{},
	}, errors.Error(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR)).AnyTimes()

	changefeedRW.EXPECT().UnlockStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().ResumeChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed:  true,
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		changefeedRW.EXPECT().LockStatus(gomock.Any(), gomock.Any()).Return(nil).Times(1)

		_, err := GetManager().Resume(context.TODO(), cluster.ResumeChangeFeedTaskReq{
			"taskId",
		})
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 10)
	})
	t.Run("cluster error", func(t *testing.T) {
		_, err := GetManager().Resume(context.TODO(), cluster.ResumeChangeFeedTaskReq{
			"errorId",
		})
		assert.Error(t, err)
	})

	t.Run("task error", func(t *testing.T) {
		_, err := GetManager().Resume(context.TODO(), cluster.ResumeChangeFeedTaskReq{
			"errorId",
		})
		assert.Error(t, err)
	})

	t.Run("lock error", func(t *testing.T) {
		changefeedRW.EXPECT().LockStatus(gomock.Any(), gomock.Any()).Return(errors.Error(errors.TIEM_MARSHAL_ERROR)).Times(1)

		_, err := GetManager().Resume(context.TODO(), cluster.ResumeChangeFeedTaskReq{
			"taskId",
		})
		assert.Error(t, err)
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
	}, []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: "123455678", RoleType: string(constants.Root)},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().LockStatus(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	changefeedRW.EXPECT().UnlockStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	changefeedRW.EXPECT().UpdateConfig(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().UpdateChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed:  true,
	}, nil).AnyTimes()

	mockSecond.EXPECT().PauseChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed:  true,
	}, nil).AnyTimes()

	mockSecond.EXPECT().ResumeChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed:  true,
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		changefeedRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&changefeed.ChangeFeedTask{
			Entity: common.Entity{
				Status: string(constants.ChangeFeedStatusNormal),
				ID:     "taskId",
			},
			ClusterId:  "clusterId",
			Downstream: &changefeed.TiDBDownstream{},
		}, nil).Times(1)

		_, err := GetManager().Update(context.TODO(), cluster.UpdateChangeFeedTaskReq{
			Name:           "aa",
			FilterRules:    []string{"*.*"},
			DownstreamType: "tidb",
			Downstream: changefeed.TiDBDownstream{
				Port: 11,
			},
		})
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 10)
	})
	t.Run("error", func(t *testing.T) {
		changefeedRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&changefeed.ChangeFeedTask{
			Entity: common.Entity{
				Status: string(constants.ChangeFeedStatusNormal),
				ID:     "taskId",
			},
			ClusterId:  "clusterId",
			Downstream: &changefeed.TiDBDownstream{},
		}, errors.Error(errors.TIEM_MARSHAL_ERROR)).Times(1)

		_, err := GetManager().Update(context.TODO(), cluster.UpdateChangeFeedTaskReq{
			Name:           "aa",
			FilterRules:    []string{"*.*"},
			DownstreamType: "tidb",
			Downstream: changefeed.TiDBDownstream{
				Port: 11,
			},
		})
		assert.Error(t, err)
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
	}, []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: "123455678", RoleType: string(constants.Root)},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)
	changefeedRW.EXPECT().QueryByClusterId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*changefeed.ChangeFeedTask{
		{
			Entity: common.Entity{
				ID: "1111",
			},
			ClusterId:  "clusterId",
			Downstream: &changefeed.TiDBDownstream{},
		},
		{
			Entity: common.Entity{
				ID: "2222",
			},
			ClusterId:  "clusterId",
			Downstream: &changefeed.TiDBDownstream{},
		},
		{
			Entity: common.Entity{
				ID: "3333",
			},
			ClusterId:  "clusterId",
			Downstream: &changefeed.TiDBDownstream{},
		},
	}, int64(3), nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().QueryChangeFeedTasks(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedQueryResp{
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
		assert.Contains(t, []string{resp[0].DownstreamSyncTS, resp[1].DownstreamSyncTS, resp[2].DownstreamSyncTS}, "9999")
		assert.Contains(t, []string{resp[0].DownstreamSyncTS, resp[1].DownstreamSyncTS, resp[2].DownstreamSyncTS}, "0")

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
	}, []*management.DBUser{
		{ClusterID: "clusterId", Name: "root", Password: "123455678", RoleType: string(constants.Root)},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().DetailChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedDetailResp{
		ChangeFeedInfo: secondparty.ChangeFeedInfo{
			CheckPointTSO: 9999,
		},
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		changefeedRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&changefeed.ChangeFeedTask{
			Entity: common.Entity{
				ID: "taskId",
			},
			ClusterId:  "clusterId",
			Downstream: &changefeed.TiDBDownstream{},
		}, nil).Times(1)

		resp, err := GetManager().Detail(context.TODO(), cluster.DetailChangeFeedTaskReq{
			"taskId",
		})
		assert.NoError(t, err)
		assert.Equal(t, "taskId", resp.ID)
		assert.Equal(t, "9999", resp.DownstreamSyncTS)

	})
	t.Run("error", func(t *testing.T) {
		changefeedRW.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&changefeed.ChangeFeedTask{
			Entity: common.Entity{
				ID: "taskId",
			},
			ClusterId:  "clusterId",
			Downstream: &changefeed.TiDBDownstream{},
		}, errors.Error(errors.TIEM_UNMARSHAL_ERROR)).Times(1)

		_, err := GetManager().Detail(context.TODO(), cluster.DetailChangeFeedTaskReq{
			"taskId",
		})
		assert.Error(t, err)

	})
}

func TestManager_updateExecutor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().UpdateChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: false,
		Succeed:  false,
	}, nil).AnyTimes()

	clusterMeta := &meta.ClusterMeta{
		Instances: map[string][]*management.ClusterInstance{
			"CDC": {
				{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
			},
		},
	}

	task := &changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId:  "clusterId",
		Downstream: &changefeed.TiDBDownstream{},
	}
	t.Run("failed", func(t *testing.T) {
		err := GetManager().updateExecutor(context.TODO(), clusterMeta, task)
		assert.Error(t, err)
	})

}
func TestManager_pauseExecutor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().PauseChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: false,
		Succeed:  false,
	}, nil).AnyTimes()

	clusterMeta := &meta.ClusterMeta{
		Instances: map[string][]*management.ClusterInstance{
			"CDC": {
				{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
			},
		},
	}

	task := &changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId:  "clusterId",
		Downstream: &changefeed.TiDBDownstream{},
	}
	t.Run("failed", func(t *testing.T) {
		err := GetManager().pauseExecutor(context.TODO(), clusterMeta, task)
		assert.Error(t, err)
	})

}
func TestManager_resumeExecutor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().ResumeChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: false,
		Succeed:  false,
	}, nil).AnyTimes()

	clusterMeta := &meta.ClusterMeta{
		Instances: map[string][]*management.ClusterInstance{
			"CDC": {
				{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
			},
		},
	}

	task := &changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId:  "clusterId",
		Downstream: &changefeed.TiDBDownstream{},
	}
	t.Run("failed", func(t *testing.T) {
		err := GetManager().resumeExecutor(context.TODO(), clusterMeta, task)
		assert.Error(t, err)
	})

}
func TestManager_createExecutor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().CreateChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: false,
		Succeed:  false,
	}, nil).AnyTimes()

	clusterMeta := &meta.ClusterMeta{
		Instances: map[string][]*management.ClusterInstance{
			"CDC": {
				{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
			},
		},
	}

	task := &changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "taskId",
		},
		ClusterId:  "clusterId",
		Downstream: &changefeed.TiDBDownstream{},
	}
	t.Run("failed", func(t *testing.T) {
		err := GetManager().createExecutor(context.TODO(), clusterMeta, task)
		assert.Error(t, err)
	})

}

func Test_ClusterError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), "NotFound").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, []*management.DBUser{
		{ClusterID: "NotFound", Name: "root", Password: "123455678", RoleType: string(constants.Root)},
	}, errors.Error(errors.TIEM_MARSHAL_ERROR)).AnyTimes()

	clusterRW.EXPECT().GetMeta(gomock.Any(), "WithoutCDC").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "PD", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, []*management.DBUser{
		{ClusterID: "WithoutCDC", Name: "root", Password: "123455678", RoleType: string(constants.Root)},
	}, nil).AnyTimes()

	changefeedRW := mockchangefeed.NewMockReaderWriter(ctrl)
	models.SetChangeFeedReaderWriter(changefeedRW)

	changefeedRW.EXPECT().Get(gomock.Any(), "NotFound").Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "NotFound",
		},
		ClusterId:  "NotFound",
		Downstream: &changefeed.TiDBDownstream{},
	}, nil).AnyTimes()
	changefeedRW.EXPECT().Get(gomock.Any(), "WithoutCDC").Return(&changefeed.ChangeFeedTask{
		Entity: common.Entity{
			ID: "WithoutCDC",
		},
		ClusterId:  "WithoutCDC",
		Downstream: &changefeed.TiDBDownstream{},
	}, nil).AnyTimes()
	t.Run("create", func(t *testing.T) {
		_, err := GetManager().Create(context.TODO(), cluster.CreateChangeFeedTaskReq{
			ClusterID: "NotFound",
		})
		assert.Error(t, err)
		_, err = GetManager().Create(context.TODO(), cluster.CreateChangeFeedTaskReq{
			ClusterID: "WithoutCDC",
		})
		assert.Error(t, err)
	})
	t.Run("delete", func(t *testing.T) {
		_, err := GetManager().Delete(context.TODO(), cluster.DeleteChangeFeedTaskReq{
			ID: "NotFound",
		})
		assert.Error(t, err)
		_, err = GetManager().Delete(context.TODO(), cluster.DeleteChangeFeedTaskReq{
			ID: "WithoutCDC",
		})
		assert.Error(t, err)
	})
	t.Run("pause", func(t *testing.T) {
		_, err := GetManager().Pause(context.TODO(), cluster.PauseChangeFeedTaskReq{
			ID: "NotFound",
		})
		assert.Error(t, err)
		_, err = GetManager().Pause(context.TODO(), cluster.PauseChangeFeedTaskReq{
			ID: "WithoutCDC",
		})
		assert.Error(t, err)
	})
	t.Run("resume", func(t *testing.T) {
		_, err := GetManager().Resume(context.TODO(), cluster.ResumeChangeFeedTaskReq{
			ID: "NotFound",
		})
		assert.Error(t, err)
		_, err = GetManager().Resume(context.TODO(), cluster.ResumeChangeFeedTaskReq{
			ID: "WithoutCDC",
		})
		assert.Error(t, err)
	})
	t.Run("update", func(t *testing.T) {
		_, err := GetManager().Update(context.TODO(), cluster.UpdateChangeFeedTaskReq{
			ID: "NotFound",
		})
		assert.Error(t, err)
		_, err = GetManager().Update(context.TODO(), cluster.UpdateChangeFeedTaskReq{
			ID: "WithoutCDC",
		})
		assert.Error(t, err)
	})
	t.Run("detail", func(t *testing.T) {
		_, err := GetManager().Detail(context.TODO(), cluster.DetailChangeFeedTaskReq{
			ID: "NotFound",
		})
		assert.Error(t, err)
		_, err = GetManager().Detail(context.TODO(), cluster.DetailChangeFeedTaskReq{
			ID: "WithoutCDC",
		})
		assert.Error(t, err)
	})
	t.Run("query", func(t *testing.T) {
		_, _, err := GetManager().Query(context.TODO(), cluster.QueryChangeFeedTaskReq{
			ClusterId: "NotFound",
		})
		assert.Error(t, err)
		_, _, err = GetManager().Query(context.TODO(), cluster.QueryChangeFeedTaskReq{
			ClusterId: "WithoutCDC",
		})
		assert.Error(t, err)
	})

}

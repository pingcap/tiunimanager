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
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/changefeed"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockchangefeed"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	mock_secondparty_v2 "github.com/pingcap-inc/tiem/test/mocksecondparty_v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestManager_CreateBetweenClusters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), "sourceId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "TiDB", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, nil).AnyTimes()
	clusterRW.EXPECT().GetMeta(gomock.Any(), "targetId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.1.1"}, Ports: []int32{111}},
		{Type: "TiDB", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.1.2"}, Ports: []int32{111}},
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
		id, err := GetChangeFeedService().CreateBetweenClusters(context.TODO(), "sourceId", "targetId", constants.ClusterRelationCloneFrom)
		assert.NoError(t, err)
		assert.Equal(t, "11111", id)
		time.Sleep(time.Millisecond * 10)
	})
}

func TestManager_ReverseBetweenClusters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), "sourceId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.1"}, Ports: []int32{111}},
		{Type: "TiDB", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.0.2"}, Ports: []int32{111}},
	}, nil).AnyTimes()
	clusterRW.EXPECT().GetMeta(gomock.Any(), "targetId").Return(&management.Cluster{}, []*management.ClusterInstance{
		{Type: "CDC", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.1.1"}, Ports: []int32{111}},
		{Type: "TiDB", Entity: common.Entity{Status: string(constants.ClusterInstanceRunning)}, HostIP: []string{"127.0.1.2"}, Ports: []int32{111}},
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
	changefeedRW.EXPECT().QueryByClusterId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*changefeed.ChangeFeedTask{
		{
			Entity: common.Entity{
				Status: string(constants.ChangeFeedStatusNormal),
				ID: "1111",
			},
			Type: "tidb",
			ClusterId: "sourceId",
			Downstream: &changefeed.TiDBDownstream{
				TargetClusterId: "targetId",
			},
		},
		{
			Entity: common.Entity{
				ID: "2222",
			},
			Type: "mysql",
			ClusterId: "sourceId",
			Downstream: &changefeed.MysqlDownstream{
			},
		},
		{
			Entity: common.Entity{
				Status: string(constants.ChangeFeedStatusFinished),
				ID:     "3333",
			},
			ClusterId: "sourceId",
			Downstream: &changefeed.TiDBDownstream{
				TargetClusterId: "targetId",
			},
		},
	}, int64(3), nil).AnyTimes()
	changefeedRW.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockSecond := mock_secondparty_v2.NewMockSecondPartyService(ctrl)
	secondparty.Manager = mockSecond

	mockSecond.EXPECT().CreateChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed: true,
	}, nil).AnyTimes()

	mockSecond.EXPECT().DeleteChangeFeedTask(gomock.Any(), gomock.Any()).Return(secondparty.ChangeFeedCmdAcceptResp{
		Accepted: true,
		Succeed: true,
	}, nil).AnyTimes()

	t.Run("normal", func(t *testing.T) {
		id, err := GetChangeFeedService().ReverseBetweenClusters(context.TODO(), "sourceId", "targetId", constants.ClusterRelationCloneFrom)
		assert.NoError(t, err)
		assert.Equal(t, "11111", id)
		time.Sleep(time.Millisecond * 10)
	})
}

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
 *                                                                            *
 ******************************************************************************/

package management

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetDashboardInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)

	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:        "2145635758",
			TenantId:  "324567",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:              "koojdafij",
		DBUser:            "kodjsfn",
		DBPassword:        "mypassword",
		Type:              "TiDB",
		Version:           "v5.0.0",
		Tags:              []string{"111", "333"},
		OwnerId:           "436534636u",
		ParameterGroupID:  "352467890",
		Copies:            4,
		Region:            "Region1",
		CpuArchitecture:   "x86_64",
		MaintenanceStatus: constants.ClusterMaintenanceCreating,
	}, []*management.ClusterInstance{
		{
			Entity: common.Entity{
				Status: string(constants.ClusterInstanceRunning),
			},
			Zone:     "zone1",
			CpuCores: 4,
			Memory:   8,
			Type:     "TiDB",
			Version:  "v5.0.0",
			Ports:    []int32{10001, 10002, 10003, 10004},
			HostIP:   []string{"127.0.0.1"},
		},
		{
			Entity: common.Entity{
				Status: string(constants.ClusterInstanceRunning),
			},
			Zone:     "zone1",
			CpuCores: 4,
			Memory:   8,
			Type:     "TiKV",
			Version:  "v5.0.0",
			Ports:    []int32{20001, 20002, 20003, 20004},
			HostIP:   []string{"127.0.0.2"},
		},
		{
			Entity: common.Entity{
				Status: string(constants.ClusterInstanceRunning),
			},
			Zone:     "zone1",
			CpuCores: 4,
			Memory:   8,
			Type:     "PD",
			Version:  "v5.0.0",
			Ports:    []int32{30001, 30002, 30003, 30004},
			HostIP:   []string{"127.0.0.3"},
		},
		{
			Entity: common.Entity{
				Status: string(constants.ClusterInstanceRunning),
			},
			Zone:     "zone1",
			CpuCores: 3,
			Memory:   7,
			Type:     "PD",
			Version:  "v5.0.0",
			Ports:    []int32{30001, 30002, 30003, 30004},
			HostIP:   []string{"127.0.0.3"},
		},
	}, nil)

	_, err := GetDashboardInfo(context.TODO(), cluster.GetDashboardInfoReq{ClusterID: "2145635758"})
	assert.Error(t, err)
}


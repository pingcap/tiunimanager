/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package backuprestore

import (
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/common"
	mock_br_service "github.com/pingcap-inc/tiem/test/mockbr"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockbr"
	"github.com/pingcap-inc/tiem/test/mockmodels/mockclustermanagement"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAutoBackupManager(t *testing.T) {
	manager := NewAutoBackupManager()
	assert.NotNil(t, manager)
}

func Test_AutoBackup_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	strategies := make([]*backuprestore.BackupStrategy, 0)
	brRW := mockbr.NewMockReaderWriter(ctrl)
	brRW.EXPECT().QueryBackupStrategy(gomock.Any(), gomock.Any(), gomock.Any()).Return(strategies, nil).AnyTimes()
	models.SetBRReaderWriter(brRW)

	handler := &autoBackupHandler{}
	handler.Run()
}

func Test_AutoBackup_doBackup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clusterRW := mockclustermanagement.NewMockReaderWriter(ctrl)
	models.SetClusterReaderWriter(clusterRW)
	clusterRW.EXPECT().GetMeta(gomock.Any(), gomock.Any()).Return(&management.Cluster{
		Entity: common.Entity{
			ID:       "id-xxxx",
			TenantId: "tid-xxx",
		},
	}, make([]*management.ClusterInstance, 0), make([]*management.DBUser, 0), nil).AnyTimes()
	clusterRW.EXPECT().SetMaintenanceStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockBRService := mock_br_service.NewMockBRService(ctrl)
	mockBRService.EXPECT().BackupCluster(gomock.Any(), gomock.Any(),
		gomock.Any()).Return(cluster.BackupClusterDataResp{}, nil).AnyTimes()
	MockBRService(mockBRService)
	defer MockBRService(NewBRManager())

	strategy := &backuprestore.BackupStrategy{
		ClusterID:  "cls-xxxx",
		BackupDate: "Monday,Friday",
		StartHour:  0,
		EndHour:    1,
	}
	handler := &autoBackupHandler{}
	handler.doBackup(strategy)
}

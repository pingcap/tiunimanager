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

package backuprestore

import (
	"context"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message/cluster"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/cluster/backuprestore"
	"github.com/robfig/cron"
	"time"
)

type autoBackupManager struct {
	JobCron *cron.Cron
	JobSpec string
}

type autoBackupHandler struct {
}

func NewAutoBackupManager() *autoBackupManager {
	mgr := &autoBackupManager{
		JobCron: cron.New(),
		JobSpec: "0 0 * * * *", // every integer hour
	}
	err := mgr.JobCron.AddJob(mgr.JobSpec, &autoBackupHandler{})
	if err != nil {
		framework.Log().Fatalf("add auto backup cron job failed, %s", err.Error())
		return nil
	}
	go mgr.start()

	return mgr
}

func (mgr *autoBackupManager) start() {
	time.Sleep(5 * time.Second) //wait db client ready
	mgr.JobCron.Start()
	defer mgr.JobCron.Stop()

	select {}
}

func (auto *autoBackupHandler) Run() {
	curWeekDay := time.Now().Weekday().String()
	curHour := time.Now().Hour()

	framework.Log().Infof("begin AutoBackupHandler Run at WeekDay: %s, Hour: %d", curWeekDay, curHour)
	defer framework.Log().Infof("end AutoBackupHandler Run")

	rw := models.GetBRReaderWriter()
	strategies, err := rw.QueryBackupStrategy(context.TODO(), curWeekDay, uint32(curHour))
	if err != nil {
		framework.Log().Errorf("query backup strategy by weekday %s, hour: %d failed, %s", curWeekDay, curHour, err.Error())
		return
	}

	framework.Log().Infof("WeekDay %s, Hour: %d need do auto backup for %d clusters", curWeekDay, curHour, len(strategies))
	for _, strategy := range strategies {
		go auto.doBackup(strategy)
	}
}

func (auto *autoBackupHandler) doBackup(strategy *backuprestore.BackupStrategy) {
	framework.Log().Infof("begin do auto backup for cluster %s", strategy.ClusterID)
	defer framework.Log().Infof("end do auto backup for cluster %s", strategy.ClusterID)

	meta, err := meta.Get(context.Background(), strategy.ClusterID)
	if err != nil || meta.Cluster == nil {
		framework.LogWithContext(context.Background()).Errorf("load cluster meta %s failed, %v", strategy.ClusterID, err)
		return
	}

	ctx := framework.NewMicroContextWithKeyValuePairs(context.Background(), map[string]string{framework.TiUniManager_X_TENANT_ID_KEY: meta.Cluster.TenantId})
	_, err = GetBRService().BackupCluster(ctx, cluster.BackupClusterDataReq{
		ClusterID:  strategy.ClusterID,
		BackupMode: string(constants.BackupModeAuto),
	}, true)
	if err != nil {
		framework.LogWithContext(context.Background()).Errorf("do backup for cluster %s failed, %s", strategy.ClusterID, err.Error())
		return
	}
}

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
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
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

	_, err := GetBRManager().BackupCluster(framework.NewMicroCtxFromGinCtx(&gin.Context{}), &cluster.BackupClusterDataReq{
		ClusterID:  strategy.ClusterID,
		BackupMode: string(constants.BackupModeAuto),
	})
	if err != nil {
		framework.Log().Errorf("do backup for cluster %s failed, %s", strategy.ClusterID, err.Error())
		return
	}
}

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

package domain

import (
	ctx "context"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/robfig/cron"
	"time"
)

type AutoBackupManager struct {
	JobCron *cron.Cron
	JobSpec string
}

type autoBackupHandler struct {
}

func InitAutoBackupCronJob() {
	mgr := NewAutoBackupManager()
	go mgr.Start()
}

func NewAutoBackupManager() *AutoBackupManager {
	mgr := &AutoBackupManager{
		JobCron: cron.New(),
		JobSpec: "0 0 * * * *", // every integer hour
	}
	err := mgr.JobCron.AddJob(mgr.JobSpec, &autoBackupHandler{})
	if err != nil {
		getLogger().Fatalf("add auto backup cron job failed, %s", err.Error())
		return nil
	}

	return mgr
}

func (mgr *AutoBackupManager) Start() {
	time.Sleep(5 * time.Second) //wait db client ready
	mgr.JobCron.Start()
	defer mgr.JobCron.Stop()

	select {}
}

func (auto *autoBackupHandler) Run() {
	curWeekDay := time.Now().Weekday().String()
	curHour := time.Now().Hour()

	getLogger().Infof("begin AutoBackupHandler Run at WeekDay: %s, Hour: %d", curWeekDay, curHour)
	defer getLogger().Infof("end AutoBackupHandler Run")

	resp, err := client.DBClient.QueryBackupStrategyByTime(ctx.TODO(), &dbpb.DBQueryBackupStrategyByTimeRequest{
		Weekday:   curWeekDay,
		StartHour: uint32(curHour),
	})
	if err != nil {
		getLogger().Errorf("query backup strategy by weekday %s, hour: %d failed, %s", curWeekDay, curHour, err.Error())
		return
	}

	getLogger().Infof("WeekDay %s, Hour: %d need do auto backup for %d clusters", curWeekDay, curHour, len(resp.GetStrategys()))
	for _, strategy := range resp.GetStrategys() {
		go auto.doBackup(strategy)
	}
}

func (auto *autoBackupHandler) doBackup(strategy *dbpb.DBBackupStrategyDTO) {
	getLogger().Infof("begin do auto backup for cluster %s", strategy.GetClusterId())
	defer getLogger().Infof("end do auto backup for cluster %s", strategy.GetClusterId())

	ope := &clusterpb.OperatorDTO{
		Id:             SystemOperator,
		Name:           SystemOperator,
		ManualOperator: false,
	}
	_, err := Backup(framework.NewMicroCtxFromGinCtx(&gin.Context{}), ope, strategy.GetClusterId(), "", "", BackupModeAuto, "")
	if err != nil {
		getLogger().Errorf("do backup for cluster %s failed, %s", strategy.GetClusterId(), err.Error())
		return
	}
}

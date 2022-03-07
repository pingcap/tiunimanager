/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

package check

import (
	ctx "context"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/robfig/cron"
	"time"
)

type autoCheckManager struct {
	JobCron *cron.Cron
	JobSpec string
}

type autoCheckHandler struct {
}

func NewAutoCheckManager() *autoCheckManager {
	mgr := &autoCheckManager{
		JobCron: cron.New(),
		JobSpec: "0 0 23 * * ?", // every day 23:00:00
	}
	err := mgr.JobCron.AddJob(mgr.JobSpec, &autoCheckHandler{})
	if err != nil {
		framework.Log().Fatalf("add auto check cron job failed, %s", err.Error())
		return nil
	}
	go mgr.start()

	return mgr
}

func (mgr *autoCheckManager) start() {
	time.Sleep(5 * time.Second) //wait db client ready
	mgr.JobCron.Start()
	defer mgr.JobCron.Stop()

	select {}
}

func (auto *autoCheckHandler) Run() {
	curWeekDay := time.Now().Weekday().String()
	curHour := time.Now().Hour()

	framework.Log().Infof("begin AutoCheckHandler Run at WeekDay: %s, Hour: %d", curWeekDay, curHour)
	defer framework.Log().Infof("end AutoCheckHandler Run")

	go func() {
		framework.Log().Infof("Start to check platform")
		GetCheckService().Check(ctx.TODO(), message.CheckPlatformReq{})
	}()
}

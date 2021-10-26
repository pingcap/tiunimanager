
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

package service

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/thirdparty/metrics"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type DBServiceHandler struct {
	dao *models.DAOManager
}

func NewDBServiceHandler(dataDir string, fw *framework.BaseFramework) *DBServiceHandler {
	handler := new(DBServiceHandler)
	dao := models.NewDAOManager(fw)
	handler.SetDao(dao)

	dao.InitDB(dataDir)
	if dao.Db().Migrator().HasTable(&models.Tenant{}) {
		framework.Log().Warn("data existed, skip initialization")
		return handler
	} else {
		dao.InitTables()
		dao.InitData()
	}
	return handler
}

func (handler *DBServiceHandler) Dao() *models.DAOManager {
	return handler.dao
}

func (handler *DBServiceHandler) SetDao(dao *models.DAOManager) {
	handler.dao = dao
}

func (handler *DBServiceHandler) HandleMetrics(start time.Time, funcName string, code string) {
	duration := time.Now().Sub(start)
	framework.Current.GetMetrics().SqliteDurationHistogramMetric.With(prometheus.Labels{
		metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
		metrics.MethodLabel: funcName,
		metrics.CodeLabel: code}).
		Observe(duration.Seconds())
	framework.Current.GetMetrics().SqliteRequestsCounterMetric.With(prometheus.Labels{
		metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
		metrics.MethodLabel: funcName,
		metrics.CodeLabel: code}).
		Inc()
}

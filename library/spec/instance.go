/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

/*******************************************************************************
 * @File: instance
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/24
*******************************************************************************/

package spec

import (
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

type TiDBClusterComponent string

const (
	TiDBClusterComponent_TiDB       		= TiDBClusterComponent(spec.ComponentTiDB)
	TiDBClusterComponent_TiKV       		= TiDBClusterComponent(spec.ComponentTiKV)
	TiDBClusterComponent_PD       			= TiDBClusterComponent(spec.ComponentPD)
	TiDBClusterComponent_TiFlash       		= TiDBClusterComponent(spec.ComponentTiFlash)
	TiDBClusterComponent_TiFlashLearner     = TiDBClusterComponent("tiflash-learner")
	TiDBClusterComponent_Grafana       		= TiDBClusterComponent(spec.ComponentGrafana)
	TiDBClusterComponent_Drainer       		= TiDBClusterComponent(spec.ComponentDrainer)
	TiDBClusterComponent_Pump       		= TiDBClusterComponent(spec.ComponentPump)
	TiDBClusterComponent_CDC       			= TiDBClusterComponent(spec.ComponentCDC)
	TiDBClusterComponent_TiSpark       		= TiDBClusterComponent(spec.ComponentTiSpark)
	TiDBClusterComponent_TiSparkMasters     = TiDBClusterComponent("tispark_masters")
	TiDBClusterComponent_TiSparkWorkers     = TiDBClusterComponent("tispark_workers")
	TiDBClusterComponent_Spark       		= TiDBClusterComponent(spec.ComponentSpark)
	TiDBClusterComponent_Alertmanager		= TiDBClusterComponent(spec.ComponentAlertmanager)
	TiDBClusterComponent_DMMaster       	= TiDBClusterComponent(spec.ComponentDMMaster)
	TiDBClusterComponent_DMWorker       	= TiDBClusterComponent(spec.ComponentDMWorker)
	TiDBClusterComponent_Prometheus       	= TiDBClusterComponent(spec.ComponentPrometheus)
	TiDBClusterComponent_Pushwaygate       	= TiDBClusterComponent(spec.ComponentPushwaygate)
	TiDBClusterComponent_BlackboxExporter	= TiDBClusterComponent(spec.ComponentBlackboxExporter)
	TiDBClusterComponent_NodeExporter       = TiDBClusterComponent(spec.ComponentNodeExporter)
	TiDBClusterComponent_CheckCollector    	= TiDBClusterComponent(spec.ComponentCheckCollector)
)
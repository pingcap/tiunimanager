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

/**
 * @Author: guobob
 * @Description:
 * @File:  dbreplay_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/14 17:08
 */

package dbreplay

import (
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewTrafficReplayRun(t *testing.T){
	task := new(cluster.TrafficReplayTask)
	task.ProductClusterID="cluster-product-cluster-id"
	task.SimulationClusterID="cluster-simulation-cluster-id"
	tr := NewTrafficReplayRun(task)
	assert.New(t).NotNil(tr)
	assert.New(t).Equal(task.ProductClusterID,tr.Deploy.Basic.ClusterIDs[0])
	assert.New(t).Equal(task.SimulationClusterID,tr.Deploy.Basic.ClusterIDs[1])
}


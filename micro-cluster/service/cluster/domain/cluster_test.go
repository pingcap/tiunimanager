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

package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCluster_Delete(t *testing.T) {
	cluster := Cluster{
		Status: ClusterStatusOnline,
	}

	cluster.Delete()
	assert.Equal(t, ClusterStatusDeleted, cluster.Status)
}

func TestCluster_Offline(t *testing.T) {
	cluster := Cluster{
		Status: ClusterStatusOnline,
	}

	cluster.Offline()
	assert.Equal(t, ClusterStatusOffline, cluster.Status)
}

func TestCluster_Online(t *testing.T) {
	cluster := Cluster{
		Status: ClusterStatusOffline,
	}

	cluster.Online()
	assert.Equal(t, ClusterStatusOnline, cluster.Status)
}

func TestCluster_Restart(t *testing.T) {
	cluster := Cluster{
		Status: ClusterStatusOnline,
	}

	cluster.Restart()
	assert.Equal(t, ClusterStatusRestarting, cluster.Status)
}

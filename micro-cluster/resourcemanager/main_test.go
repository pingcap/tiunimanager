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

package resourcemanager

import (
	"os"
	"testing"

	"github.com/pingcap-inc/tiunimanager/library/framework"
	"github.com/pingcap-inc/tiunimanager/micro-cluster/resourcemanager/management"
	"github.com/pingcap-inc/tiunimanager/micro-cluster/resourcemanager/resourcepool"
	"github.com/pingcap-inc/tiunimanager/models"
)

var resourceManager *ResourceManager

func NewMockResourceManager() *ResourceManager {

	m := new(ResourceManager)
	m.resourcePool = new(resourcepool.ResourcePool)
	m.management = new(management.Management)
	m.resourcePool.InitResourcePool()
	m.management.InitManagement()
	return m
}

func TestMain(m *testing.M) {
	framework.InitBaseFrameworkForUt(framework.ClusterService)
	models.MockDB()
	resourceManager = NewMockResourceManager()
	os.Exit(m.Run())
}

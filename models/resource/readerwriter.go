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

package resource

import (
	"context"

	"github.com/pingcap-inc/tiem/common/structs"
	rp "github.com/pingcap-inc/tiem/models/resource/resourcepool"

	"gorm.io/gorm"
)

type QueryCond struct {
	Status  string
	Stat    string
	Purpose string
	Offset  int
	Limit   int
}

type Location struct {
	Region string
	Zone   string
	Rack   string
	Host   string
}
type HostCondition struct {
	Status *int32
	Stat   *int32
	Arch   *string
}
type DiskCondition struct {
	Type     *string
	Capacity *int32
	Status   *int32
}
type StockFilter struct {
	Location      Location
	HostCondition HostCondition
	DiskCondition DiskCondition
}

type Stock struct {
	FreeCpuCores     int
	FreeMemory       int
	FreeDiskCount    int
	FreeDiskCapacity int
}

type ResourceReaderWriter interface {
	SetDb(db *gorm.DB)
	Db(ctx context.Context) *gorm.DB

	Create(ctx context.Context, hosts []rp.Host) ([]string, error)
	Delete(ctx context.Context, hostIds []string) (err error)
	Get(ctx context.Context, hostId string) (rp.Host, error)
	Query(ctx context.Context, filter structs.HostFilter) (hosts []rp.Host, total int64, err error)

	UpdateHostStatus(ctx context.Context, status string) (err error)
	ReserveHost(ctx context.Context, reserved bool) (err error)
	GetHierarchy(ctx context.Context, filter structs.HostFilter, level int32, depth int32) (root structs.HierarchyTreeNode, err error)
	GetStocks(ctx context.Context, filter StockFilter) (stock Stock, err error)
}

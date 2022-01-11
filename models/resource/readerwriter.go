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
	resource_structs "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	rp "github.com/pingcap-inc/tiem/models/resource/resourcepool"
)

// Use HostItem to store filtered hosts records to build hierarchy tree
type HostItem struct {
	Region string
	Az     string
	Rack   string
	Ip     string
	Name   string
}
type ReaderWriter interface {
	// Create a batch of hosts
	Create(ctx context.Context, hosts []rp.Host) ([]string, error)
	// Delete a batch of hosts
	Delete(ctx context.Context, hostIds []string) (err error)
	// Query hosts, if specify HostID in HostFilter, return a single host info
	Query(ctx context.Context, filter *structs.HostFilter, offset int, limit int) (hosts []rp.Host, total int64, err error)

	UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error)
	UpdateHostReserved(ctx context.Context, hostIds []string, reserved bool) (err error)

	// Get all filtered hosts to build hierarchy tree
	GetHostItems(ctx context.Context, filter *structs.HostFilter, level int32, depth int32) (items []HostItem, err error)
	// Get a list of stock on each host
	GetHostStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks []structs.Stocks, err error)

	// Alloc/Recycle resources, used by cluster module internal
	AllocResources(ctx context.Context, batchReq *resource_structs.BatchAllocRequest) (results *resource_structs.BatchAllocResponse, err error)
	RecycleResources(ctx context.Context, request *resource_structs.RecycleRequest) (err error)
}

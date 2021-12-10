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
	// Init all table for resource manager
	InitTables(ctx context.Context) error
	Create(ctx context.Context, hosts []rp.Host) ([]string, error)
	Delete(ctx context.Context, hostIds []string) (err error)
	Query(ctx context.Context, filter *structs.HostFilter, offset int, limit int) (hosts []rp.Host, err error)

	UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error)
	UpdateHostReserved(ctx context.Context, hostIds []string, reserved bool) (err error)
	// Get all filtered hosts to build hierarchy tree
	GetHostItems(ctx context.Context, filter *structs.HostFilter, level int32, depth int32) (items []HostItem, err error)
	GetHostStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks []structs.Stocks, err error)
}

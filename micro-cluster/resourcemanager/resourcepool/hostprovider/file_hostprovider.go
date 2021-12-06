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

package hostprovider

import (
	"context"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/models/resource"
	"github.com/pingcap-inc/tiem/models/resource/resourcepool"
)

type FileHostProvider struct {
	rw resource.ResourceReaderWriter
}

func GetFileHostProvider() HostProvider {
	hostProvider := new(FileHostProvider)
	hostProvider.rw = resource.NewGormChangeFeedReadWrite()
	return hostProvider
}

func (p *FileHostProvider) ResourceToDBModel(src *structs.HostInfo, dst *resourcepool.Host) {
	dst.HostName = src.HostName
	dst.IP = src.IP
	dst.UserName = src.UserName
	dst.Passwd = src.Passwd
	dst.Arch = src.Arch
	dst.OS = src.OS
	dst.Kernel = src.Kernel
	dst.FreeCpuCores = src.FreeCpuCores
	dst.FreeMemory = src.FreeMemory
	dst.Spec = src.Spec
	dst.CpuCores = src.CpuCores
	dst.Memory = src.Memory
	dst.Nic = src.Nic
	dst.Region = src.Region
	dst.AZ = resourcepool.GenDomainCodeByName(dst.Region, src.AZ)
	dst.Rack = resourcepool.GenDomainCodeByName(dst.AZ, src.Rack)
	dst.Status = src.Status
	dst.Stat = src.Stat
	dst.ClusterType = src.ClusterType
	dst.Purpose = src.Purpose
	dst.DiskType = src.DiskType
	dst.Reserved = src.Reserved
	dst.Traits = src.Traits
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, resourcepool.Disk{
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   disk.Status,
			Type:     disk.Type,
		})
	}
}

func (p *FileHostProvider) DBModelToResource(src *resourcepool.Host, dst *structs.HostInfo) {
	dst.ID = src.ID
	dst.HostName = src.HostName
	dst.IP = src.IP
	dst.Arch = src.Arch
	dst.OS = src.OS
	dst.Kernel = src.Kernel
	dst.FreeCpuCores = src.FreeCpuCores
	dst.FreeMemory = src.FreeMemory
	dst.Spec = src.Spec
	dst.CpuCores = src.CpuCores
	dst.Memory = src.Memory
	dst.Nic = src.Nic
	dst.Region = src.Region
	dst.AZ = resourcepool.GetDomainNameFromCode(src.AZ)
	dst.Rack = resourcepool.GetDomainNameFromCode(src.Rack)
	dst.Status = src.Status
	dst.ClusterType = src.ClusterType
	dst.Purpose = src.Purpose
	dst.DiskType = src.DiskType
	dst.CreatedAt = src.CreatedAt.Unix()
	dst.UpdatedAt = src.UpdatedAt.Unix()
	dst.Reserved = src.Reserved
	dst.Traits = src.Traits
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, structs.DiskInfo{
			ID:       disk.ID,
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   disk.Status,
			Type:     disk.Type,
		})
	}
	// Update Host's load stat after diskInfo is updated
	dst.Stat = src.Stat
	if dst.Stat == string(constants.HostLoadInUsed) {
		stat, isExhaust := dst.IsExhaust()
		if isExhaust {
			dst.Stat = string(stat)
		}
	}
}

func (p *FileHostProvider) SetResourceReaderWriter(rw resource.ResourceReaderWriter) {
	p.rw = rw
}

func (p *FileHostProvider) ImportHosts(ctx context.Context, hosts []structs.HostInfo) (hostIds []string, err error) {
	var dbModelHosts []resourcepool.Host
	for _, host := range hosts {
		var dbHost resourcepool.Host
		p.ResourceToDBModel(&host, &dbHost)
		dbModelHosts = append(dbModelHosts, dbHost)
	}
	return p.rw.Create(ctx, dbModelHosts)
}

func (p *FileHostProvider) DeleteHosts(ctx context.Context, hostIds []string) (err error) {
	return nil
}

func (p *FileHostProvider) Query(ctx context.Context, filter structs.HostFilter) (hosts []*structs.HostInfo, err error) {
	return nil, nil
}

func (p *FileHostProvider) UpdateHostStatus(ctx context.Context, hostId []string, status string) (err error) {
	return nil
}

func (p *FileHostProvider) UpdateHostReserved(ctx context.Context, hostId []string, reserved bool) (err error) {
	return nil
}

func (p *FileHostProvider) GetHierarchy(ctx context.Context, filter structs.HostFilter, level int32, depth int32) (root *structs.HierarchyTreeNode, err error) {
	return nil, nil
}

func (p *FileHostProvider) GetStocks(ctx context.Context, location structs.Location, hostFilter structs.HostFilter, diskFilter structs.DiskFilter) (stocks *structs.Stocks, err error) {
	return nil, nil
}

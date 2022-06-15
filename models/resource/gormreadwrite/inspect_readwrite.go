/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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

package gormreadwrite

import (
	"context"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	cl "github.com/pingcap/tiunimanager/models/cluster/management"
	mm "github.com/pingcap/tiunimanager/models/resource/management"
	rp "github.com/pingcap/tiunimanager/models/resource/resourcepool"
)

type UsedCores struct {
	HostId       string
	UsedCpuCores int
}

type UsedMem struct {
	HostId     string
	UsedMemory int
}

type UsedDisks struct {
	HostId string
	DiskId string
}

func (rw *GormResourceReadWrite) GetUsedCpuCores(ctx context.Context, hostIds []string) (resultFromHostTable, resultFromUsedTable, resultFromInstTable map[string]int, err error) {
	tx := rw.DB(ctx).Begin()
	var result1 []UsedCores
	err = tx.Model(&rp.Host{}).Select("id as host_id, cpu_cores - free_cpu_cores as used_cpu_cores").Where("id in ?", hostIds).Scan(&result1).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get used cores from hosts %v error, %v", hostIds, err)
	}

	var result2 []UsedCores
	err = tx.Model(&mm.UsedCompute{}).Select("host_id, sum(cpu_cores) as used_cpu_cores").
		Group("host_id").Having("host_id in ?", hostIds).Scan(&result2).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get used cores from used_computes %v error, %v", hostIds, err)
	}

	var result3 []UsedCores
	err = tx.Model(&cl.ClusterInstance{}).Select("host_id, sum(cpu_cores) as used_cpu_cores").
		Group("host_id").Having("host_id in ?", hostIds).Scan(&result3).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get used cores from cluster_instances %v error, %v", hostIds, err)
	}
	tx.Commit()

	resultFromHostTable = make(map[string]int)
	resultFromUsedTable = make(map[string]int)
	resultFromInstTable = make(map[string]int)
	for i := range result1 {
		resultFromHostTable[result1[i].HostId] = result1[i].UsedCpuCores
	}
	for i := range result2 {
		resultFromUsedTable[result2[i].HostId] = result2[i].UsedCpuCores
	}
	for i := range result3 {
		resultFromInstTable[result3[i].HostId] = result3[i].UsedCpuCores
	}

	return
}

func (rw *GormResourceReadWrite) GetUsedMemory(ctx context.Context, hostIds []string) (resultFromHostTable, resultFromUsedTable, resultFromInstTable map[string]int, err error) {
	tx := rw.DB(ctx).Begin()
	var result1 []UsedMem
	err = tx.Model(&rp.Host{}).Select("id as host_id, memory - free_memory as used_memory").Where("id in ?", hostIds).Scan(&result1).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get memory from hosts %v error, %v", hostIds, err)
	}

	var result2 []UsedMem
	err = tx.Model(&mm.UsedCompute{}).Select("host_id, sum(memory) as used_memory").
		Group("host_id").Having("host_id in ?", hostIds).Scan(&result2).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get used memory from used_computes %v error, %v", hostIds, err)
	}

	var result3 []UsedMem
	err = tx.Model(&cl.ClusterInstance{}).Select("host_id, sum(memory) as used_memory").
		Group("host_id").Having("host_id in ?", hostIds).Scan(&result3).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get used memory from cluster_instances %v error, %v", hostIds, err)
	}
	tx.Commit()

	resultFromHostTable = make(map[string]int)
	resultFromUsedTable = make(map[string]int)
	resultFromInstTable = make(map[string]int)
	for i := range result1 {
		resultFromHostTable[result1[i].HostId] = result1[i].UsedMemory
	}
	for i := range result2 {
		resultFromUsedTable[result2[i].HostId] = result2[i].UsedMemory
	}
	for i := range result3 {
		resultFromInstTable[result3[i].HostId] = result3[i].UsedMemory
	}
	return
}

func (rw *GormResourceReadWrite) GetUsedDisks(ctx context.Context, hostIds []string) (resultFromHostTable, resultFromUsedTable, resultFromInstTable map[string]*[]string, err error) {
	tx := rw.DB(ctx).Begin()
	var result1 []UsedDisks
	err = tx.Model(&rp.Disk{}).Select("host_id, id as disk_id").Where("status = ?", constants.DiskExhaust).
		Group("host_id").Group("id").Having("host_id in ?", hostIds).Order("host_id").Order("id").Scan(&result1).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get used disks from disks %v error, %v", hostIds, err)
	}

	var result2 []UsedDisks
	err = tx.Model(&mm.UsedDisk{}).Select("host_id, disk_id").Where("host_id in ?", hostIds).Order("host_id").Order("disk_id").Scan(&result2).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get used disks from used_disks %v error, %v", hostIds, err)
	}

	var result3 []UsedDisks
	err = tx.Model(&cl.ClusterInstance{}).Select("host_id, disk_id").Group("host_id").Group("disk_id").
		Having("host_id in ?", hostIds).Order("host_id").Order("disk_id").Scan(&result3).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get used disks from cluster_instances %v error, %v", hostIds, err)
	}
	tx.Commit()

	resultFromHostTable = rw.buildDiskMapByArr(result1)
	resultFromUsedTable = rw.buildDiskMapByArr(result2)
	resultFromInstTable = rw.buildDiskMapByArr(result3)

	return
}

func (rw *GormResourceReadWrite) buildDiskMapByArr(items []UsedDisks) (result map[string]*[]string) {
	result = make(map[string]*[]string)
	for i := range items {
		if ptr, ok := result[items[i].HostId]; ok {
			*ptr = append(*ptr, items[i].DiskId)
		} else {
			diskArr := []string{items[i].DiskId}
			result[items[i].HostId] = &diskArr
		}
	}
	return
}

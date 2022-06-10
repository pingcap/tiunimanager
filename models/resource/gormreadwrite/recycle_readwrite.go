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

package gormreadwrite

import (
	"context"

	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/common/errors"
	resource_structs "github.com/pingcap-inc/tiunimanager/micro-cluster/resourcemanager/management/structs"
	mm "github.com/pingcap-inc/tiunimanager/models/resource/management"
	rp "github.com/pingcap-inc/tiunimanager/models/resource/resourcepool"
	"gorm.io/gorm"
)

func (rw *GormResourceReadWrite) RecycleResources(ctx context.Context, request *resource_structs.RecycleRequest) (err error) {
	tx := rw.DB(ctx).Begin()
	for i, require := range request.RecycleReqs {
		err = rw.recycleForSingleRequire(ctx, tx, &require)
		if err != nil {
			tx.Rollback()
			return errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_RECYCLE_ERROR, "recycle resources failed on request %d, %v", i, err)
		}
	}
	tx.Commit()
	return
}

func (rw *GormResourceReadWrite) recycleForSingleRequire(ctx context.Context, tx *gorm.DB, req *resource_structs.RecycleRequire) (err error) {
	switch resource_structs.RecycleType(req.RecycleType) {
	case resource_structs.RecycleHolder:
		return rw.recycleHolderResource(ctx, tx, req.HolderID)
	case resource_structs.RecycleOperate:
		return rw.recycleResourceForRequest(ctx, tx, req.RequestID)
	case resource_structs.RecycleHost:
		return rw.recycleHostResource(ctx, tx, req.HolderID, req.RequestID, req.HostID, &req.ComputeReq, req.DiskReq, req.PortReq)
	default:
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "invalid recycle resource type %d", req.RecycleType)
	}
}

type ComputeStatistic struct {
	HostId        string
	TotalCpuCores int
	TotalMemory   int
}

func (rw *GormResourceReadWrite) recycleHolderResource(ctx context.Context, tx *gorm.DB, holderId string) (err error) {
	if len(holderId) == 0 {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "empty holderId")
	}
	var usedCompute []ComputeStatistic
	err = tx.Model(&mm.UsedCompute{}).Select("host_id, sum(cpu_cores) as total_cpu_cores, sum(memory) as total_memory").Where("holder_id = ?", holderId).Group("host_id").Scan(&usedCompute).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get cluster %s total used compute failed, %v", holderId, err)
	}

	var usedDisks []string
	err = tx.Model(&mm.UsedDisk{}).Select("disk_id").Where("holder_id = ?", holderId).Scan(&usedDisks).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get cluster %s total used disks failed, %v", holderId, err)
	}

	// update stat for hosts and disks
	err = rw.recycleResourcesInHosts(ctx, tx, usedCompute, usedDisks)
	if err != nil {
		return err
	}

	// Drop used resources in used_computes/used_disks/used_ports
	err = rw.recycleUsedTablesByHolder(ctx, tx, holderId)
	if err != nil {
		return err
	}

	return nil
}

func (rw *GormResourceReadWrite) recycleResourceForRequest(ctx context.Context, tx *gorm.DB, requestId string) (err error) {
	var usedCompute []ComputeStatistic
	err = tx.Model(&mm.UsedCompute{}).Select("host_id, sum(cpu_cores) as total_cpu_cores, sum(memory) as total_memory").Where("request_id = ?", requestId).Group("host_id").Scan(&usedCompute).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get request %s total used compute failed, %v", requestId, err)
	}

	var usedDisks []string
	err = tx.Model(&mm.UsedDisk{}).Select("disk_id").Where("request_id = ?", requestId).Scan(&usedDisks).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get request %s total used disks failed, %v", requestId, err)
	}

	// update stat for hosts and disks
	err = rw.recycleResourcesInHosts(ctx, tx, usedCompute, usedDisks)
	if err != nil {
		return err
	}

	// Drop used resources in used_computes/used_disks/used_ports
	err = rw.recycleUsedTablesByRequestId(ctx, tx, requestId)
	if err != nil {
		return err
	}

	return nil
}

func (rw *GormResourceReadWrite) recycleHostResource(ctx context.Context, tx *gorm.DB, clusterId string, requestId string, hostId string, computeReq *resource_structs.ComputeRequirement, diskReq []resource_structs.DiskResource, portReqs []resource_structs.PortResource) (err error) {
	var usedCompute []ComputeStatistic
	var usedDisks []string
	var usedPorts []int32
	if computeReq != nil {
		usedCompute = append(usedCompute, ComputeStatistic{
			HostId:        hostId,
			TotalCpuCores: int(computeReq.CpuCores),
			TotalMemory:   int(computeReq.Memory),
		})
	}

	for _, diskResouce := range diskReq {
		usedDisks = append(usedDisks, diskResouce.DiskId)
	}

	// Update Host Status and Disk Status
	err = rw.recycleResourcesInHosts(ctx, tx, usedCompute, usedDisks)
	if err != nil {
		return err
	}

	for _, portReq := range portReqs {
		usedPorts = append(usedPorts, portReq.Ports...)
	}

	err = rw.recycleUsedTablesBySpecify(ctx, tx, clusterId, requestId, hostId, computeReq.CpuCores, computeReq.Memory, usedDisks, usedPorts)
	if err != nil {
		return err
	}

	return nil
}

func (rw *GormResourceReadWrite) recycleResourcesInHosts(ctx context.Context, tx *gorm.DB, usedCompute []ComputeStatistic, usedDisks []string) (err error) {
	for _, diskId := range usedDisks {
		var disk rp.Disk
		tx.First(&disk, "ID = ?", diskId).First(&disk)
		if constants.DiskStatus(disk.Status).IsExhaust() {
			err = tx.Model(&disk).Update("Status", string(constants.DiskAvailable)).Error
			if err != nil {
				return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "update disk(%s) status while recycle failed, %v", diskId, err)
			}
		} else {
			return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "disk %s status not expected(%s) while recycle", diskId, disk.Status)
		}
	}
	for _, usedCompute := range usedCompute {
		var host rp.Host
		tx.First(&host, "ID = ?", usedCompute.HostId)
		host.FreeCpuCores += int32(usedCompute.TotalCpuCores)
		host.FreeMemory += int32(usedCompute.TotalMemory)

		if host.IsLoadless() {
			host.Stat = string(constants.HostLoadLoadLess)
		} else {
			host.Stat = string(constants.HostLoadInUsed)
		}

		err = tx.Model(&host).Select("FreeCpuCores", "FreeMemory", "Stat").Where("id = ?", usedCompute.HostId).Updates(rp.Host{FreeCpuCores: host.FreeCpuCores, FreeMemory: host.FreeMemory, Stat: host.Stat}).Error
		if err != nil {
			return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "update host(%s) stat while recycle failed, %v", usedCompute.HostId, err)
		}
	}
	return nil
}

func (rw *GormResourceReadWrite) recycleUsedTablesByHolder(ctx context.Context, tx *gorm.DB, holderId string) (err error) {
	err = tx.Where("holder_id = ?", holderId).Delete(&mm.UsedCompute{}).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "recycle UsedCompute for cluster %s failed, %v", holderId, err)
	}

	err = tx.Where("holder_id = ?", holderId).Delete(&mm.UsedDisk{}).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "recycle UsedDisk for cluster %s failed, %v", holderId, err)
	}

	err = tx.Where("holder_id = ?", holderId).Delete(&mm.UsedPort{}).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "recycle UsedPort for cluster %s failed, %v", holderId, err)
	}
	return nil
}

func (rw *GormResourceReadWrite) recycleUsedTablesByRequestId(ctx context.Context, tx *gorm.DB, requestId string) (err error) {
	err = tx.Where("request_id = ?", requestId).Delete(&mm.UsedCompute{}).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "recycle UsedCompute for request %s failed, %v", requestId, err)
	}

	err = tx.Where("request_id = ?", requestId).Delete(&mm.UsedDisk{}).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "recycle UsedDisk for request %s failed, %v", requestId, err)
	}

	err = tx.Where("request_id = ?", requestId).Delete(&mm.UsedPort{}).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "recycle UsedPort for request %s failed, %v", requestId, err)
	}
	return nil
}

func (rw *GormResourceReadWrite) recycleUsedTablesBySpecify(ctx context.Context, tx *gorm.DB, holderId, requestId, hostId string, cpuCores int32, memory int32, diskIds []string, ports []int32) (err error) {
	recycleCompute := mm.UsedCompute{
		HostId:   hostId,
		CpuCores: -cpuCores,
		Memory:   -memory,
	}
	recycleCompute.Holder.HolderId = holderId
	recycleCompute.RequestId = requestId
	err = tx.Create(&recycleCompute).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "recycle host(%s) to used_computes table failed: %v", hostId, err)
	}
	var usedCompute ComputeStatistic
	err = tx.Model(&mm.UsedCompute{}).Select("host_id, sum(cpu_cores) as total_cpu_cores, sum(memory) as total_memory").Where(
		"holder_id = ?", holderId).Group("host_id").Having("host_id = ?", hostId).Scan(&usedCompute).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "get host %s total used compute for cluster %s failed, %v", hostId, holderId, err)
	}
	if usedCompute.TotalCpuCores == 0 && usedCompute.TotalMemory == 0 {
		err = tx.Where("host_id = ? and holder_id = ?", hostId, holderId).Delete(&mm.UsedCompute{}).Error
		if err != nil {
			return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "clean up UsedCompute in host %s for cluster %s failed, %v", hostId, holderId, err)
		}
	}

	for _, diskId := range diskIds {
		err = tx.Where("host_id = ? and holder_id = ? and disk_id = ?", hostId, holderId, diskId).Delete(&mm.UsedDisk{}).Error
		if err != nil {
			return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "recycle UsedDisk for disk %s in host %s failed, %v", diskId, hostId, err)
		}
	}

	for _, port := range ports {
		err = tx.Where("host_id = ? and holder_id = ? and port = ?", hostId, holderId, port).Delete(&mm.UsedPort{}).Error
		if err != nil {
			return errors.NewErrorf(errors.TIUNIMANAGER_SQL_ERROR, "recycle UsedPort for %d in host %s failed, %v", port, hostId, err)
		}
	}
	return nil
}

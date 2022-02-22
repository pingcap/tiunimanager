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

package hostInspector

import (
	"context"
	"reflect"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/resource"
)

type HostInspect struct {
	resouceRW resource.ReaderWriter
}

func NewHostInspector() HostInspector {
	hostInspector := new(HostInspect)
	hostInspector.resouceRW = models.GetResourceReaderWriter()
	return hostInspector
}

func (p *HostInspect) SetResourceReaderWriter(rw resource.ReaderWriter) {
	p.resouceRW = rw
}

func (p *HostInspect) CheckCpuCores(ctx context.Context, host *structs.HostInfo) (result *structs.CheckInt32, err error) {
	return
}
func (p *HostInspect) CheckMemorySize(ctx context.Context, host *structs.HostInfo) (result *structs.CheckInt32, err error) {
	return
}
func (p *HostInspect) CheckDiskSize(ctx context.Context, host *structs.HostInfo) (inconsistDisks map[string]*structs.CheckInt32, err error) {
	return
}
func (p *HostInspect) CheckDiskRatio(ctx context.Context, host *structs.HostInfo) (inconsistDisks map[string]*structs.CheckInt32, err error) {
	return
}

func (p *HostInspect) CheckCpuAllocated(ctx context.Context, hosts []structs.HostInfo) (result map[string]*structs.CheckInt32, err error) {
	log := framework.LogWithContext(ctx)
	var hostIds []string
	for i := range hosts {
		hostIds = append(hostIds, hosts[i].ID)
	}
	resultFromHostTable, resultFromUsedTable, resultFromInstTable, err := p.resouceRW.GetUsedCpuCores(ctx, hostIds)
	if err != nil {
		log.Errorf("get used cpu cores for hosts %v from db failed, %v", hostIds, err)
		return nil, err
	}
	if !reflect.DeepEqual(resultFromHostTable, resultFromUsedTable) {
		// should be a allocation bug in resource module
		result = p.generateCheckInt32Result(hostIds, resultFromHostTable, resultFromUsedTable)
		for k, v := range result {
			if !(*v).Valid {
				log.Errorf("used cores mismatch between hosts table and used_compute on host %s, expected %d, got %d", k, (*v).ExpectedValue, (*v).RealValue)
			}
		}
		return result, errors.NewError(errors.TIEM_RESOURCE_CHECK_COMPUTES_ERROR, "used cores mismatch between hosts table and used_compute")
	}
	result = p.generateCheckInt32Result(hostIds, resultFromHostTable, resultFromInstTable)
	if !reflect.DeepEqual(resultFromHostTable, resultFromInstTable) {
		// should be a resource leak between resource module and cluster module
		for k, v := range result {
			if !(*v).Valid {
				log.Errorf("used cores mismatch between resource module and cluster module on host %s, expected %d, got %d", k, (*v).ExpectedValue, (*v).RealValue)
			}
		}
		return result, errors.NewError(errors.TIEM_RESOURCE_CHECK_COMPUTES_ERROR, "used cores mismatch between resource module and cluster module")
	}
	return
}
func (p *HostInspect) CheckMemAllocated(ctx context.Context, hosts []structs.HostInfo) (result map[string]*structs.CheckInt32, err error) {
	log := framework.LogWithContext(ctx)
	var hostIds []string
	for i := range hosts {
		hostIds = append(hostIds, hosts[i].ID)
	}
	resultFromHostTable, resultFromUsedTable, resultFromInstTable, err := p.resouceRW.GetUsedMemory(ctx, hostIds)
	if err != nil {
		log.Errorf("get used memory for hosts %v from db failed, %v", hostIds, err)
		return nil, err
	}
	if !reflect.DeepEqual(resultFromHostTable, resultFromUsedTable) {
		// should be a allocation bug in resource module
		result = p.generateCheckInt32Result(hostIds, resultFromHostTable, resultFromUsedTable)
		for k, v := range result {
			if !(*v).Valid {
				log.Errorf("used memory mismatch between hosts table and used_compute on host %s, expected %d, got %d", k, (*v).ExpectedValue, (*v).RealValue)
			}
		}
		return result, errors.NewError(errors.TIEM_RESOURCE_CHECK_COMPUTES_ERROR, "used memory mismatch between hosts table and used_compute")
	}
	result = p.generateCheckInt32Result(hostIds, resultFromHostTable, resultFromInstTable)
	if !reflect.DeepEqual(resultFromHostTable, resultFromInstTable) {
		// should be a resource leak between resource module and cluster module
		for k, v := range result {
			if !(*v).Valid {
				log.Errorf("used memory mismatch between resource module and cluster module on host %s, expected %d, got %d", k, (*v).ExpectedValue, (*v).RealValue)
			}
		}
		return result, errors.NewError(errors.TIEM_RESOURCE_CHECK_COMPUTES_ERROR, "used memory mismatch between resource module and cluster module")
	}
	return
}

func (p *HostInspect) CheckDiskAllocated(ctx context.Context, hosts []structs.HostInfo) (result map[string]map[string]*structs.CheckString, err error) {
	log := framework.LogWithContext(ctx)
	var hostIds []string
	for i := range hosts {
		hostIds = append(hostIds, hosts[i].ID)
	}
	resultFromHostTable, resultFromUsedTable, resultFromInstTable, err := p.resouceRW.GetUsedDisks(ctx, hostIds)
	if err != nil {
		log.Errorf("get used memory for hosts %v from db failed, %v", hostIds, err)
		return nil, err
	}
	if !reflect.DeepEqual(resultFromHostTable, resultFromUsedTable) {
		// should be a allocation bug in resource module
		result = p.generateCheckStatusResult(hostIds, resultFromHostTable, resultFromUsedTable)
		for hostId, disks := range result {
			for diskId, status := range disks {
				if !status.Valid {
					log.Errorf("used disk status mismatch between hosts table and used_compute on host %s, disk %s, expected %s, got %s",
						hostId, diskId, status.ExpectedValue, status.RealValue)
				}
			}
		}
		return result, errors.NewError(errors.TIEM_RESOURCE_CHECK_DISKS_ERROR, "used disks mismatch between hosts table and used_compute")
	}
	result = p.generateCheckStatusResult(hostIds, resultFromHostTable, resultFromInstTable)
	if !reflect.DeepEqual(resultFromHostTable, resultFromInstTable) {
		// should be a resource leak between resource module and cluster module
		for hostId, disks := range result {
			for diskId, status := range disks {
				if !status.Valid {
					log.Errorf("used disk status mismatch between hosts table and used_compute on host %s, disk %s, expected %s, got %s",
						hostId, diskId, status.ExpectedValue, status.RealValue)
				}
			}
		}
		return result, errors.NewError(errors.TIEM_RESOURCE_CHECK_DISKS_ERROR, "used memory mismatch between resource module and cluster module")
	}

	return
}

func (p *HostInspect) generateCheckInt32Result(hostIds []string, map1 map[string]int, map2 map[string]int) (result map[string]*structs.CheckInt32) {
	const invalidCount = -1
	result = make(map[string]*structs.CheckInt32)
	for _, hostId := range hostIds {
		v1, ok1 := map1[hostId]
		check := new(structs.CheckInt32)
		if ok1 {
			check.ExpectedValue = int32(v1)
		} else {
			check.ExpectedValue = invalidCount
		}
		v2, ok2 := map2[hostId]
		if ok2 {
			check.RealValue = int32(v2)
		} else {
			check.RealValue = invalidCount
		}
		if check.ExpectedValue == check.RealValue {
			check.Valid = true
		} else {
			check.Valid = false
		}
		result[hostId] = check
	}
	return
}

func (p *HostInspect) generateCheckStatusResult(hostIds []string, map1, map2 map[string]*[]string) (result map[string]map[string]*structs.CheckString) {
	result = make(map[string]map[string]*structs.CheckString)

	for _, hostId := range hostIds {
		v1, ok1 := map1[hostId]
		if ok1 {
			disksStatus, ok := result[hostId]
			if !ok {
				result[hostId] = make(map[string]*structs.CheckString)
				disksStatus = result[hostId]
			}
			for _, diskId := range *v1 {
				check, diskExist := disksStatus[diskId]
				if diskExist {
					check.ExpectedValue = string(constants.DiskExhaust)
					if check.ExpectedValue == check.RealValue {
						check.Valid = true
					}
				} else {
					check = new(structs.CheckString)
					check.Valid = false
					check.ExpectedValue = string(constants.DiskExhaust)
					disksStatus[diskId] = check
				}
			}
		}
		v2, ok2 := map2[hostId]
		if ok2 {
			disksStatus, ok := result[hostId]
			if !ok {
				result[hostId] = make(map[string]*structs.CheckString)
				disksStatus = result[hostId]
			}
			for _, diskId := range *v2 {
				check, diskExist := disksStatus[diskId]
				if diskExist {
					check.RealValue = string(constants.DiskExhaust)
					if check.ExpectedValue == check.RealValue {
						check.Valid = true
					}
				} else {
					check = new(structs.CheckString)
					check.Valid = false
					check.RealValue = string(constants.DiskExhaust)
					disksStatus[diskId] = check
				}
			}

		}
	}
	return
}

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

package allocrecycle

import (
	"context"
	"fmt"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/resource"
	"github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"github.com/pingcap-inc/tiem/util/aws"
)

type AwsManagement struct {
	rw resource.ReaderWriter
}

func NewAwsManagement() structs.AllocatorRecycler {
	awsManagement := new(AwsManagement)
	awsManagement.rw = models.GetResourceReaderWriter()
	return awsManagement
}

func (m *AwsManagement) SetResourceReaderWriter(rw resource.ReaderWriter) {
	m.rw = rw
}

func setupPortResult(portRequires []structs.PortRequirement) (portResults []structs.PortResource) {
	for _, portRequire := range portRequires {
		var portResult structs.PortResource
		portResult.Start = portRequire.Start
		portResult.End = portRequire.End
		for i := 0; i < int(portRequire.PortCnt); i++ {
			portResult.Ports = append(portResult.Ports, portRequire.Start+int32(i))
		}
		portResults = append(portResults, portResult)
	}
	return
}

func (m *AwsManagement) recordHosts(ctx context.Context, hosts []resourcepool.Host) (hostIds []string, err error) {
	framework.LogWithContext(ctx).Infof("record hosts %v", hosts)
	return m.rw.Create(ctx, hosts)
}

func (m *AwsManagement) AllocResources(ctx context.Context, batchReq *structs.BatchAllocRequest) (results *structs.BatchAllocResponse, err error) {
	var newHosts []resourcepool.Host
	results = new(structs.BatchAllocResponse)
	for _, allocReq := range batchReq.BatchRequests {
		var allocRsp structs.AllocRsp
		allocRsp.Applicant.HolderId = allocReq.Applicant.HolderId
		allocRsp.Applicant.RequestId = allocReq.Applicant.RequestId
		allocRsp.Applicant.TakeoverOperation = allocReq.Applicant.TakeoverOperation
		for i, require := range allocReq.Requires {
			framework.LogWithContext(ctx).Infof("call cloud sdk for %d %v %v %v for %d hosts", require.Strategy, require.Location, require.Require.ComputeReq, require.Require.DiskReq, require.Count)
			if require.Strategy == structs.ClusterPorts {
				var compute structs.Compute
				compute.Reqseq = int32(i)
				compute.PortRes = setupPortResult(require.Require.PortReq)
				allocRsp.Results = append(allocRsp.Results, compute)
				continue
			}
			computes, err := aws.CreateCloudInstances(require.Location.Zone, require.Count)
			if err != nil {
				errMsg := fmt.Sprintf("create %v %v %v aws %d vm failed, %v", require.Location, require.Require.ComputeReq, require.Require.DiskReq, require.Count, err)
				framework.LogWithContext(ctx).Errorln(errMsg)
				return nil, errors.NewError(errors.TIEM_RESOURCE_CREATE_AWS_HOST_ERROR, errMsg)
			}
			for j := range computes {
				computes[j].Reqseq = int32(i)
				computes[j].Location = require.Location
				computes[j].ComputeRes.CpuCores = require.Require.ComputeReq.CpuCores
				computes[j].ComputeRes.Memory = require.Require.ComputeReq.Memory
				computes[j].DiskRes.DiskId = "AWS-Fake-Disk"
				computes[j].DiskRes.DiskName = "vda"
				computes[j].DiskRes.Path = "/"
				computes[j].DiskRes.Type = require.Require.DiskReq.DiskType
				computes[j].DiskRes.Capacity = require.Require.DiskReq.Capacity
				computes[j].PortRes = setupPortResult(require.Require.PortReq)
				computes[j].Passwd = ""

				host := toModelHost(computes[j])
				host.Region = require.Location.Region
				host.AZ = require.Location.Zone
				host.Rack = require.Location.Rack
				framework.LogWithContext(ctx).Infof("Get a new host %v", *host)
				newHosts = append(newHosts, *host)
			}
			allocRsp.Results = append(allocRsp.Results, computes...)
		}
		results.BatchResults = append(results.BatchResults, &allocRsp)
		framework.LogWithContext(ctx).Infof("Append Response %v", allocRsp)
	}

	hostIds, err := m.recordHosts(ctx, newHosts)
	if err != nil {
		errMsg := fmt.Sprintf("record %v to db failed, %v", newHosts, err)
		framework.LogWithContext(ctx).Errorln(errMsg)
		return nil, errors.NewError(errors.TIEM_RESOURCE_RECORD_CLOUD_HOST_ERROR, errMsg)
	}
	framework.LogWithContext(ctx).Infof("Alloc Cloud Host %v Succeed", hostIds)

	return results, nil
}
func (m *AwsManagement) RecycleResources(ctx context.Context, request *structs.RecycleRequest) (err error) {
	return m.rw.RecycleResources(ctx, request)
}

func toModelHost(c structs.Compute) (h *resourcepool.Host) {
	var host resourcepool.Host
	host.HostName = c.HostName
	host.IP = c.HostIp
	host.ID = c.HostId
	host.CpuCores = c.ComputeRes.CpuCores
	host.Memory = c.ComputeRes.Memory
	host.FreeCpuCores = 0
	host.FreeMemory = 0
	host.Status = string(constants.HostOnline)
	host.Stat = string(constants.HostLoadInUsed)
	host.Disks = append(host.Disks, resourcepool.Disk{
		ID:       c.DiskRes.DiskId,
		Path:     c.DiskRes.Path,
		Name:     c.DiskRes.DiskName,
		Type:     c.DiskRes.Type,
		Capacity: c.DiskRes.Capacity,
		Status:   string(constants.DiskExhaust),
	})
	return &host
}

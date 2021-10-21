
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

package hostresource

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	crypto "github.com/pingcap-inc/tiem/library/thirdparty/encrypt"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/micro-api/controller"

	"google.golang.org/grpc/codes"
)

func copyHostFromRsp(src *clusterpb.HostInfo, dst *HostInfo) {
	dst.ID = src.HostId
	dst.HostName = src.HostName
	dst.IP = src.Ip
	dst.Arch = src.Arch
	dst.OS = src.Os
	dst.Kernel = src.Kernel
	dst.FreeCpuCores = src.FreeCpuCores
	dst.FreeMemory = src.FreeMemory
	dst.Spec = src.Spec
	dst.CpuCores = src.CpuCores
	dst.Memory = src.Memory
	dst.Nic = src.Nic
	dst.Region = src.Region
	dst.AZ = src.Az
	dst.Rack = src.Rack
	dst.Status = src.Status
	dst.Stat = src.Stat
	dst.Purpose = src.Purpose
	dst.DiskType = src.DiskType
	dst.CreatedAt = src.CreateAt
	dst.Reserved = src.Reserved
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, DiskInfo{
			ID:       disk.DiskId,
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   int32(disk.Status),
			Type:     disk.Type,
		})
	}
}

func genHostSpec(cpuCores int32, mem int32) string {
	return fmt.Sprintf("%dC%dG", cpuCores, mem)
}

func copyHostToReq(src *HostInfo, dst *clusterpb.HostInfo) error {
	dst.HostName = src.HostName
	dst.Ip = src.IP
	dst.UserName = src.UserName
	passwd, err := crypto.AesEncryptCFB(src.Passwd)
	if err != nil {
		return err
	}
	dst.Passwd = passwd
	dst.Arch = src.Arch
	dst.Os = src.OS
	dst.Kernel = src.Kernel
	dst.FreeCpuCores = src.FreeCpuCores
	dst.FreeMemory = src.FreeMemory
	dst.Spec = genHostSpec(src.CpuCores, src.Memory)
	dst.CpuCores = src.CpuCores
	dst.Memory = src.Memory
	dst.Nic = src.Nic
	dst.Region = src.Region
	dst.Az = src.AZ
	dst.Rack = src.Rack
	dst.Status = src.Status
	dst.Stat = src.Stat
	dst.Purpose = src.Purpose
	dst.DiskType = src.DiskType
	dst.Reserved = src.Reserved

	for _, v := range src.Disks {
		dst.Disks = append(dst.Disks, &clusterpb.Disk{
			Name:     v.Name,
			Capacity: v.Capacity,
			Status:   v.Status,
			Path:     v.Path,
			Type:     v.Type,
		})
	}
	return nil
}

func doImport(c *gin.Context, host *HostInfo) (rsp *clusterpb.ImportHostResponse, err error) {
	importReq := clusterpb.ImportHostRequest{}
	importReq.Host = new(clusterpb.HostInfo)
	err = copyHostToReq(host, importReq.Host)
	if err != nil {
		return nil, err
	}
	return client.ClusterClient.ImportHost(c, &importReq)
}

func doImportBatch(c *gin.Context, hosts []*HostInfo) (rsp *clusterpb.ImportHostsInBatchResponse, err error) {
	importReq := clusterpb.ImportHostsInBatchRequest{}
	importReq.Hosts = make([]*clusterpb.HostInfo, len(hosts))
	var userName, passwd string
	for i, host := range hosts {
		if i == 0 {
			userName, passwd = host.UserName, host.Passwd
		} else {
			if userName != host.UserName || passwd != host.Passwd {
				errMsg := fmt.Sprintf("Row %d has a diff user(%s) or passwd(%s)", i, host.UserName, host.Passwd)
				return nil, errors.New(errMsg)
			}
		}
		importReq.Hosts[i] = new(clusterpb.HostInfo)
		err = copyHostToReq(host, importReq.Hosts[i])
		if err != nil {
			return nil, err
		}
	}

	return client.ClusterClient.ImportHostsInBatch(c, &importReq)
}

// ImportHost godoc
// @Summary Import a host to TiEM System
// @Description import one host by json
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param host body HostInfo true "Host information"
// @Success 200 {object} controller.CommonResult{data=string}
// @Router /resources/host [post]
func ImportHost(c *gin.Context) {
	var host HostInfo
	if err := c.ShouldBindJSON(&host); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	for i := range host.Disks {
		if host.Disks[i].Type == "" {
			host.Disks[i].Type = string(resource.Sata)
		}
	}

	rsp, err := doImport(c, &host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}

	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}

	c.JSON(http.StatusOK, controller.Success(ImportHostRsp{HostId: rsp.HostId}))
}

func importExcelFile(r io.Reader, reserved bool) ([]*HostInfo, error) {
	xlsx, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	rows := xlsx.GetRows("Host Information")
	var hosts []*HostInfo
	for irow, row := range rows {
		if irow > 0 {
			var host HostInfo
			host.Reserved = reserved
			host.HostName = row[HOSTNAME_FIELD]
			addr := net.ParseIP(row[IP_FILED])
			if addr == nil {
				errMsg := fmt.Sprintf("Row %d has a Invalid IP Address %s", irow, row[IP_FILED])
				return nil, errors.New(errMsg)
			}
			host.IP = addr.String()
			host.UserName = row[USERNAME_FIELD]
			host.Passwd = row[PASSWD_FIELD]
			host.Region = row[REGION_FIELD]
			host.AZ = row[ZONE_FIELD]
			host.Rack = row[RACK_FIELD]
			if err = resource.ValidArch(row[ARCH_FIELD]); err != nil {
				errMsg := fmt.Sprintf("Row %d get arch(%s) failed, %v", irow, row[ARCH_FIELD], err)
				return nil, errors.New(errMsg)
			}
			host.Arch = row[ARCH_FIELD]
			host.OS = row[OS_FIELD]
			host.Kernel = row[KERNEL_FIELD]
			coreNum, err := (strconv.Atoi(row[CPU_FIELD]))
			if err != nil {
				errMsg := fmt.Sprintf("Row %d get coreNum(%s) failed, %v", irow, row[CPU_FIELD], err)
				return nil, errors.New(errMsg)
			}
			host.CpuCores = int32(coreNum)
			host.FreeCpuCores = host.CpuCores
			mem, err := (strconv.Atoi(row[MEM_FIELD]))
			if err != nil {
				errMsg := fmt.Sprintf("Row %d get memory(%s) failed, %v", irow, row[MEM_FIELD], err)
				return nil, errors.New(errMsg)
			}
			host.Memory = int32(mem)
			host.FreeMemory = host.Memory
			host.Nic = row[NIC_FIELD]

			if err = resource.ValidPurposeType(row[PURPOSE_FIELD]); err != nil {
				errMsg := fmt.Sprintf("Row %d get purpose(%s) failed, %v", irow, row[PURPOSE_FIELD], err)
				return nil, errors.New(errMsg)
			}
			host.Purpose = row[PURPOSE_FIELD]
			if err = resource.ValidDiskType(row[DISKTYPE_FIELD]); err != nil {
				errMsg := fmt.Sprintf("Row %d get disk type(%s) failed, %v", irow, row[DISKTYPE_FIELD], err)
				return nil, errors.New(errMsg)
			}
			host.DiskType = row[DISKTYPE_FIELD]
			disksStr := row[DISKS_FIELD]
			if err = json.Unmarshal([]byte(disksStr), &host.Disks); err != nil {
				errMsg := fmt.Sprintf("Row %d has a Invalid Disk Json Format, %v", irow, err)
				return nil, errors.New(errMsg)
			}
			for i := range host.Disks {
				if host.Disks[i].Type == "" {
					host.Disks[i].Type = string(resource.DiskType(host.DiskType))
				}
			}
			hosts = append(hosts, &host)
		}
	}
	return hosts, nil
}

// ImportHosts godoc
// @Summary Import a batch of hosts to TiEM
// @Description import hosts by xlsx file
// @Tags resource
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param hostReserved formData string true "whether hosts are reserved(won't be allocated) after import" default(false)
// @Param file formData file true "hosts information in a xlsx file"
// @Success 200 {object} controller.CommonResult{data=[]string}
// @Router /resources/hosts [post]
func ImportHosts(c *gin.Context) {
	reservedStr := c.PostForm("hostReserved")
	reserved, err := strconv.ParseBool(reservedStr)
	if err != nil {
		errmsg := fmt.Sprintf("GetFormData Error: %v", err)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		errmsg := fmt.Sprintf("GetFormFile Error: %v", err)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	hosts, err := importExcelFile(file, reserved)
	if err != nil {
		errmsg := fmt.Sprintf("Import File Error: %v", err)
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}

	rsp, err := doImportBatch(c, hosts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}

	c.JSON(http.StatusOK, controller.Success(ImportHostsRsp{HostIds: rsp.HostIds}))
}

// ListHost godoc
// @Summary Show all hosts list in TiEM
// @Description get hosts lit
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param hostQuery query HostQuery false "list condition"
// @Success 200 {object} controller.ResultWithPage{data=[]HostInfo}
// @Router /resources/hosts [get]
func ListHost(c *gin.Context) {
	hostQuery := HostQuery{
		Status: -1,
	}
	if err := c.ShouldBindQuery(&hostQuery); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	if !resource.HostStatus(hostQuery.Status).IsValid() {
		errmsg := fmt.Sprintf("Input Status %d is Invalid", hostQuery.Status)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}

	listHostReq := clusterpb.ListHostsRequest{
		Purpose: hostQuery.Purpose,
		Status:  int32(hostQuery.Status),
	}
	listHostReq.PageReq = new(clusterpb.PageDTO)
	listHostReq.PageReq.Page = int32(hostQuery.Page)
	listHostReq.PageReq.PageSize = int32(hostQuery.PageSize)

	rsp, err := client.ClusterClient.ListHost(c, &listHostReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}
	var res ListHostRsp
	for _, v := range rsp.HostList {
		var host HostInfo
		copyHostFromRsp(v, &host)
		res.Hosts = append(res.Hosts, host)
	}
	c.JSON(http.StatusOK, controller.SuccessWithPage(res.Hosts, controller.Page{Page: int(rsp.PageReq.Page), PageSize: int(rsp.PageReq.PageSize), Total: int(rsp.PageReq.Total)}))
}

// HostDetails godoc
// @Summary Show a host
// @Description get one host by id
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param hostId path string true "host ID"
// @Success 200 {object} controller.CommonResult{data=HostInfo}
// @Router /resources/hosts/{hostId} [get]
func HostDetails(c *gin.Context) {

	hostId := c.Param("hostId")

	HostDetailsReq := clusterpb.CheckDetailsRequest{
		HostId: hostId,
	}

	rsp, err := client.ClusterClient.CheckDetails(c, &HostDetailsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}
	var res HostDetailsRsp
	copyHostFromRsp(rsp.Details, &(res.Host))
	c.JSON(http.StatusOK, controller.Success(res))
}

// RemoveHost godoc
// @Summary Remove a host
// @Description remove a host by id
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param hostId path string true "host id"
// @Success 200 {object} controller.CommonResult{data=string}
// @Router /resources/hosts/{hostId} [delete]
func RemoveHost(c *gin.Context) {

	hostId := c.Param("hostId")

	RemoveHostReq := clusterpb.RemoveHostRequest{
		HostId: hostId,
	}

	rsp, err := client.ClusterClient.RemoveHost(c, &RemoveHostReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}
	c.JSON(http.StatusOK, controller.Success(rsp.Rs.Message))
}

func detectDuplicateElement(hostIds []string) (string, bool) {
	temp := map[string]struct{}{}
	hasDuplicate := false
	var duplicateStr string
	for _, item := range hostIds {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
		} else {
			hasDuplicate = true
			duplicateStr = item
			break
		}
	}
	return duplicateStr, hasDuplicate
}

// RemoveHosts godoc
// @Summary Remove a batch of hosts
// @Description remove hosts by a list
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param hostIds body []string true "list of host IDs"
// @Success 200 {object} controller.CommonResult{data=string}
// @Router /resources/hosts/ [delete]
func RemoveHosts(c *gin.Context) {

	var hostIds []string
	if err := c.ShouldBindJSON(&hostIds); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}

	if str, dup := detectDuplicateElement(hostIds); dup {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), str+" Is Duplicated in request"))
		return
	}

	RemoveHostsReq := clusterpb.RemoveHostsInBatchRequest{
		HostIds: hostIds,
	}

	rsp, err := client.ClusterClient.RemoveHostsInBatch(c, &RemoveHostsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}
	c.JSON(http.StatusOK, controller.Success(rsp.Rs.Message))
}

// DownloadHostTemplateFile godoc
// @Summary Download the host information template file for importing
// @Description get host template xlsx file
// @Tags resource
// @Accept json
// @Produce octet-stream
// @Security ApiKeyAuth
// @Success 200 {file} file
// @Router /resources/hosts-template/ [get]
func DownloadHostTemplateFile(c *gin.Context) {
	curDir, _ := os.Getwd()
	templateName := common.TemplateFileName
	// The template file should be on tiem/etc/hostInfo_template.xlsx
	filePath := filepath.Join(curDir, common.TemplateFilePath, templateName)

	_, err := os.Stat(filePath)
	if err != nil && !os.IsExist(err) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.NotFound), err.Error()))
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+templateName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	c.File(filePath)
}

func copyAllocToReq(src []Allocation, dst *[]*clusterpb.AllocationReq) {
	for _, req := range src {
		*dst = append(*dst, &clusterpb.AllocationReq{
			FailureDomain: req.FailureDomain,
			CpuCores:      req.CpuCores,
			Memory:        req.Memory,
			Count:         req.Count,
		})
	}
}

func copyAllocFromRsp(src []*clusterpb.AllocHost, dst *[]AllocateRsp) {
	for i, host := range src {
		plainPasswd, err := crypto.AesDecryptCFB(host.Passwd)
		if err != nil {
			// AllocHosts API is for internal testing, so just panic if something wrong
			panic(err)
		}
		*dst = append(*dst, AllocateRsp{
			HostName: host.HostName,
			Ip:       host.Ip,
			UserName: host.UserName,
			Passwd:   plainPasswd,
			CpuCores: host.CpuCores,
			Memory:   host.Memory,
		})
		(*dst)[i].Disk.ID = host.Disk.DiskId
		(*dst)[i].Disk.Name = host.Disk.Name
		(*dst)[i].Disk.Path = host.Disk.Path
		(*dst)[i].Disk.Capacity = host.Disk.Capacity
		(*dst)[i].Disk.Status = host.Disk.Status
	}
}

// AllocHosts godoc
// @Summary Alloc host/disk resources for creating tidb cluster
// @Description should be used in testing env
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Alloc body AllocHostsReq true "location and spec of hosts"
// @Success 200 {object} controller.CommonResult{data=AllocHostsRsp}
// @Router /resources/allochosts [post]
func AllocHosts(c *gin.Context) {
	var allocation AllocHostsReq
	if err := c.ShouldBindJSON(&allocation); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}

	allocReq := clusterpb.AllocHostsRequest{}
	copyAllocToReq(allocation.PdReq, &allocReq.PdReq)
	copyAllocToReq(allocation.TidbReq, &allocReq.TidbReq)
	copyAllocToReq(allocation.TikvReq, &allocReq.TikvReq)
	//fmt.Println(allocReq.PdReq, allocReq.TidbReq, allocReq.TikvReq)
	rsp, err := client.ClusterClient.AllocHosts(c, &allocReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}

	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}

	var res AllocHostsRsp
	copyAllocFromRsp(rsp.PdHosts, &res.PdHosts)
	copyAllocFromRsp(rsp.TidbHosts, &res.TidbHosts)
	copyAllocFromRsp(rsp.TikvHosts, &res.TikvHosts)

	c.JSON(http.StatusOK, controller.Success(res))
}

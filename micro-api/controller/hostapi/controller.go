package hostapi

import (
	"encoding/json"
	"errors"
	"fmt"
	client2 "github.com/pingcap-inc/tiem/library/client"
	crypto "github.com/pingcap-inc/tiem/library/thirdparty/encrypt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-cluster/proto"
	"github.com/pingcap-inc/tiem/micro-metadb/service"

	"google.golang.org/grpc/codes"
)

func CopyHostFromRsp(src *cluster.HostInfo, dst *HostInfo) {
	dst.HostId = src.HostId
	dst.HostName = src.HostName
	dst.Ip = src.Ip
	dst.Os = src.Os
	dst.Kernel = src.Kernel
	dst.CpuCores = int32(src.CpuCores)
	dst.Memory = int32(src.Memory)
	dst.Spec = src.Spec
	dst.Nic = src.Nic
	dst.Dc = src.Dc
	dst.Az = src.Az
	dst.Rack = src.Rack
	dst.Status = int32(src.Status)
	dst.Purpose = src.Purpose
	dst.CreatedAt = src.CreateAt
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, Disk{
			DiskId:   disk.DiskId,
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   int32(disk.Status),
		})
	}
}

func genHostSpec(cpuCores int32, mem int32) string {
	return fmt.Sprintf("%dC%dG", cpuCores, mem)
}

func copyHostToReq(src *HostInfo, dst *cluster.HostInfo) error {
	dst.HostName = src.HostName
	dst.Ip = src.Ip
	dst.UserName = src.UserName
	passwd, err := crypto.AesEncryptCFB(src.Passwd)
	if err != nil {
		return err
	}
	dst.Passwd = passwd
	dst.Os = src.Os
	dst.Kernel = src.Kernel
	dst.CpuCores = src.CpuCores
	dst.Memory = src.Memory
	dst.Spec = genHostSpec(src.CpuCores, src.Memory)
	dst.Nic = src.Nic
	dst.Dc = src.Dc
	dst.Az = src.Az
	dst.Rack = src.Rack
	dst.Status = src.Status
	dst.Purpose = src.Purpose

	for _, v := range src.Disks {
		dst.Disks = append(dst.Disks, &cluster.Disk{
			Name:     v.Name,
			Capacity: v.Capacity,
			Status:   v.Status,
			Path:     v.Path,
		})
	}
	return nil
}

func doImport(c *gin.Context, host *HostInfo) (rsp *cluster.ImportHostResponse, err error) {
	importReq := cluster.ImportHostRequest{}
	importReq.Host = new(cluster.HostInfo)
	err = copyHostToReq(host, importReq.Host)
	if err != nil {
		return nil, err
	}
	return client2.ClusterClient.ImportHost(c, &importReq)
}

func doImportBatch(c *gin.Context, hosts []*HostInfo) (rsp *cluster.ImportHostsInBatchResponse, err error) {
	importReq := cluster.ImportHostsInBatchRequest{}
	importReq.Hosts = make([]*cluster.HostInfo, len(hosts))
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
		importReq.Hosts[i] = new(cluster.HostInfo)
		err = copyHostToReq(host, importReq.Hosts[i])
		if err != nil {
			return nil, err
		}
	}

	return client2.ClusterClient.ImportHostsInBatch(c, &importReq)
}

// ImportHost 导入主机接口
// @Summary 导入主机接口
// @Description 将给定的主机信息导入系统
// @Tags resource
// @Accept json
// @Produce json
// @Param Token header string true "登录token"
// @Param host body HostInfo true "待导入的主机信息"
// @Success 200 {object} controller.CommonResult{data=string}
// @Router /resources/host [post]
func ImportHost(c *gin.Context) {
	var host HostInfo
	if err := c.ShouldBindJSON(&host); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
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

func importExcelFile(r io.Reader) ([]*HostInfo, error) {
	xlsx, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	rows := xlsx.GetRows("主机信息")
	var hosts []*HostInfo
	for irow, row := range rows {
		if irow > 0 {
			var host HostInfo
			host.HostName = row[HOSTNAME_FIELD]
			addr := net.ParseIP(row[IP_FILED])
			if addr == nil {
				errMsg := fmt.Sprintf("Row %d has a Invalid IP Address %s", irow, row[IP_FILED])
				return nil, errors.New(errMsg)
			}
			host.Ip = addr.String()
			host.UserName = row[USERNAME_FIELD]
			host.Passwd = row[PASSWD_FIELD]
			host.Dc = row[DC_FIELD]
			host.Az = row[ZONE_FIELD]
			host.Rack = row[RACK_FIELD]
			host.Os = row[OS_FIELD]
			host.Kernel = row[KERNEL_FIELD]
			coreNum, _ := (strconv.Atoi(row[CPU_FIELD]))
			host.CpuCores = int32(coreNum)
			mem, _ := (strconv.Atoi(row[MEM_FIELD]))
			host.Memory = int32(mem)
			host.Nic = row[NIC_FIELD]
			host.Purpose = row[PURPOSE_FIELD]
			disksStr := row[DISKS_FIELD]
			if err = json.Unmarshal([]byte(disksStr), &host.Disks); err != nil {
				errMsg := fmt.Sprintf("Row %d has a Invalid Disk Json Format, %v", irow, err)
				return nil, errors.New(errMsg)
			}
			hosts = append(hosts, &host)
		}
	}
	return hosts, nil
}

// ImportHosts 批量导入主机接口
// @Summary 通过文件批量导入主机
// @Description 通过文件批量导入主机
// @Tags resource
// @Accept mpfd
// @Produce json
// @Param Token header string true "登录token"
// @Param file formData file true "包含待导入主机信息的文件"
// @Success 200 {object} controller.CommonResult{data=[]string}
// @Router /resources/hosts [post]
func ImportHosts(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		errmsg := fmt.Sprintf("GetFormFile Error: %v", err)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	hosts, err := importExcelFile(file)
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

// ListHost 查询主机列表接口
// @Summary 查询主机列表
// @Description 展示目前所有主机
// @Tags resource
// @Accept json
// @Produce json
// @Param Token header string true "登录token"
// @Param hostQuery query HostQuery false "主机列表的查询条件"
// @Success 200 {object} controller.ResultWithPage{data=[]HostInfo}
// @Router /resources/hosts [get]
func ListHost(c *gin.Context) {
	var hostQuery HostQuery
	if err := c.ShouldBindQuery(&hostQuery); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	if !HostStatus(hostQuery.Status).IsValid() {
		errmsg := fmt.Sprintf("Input Status %d is Invalid", hostQuery.Status)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}

	listHostReq := cluster.ListHostsRequest{
		Purpose: hostQuery.Purpose,
		Status:  int32(hostQuery.Status),
	}
	listHostReq.PageReq = new(cluster.PageDTO)
	listHostReq.PageReq.Page = int32(hostQuery.Page)
	listHostReq.PageReq.PageSize = int32(hostQuery.PageSize)

	rsp, err := client2.ClusterClient.ListHost(c, &listHostReq)
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
		CopyHostFromRsp(v, &host)
		res.Hosts = append(res.Hosts, host)
	}
	c.JSON(http.StatusOK, controller.SuccessWithPage(res.Hosts, controller.Page{Page: int(rsp.PageReq.Page), PageSize: int(rsp.PageReq.PageSize), Total: int(rsp.PageReq.Total)}))
}

// HostDetails 查询主机详情接口
// @Summary 查询主机详情
// @Description 展示指定的主机的详细信息
// @Tags resource
// @Accept json
// @Produce json
// @Param Token header string true "登录token"
// @Param hostId path string true "主机ID"
// @Success 200 {object} controller.CommonResult{data=HostInfo}
// @Router /resources/hosts/{hostId} [get]
func HostDetails(c *gin.Context) {

	hostId := c.Param("hostId")

	HostDetailsReq := cluster.CheckDetailsRequest{
		HostId: hostId,
	}

	rsp, err := client2.ClusterClient.CheckDetails(c, &HostDetailsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}
	var res HostDetailsRsp
	CopyHostFromRsp(rsp.Details, &(res.Host))
	c.JSON(http.StatusOK, controller.Success(res))
}

// RemoveHost 删除主机接口
// @Summary 删除指定的主机
// @Description 删除指定的主机
// @Tags resource
// @Accept json
// @Produce json
// @Param Token header string true "登录token"
// @Param hostId path string true "待删除的主机ID"
// @Success 200 {object} controller.CommonResult{data=string}
// @Router /resources/hosts/{hostId} [delete]
func RemoveHost(c *gin.Context) {

	hostId := c.Param("hostId")

	RemoveHostReq := cluster.RemoveHostRequest{
		HostId: hostId,
	}

	rsp, err := client2.ClusterClient.RemoveHost(c, &RemoveHostReq)
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

// RemoveHosts 批量删除主机接口
// @Summary 批量删除指定的主机
// @Description 批量删除指定的主机
// @Tags resource
// @Accept json
// @Produce json
// @Param Token header string true "登录token"
// @Param hostIds body []string true "待删除的主机ID数组"
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

	RemoveHostsReq := cluster.RemoveHostsInBatchRequest{
		HostIds: hostIds,
	}

	rsp, err := client2.ClusterClient.RemoveHostsInBatch(c, &RemoveHostsReq)
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

// DownloadHostTemplateFile 导出主机信息模板文件
// @Summary 导出主机信息模板文件
// @Description 将主机信息文件导出到本地
// @Tags resource
// @Accept json
// @Produce octet-stream
// @Param Token header string true "登录token"
// @Success 200 {file} file
// @Router /resources/hosts-template/ [get]
func DownloadHostTemplateFile(c *gin.Context) {
	curDir, _ := os.Getwd()
	templateName := "hostInfo_template.xlsx"
	filePath := filepath.Join(curDir, "../etc/", templateName)

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

func copyAllocToReq(src []Allocation, dst *[]*cluster.AllocationReq) {
	for _, req := range src {
		*dst = append(*dst, &cluster.AllocationReq{
			FailureDomain: req.FailureDomain,
			CpuCores:      req.CpuCores,
			Memory:        req.Memory,
			Count:         req.Count,
		})
	}
}

func copyAllocFromRsp(src []*cluster.AllocHost, dst *[]AllocateRsp) {
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
		(*dst)[i].Disk.DiskId = host.Disk.DiskId
		(*dst)[i].Disk.Name = host.Disk.Name
		(*dst)[i].Disk.Path = host.Disk.Path
		(*dst)[i].Disk.Capacity = host.Disk.Capacity
		(*dst)[i].Disk.Status = host.Disk.Status
	}
}

// AllocHosts 分配主机接口
// @Summary 分配主机接口
// @Description 按指定的配置分配主机资源
// @Tags resource
// @Accept json
// @Produce json
// @Param Token header string true "登录token"
// @Param Alloc body AllocHostsReq true "主机分配请求"
// @Success 200 {object} controller.CommonResult{data=AllocHostsRsp}
// @Router /resources/allochosts [post]
func AllocHosts(c *gin.Context) {
	var allocation AllocHostsReq
	if err := c.ShouldBindJSON(&allocation); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}

	allocReq := cluster.AllocHostsRequest{}
	copyAllocToReq(allocation.PdReq, &allocReq.PdReq)
	copyAllocToReq(allocation.TidbReq, &allocReq.TidbReq)
	copyAllocToReq(allocation.TikvReq, &allocReq.TikvReq)
	//fmt.Println(allocReq.PdReq, allocReq.TidbReq, allocReq.TikvReq)
	rsp, err := client2.ClusterClient.AllocHosts(c, &allocReq)
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

// GetFailureDomain 查询指定故障域里的资源情况
// @Summary 查询指定故障域的资源
// @Description 查询指定故障域的资源情况
// @Tags resource
// @Accept json
// @Produce json
// @Param Token header string true "登录token"
// @Param failureDomainType query int false "指定故障域类型" Enums(1, 2, 3)
// @Success 200 {object} controller.CommonResult{data=[]DomainResource}
// @Router /resources/failuredomains [get]
func GetFailureDomain(c *gin.Context) {
	var domain int
	domainStr := c.Query("failureDomainType")
	if domainStr == "" {
		domain = int(service.ZONE)
	}
	domain, err := strconv.Atoi(domainStr)
	if err != nil || domain > int(service.RACK) || domain < int(service.DATACENTER) {
		errmsg := fmt.Sprintf("Input domainType [%s] Invalid: %v", c.Query("failureDomainType"), err)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}

	GetDoaminReq := cluster.GetFailureDomainRequest{
		FailureDomainType: int32(domain),
	}

	rsp, err := client2.ClusterClient.GetFailureDomain(c, &GetDoaminReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}
	var res DomainResourceRsp
	for _, v := range rsp.FdList {
		res.Resources = append(res.Resources, DomainResource{
			ZoneName: service.GetDomainNameFromCode(v.FailureDomain),
			ZoneCode: v.FailureDomain,
			Purpose:  v.Purpose,
			SpecName: v.Spec,
			SpecCode: v.Spec,
			Count:    v.Count,
		})
	}
	c.JSON(http.StatusOK, controller.Success(res.Resources))
}

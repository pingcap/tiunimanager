package hostapi

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

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-manager/client"
	manager "github.com/pingcap/ticp/micro-manager/proto"
	"github.com/pingcap/ticp/micro-metadb/service"
	"google.golang.org/grpc/codes"
)

// Query 查询主机接口
// @Summary 查询主机接口
// @Description 查询主机
// @Tags resource
// @Accept json
// @Produce json
// @Param Token header string true "登录token"
// @Param query body HostQuery true "查询请求"
// @Success 200 {object} controller.ResultWithPage{data=[]DemoHostInfo}
// @Router /host/query [post]
func Query(c *gin.Context) {
	var req HostQuery
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	listHostReq := manager.ListHostsRequest{
		Purpose: "storage",
		Status:  0,
	}

	rsp, err := client.ManagerClient.ListHost(c, &listHostReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, rsp.Rs.Message))
	} else {
		var res []DemoHostInfo
		for _, v := range rsp.HostList {
			host := DemoHostInfo{
				HostId:   v.Ip,
				HostIp:   v.Ip,
				HostName: v.HostName,
			}
			res = append(res, host)
		}
		c.JSON(http.StatusOK, controller.SuccessWithPage(res, controller.Page{Page: 1, PageSize: 20, Total: len(res)}))
	}
}

func CopyHostFromRsp(src *manager.HostInfo, dst *HostInfo) {
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
	return fmt.Sprintf("%dU%dG", cpuCores, mem)
}

func copyHostToReq(src *HostInfo, dst *manager.HostInfo) {
	dst.HostName = src.HostName
	dst.Ip = src.Ip
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
		dst.Disks = append(dst.Disks, &manager.Disk{
			Name:     v.Name,
			Capacity: v.Capacity,
			Status:   v.Status,
			Path:     v.Path,
		})
	}
}

func doImport(c *gin.Context, host *HostInfo) (rsp *manager.ImportHostResponse, err error) {
	importReq := manager.ImportHostRequest{}
	importReq.Host = new(manager.HostInfo)
	copyHostToReq(host, importReq.Host)
	return client.ManagerClient.ImportHost(c, &importReq)
}

func doImportBatch(c *gin.Context, hosts []*HostInfo) (rsp *manager.ImportHostsInBatchResponse, err error) {
	importReq := manager.ImportHostsInBatchRequest{}
	importReq.Hosts = make([]*manager.HostInfo, len(hosts))
	for i, host := range hosts {
		importReq.Hosts[i] = new(manager.HostInfo)
		copyHostToReq(host, importReq.Hosts[i])
	}

	return client.ManagerClient.ImportHostsInBatch(c, &importReq)
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
// @Router /host [post]
func ImportHost(c *gin.Context) {
	var host HostInfo
	if err := c.ShouldBindJSON(&host); err != nil {
		_ = c.Error(err)
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
				errMsg := fmt.Sprintf("Row %d has a Invalid IP Address %s", irow, row[1])
				return nil, errors.New(errMsg)
			}
			host.Ip = addr.String()
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
// @Router /hosts [post]
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
// @Param purpose query string false "查询特定用途的主机列表"
// @Param status query string false "查询特定状态的主机列表"
// @Success 200 {object} controller.ResultWithPage{data=[]HostInfo}
// @Router /hosts [get]
func ListHost(c *gin.Context) {
	const ALL_STATS = "-1"
	statusStr := c.Query("status")
	if statusStr == "" {
		statusStr = ALL_STATS
	}
	queryStatus, err := strconv.Atoi(statusStr)
	if err != nil {
		errmsg := fmt.Sprintf("Input Status %s Invalid: %v", c.Query("status"), err)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	queryPurpose := c.Query("purpose")

	listHostReq := manager.ListHostsRequest{
		Purpose: queryPurpose,
		Status:  int32(queryStatus),
	}

	rsp, err := client.ManagerClient.ListHost(c, &listHostReq)
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
	c.JSON(http.StatusOK, controller.SuccessWithPage(res, controller.Page{Page: 1, PageSize: 20, Total: len(res.Hosts)}))
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
// @Router /host/{hostId} [get]
func HostDetails(c *gin.Context) {

	hostId := c.Param("hostId")

	HostDetailsReq := manager.CheckDetailsRequest{
		HostId: hostId,
	}

	rsp, err := client.ManagerClient.CheckDetails(c, &HostDetailsReq)
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
// @Router /host/{hostId} [delete]
func RemoveHost(c *gin.Context) {

	hostId := c.Param("hostId")

	RemoveHostReq := manager.RemoveHostRequest{
		HostId: hostId,
	}

	rsp, err := client.ManagerClient.RemoveHost(c, &RemoveHostReq)
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
// @Router /hosts/ [delete]
func RemoveHosts(c *gin.Context) {

	var hostIds []string
	if err := c.ShouldBindJSON(&hostIds); err != nil {
		_ = c.Error(err)
		return
	}

	if str, dup := detectDuplicateElement(hostIds); dup {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), str+" Is Duplicated in request"))
		return
	}

	RemoveHostsReq := manager.RemoveHostsInBatchRequest{
		HostIds: hostIds,
	}

	rsp, err := client.ManagerClient.RemoveHostsInBatch(c, &RemoveHostsReq)
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
// @Router /download_template/ [get]
func DownloadHostTemplateFile(c *gin.Context) {
	curDir, _ := os.Getwd()
	templateName := "hostInfo_template.xlsx"
	filePath := filepath.Join(curDir, templateName)

	fileTmp, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.NotFound), err.Error()))
		return
	}
	defer fileTmp.Close()

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+templateName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	c.File(filePath)
}

func copyAllocToReq(src []Allocation, dst *[]*manager.AllocationReq) {
	for _, req := range src {
		*dst = append(*dst, &manager.AllocationReq{
			FailureDomain: req.FailureDomain,
			CpuCores:      req.CpuCores,
			Memory:        req.Memory,
			Count:         req.Count,
		})
	}
}

func copyAllocFromRsp(src []*manager.AllocHost, dst *[]AllocateRsp) {
	for i, host := range src {
		*dst = append(*dst, AllocateRsp{
			HostName: host.HostName,
			Ip:       host.Ip,
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
// @Router /allochosts [post]
func AllocHosts(c *gin.Context) {
	var allocation AllocHostsReq
	if err := c.ShouldBindJSON(&allocation); err != nil {
		_ = c.Error(err)
		return
	}

	allocReq := manager.AllocHostsRequest{}
	copyAllocToReq(allocation.PdReq, &allocReq.PdReq)
	copyAllocToReq(allocation.TidbReq, &allocReq.TidbReq)
	copyAllocToReq(allocation.TikvReq, &allocReq.TikvReq)
	//fmt.Println(allocReq.PdReq, allocReq.TidbReq, allocReq.TikvReq)
	rsp, err := client.ManagerClient.AllocHosts(c, &allocReq)
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
// @Success 200 {object} controller.ResultWithPage{data=[]DomainResource}
// @Router /failuredomains [get]
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

	GetDoaminReq := manager.GetFailureDomainRequest{
		FailureDomainType: int32(domain),
	}

	rsp, err := client.ManagerClient.GetFailureDomain(c, &GetDoaminReq)
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
			FailureDomain: service.GetDomainNameFromCode(v.FailureDomain),
			DomainCode:    v.FailureDomain,
			Purpose:       v.Purpose,
			Spec:          v.Spec,
			SpecCode:      v.Spec,
			Count:         v.Count,
		})
	}
	c.JSON(http.StatusOK, controller.SuccessWithPage(res, controller.Page{Page: 1, PageSize: 20, Total: len(res.Resources)}))
}

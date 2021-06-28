package hostapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-manager/client"
	manager "github.com/pingcap/ticp/micro-manager/proto"
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
		Status:  manager.HostStatus(0),
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
	dst.HostName = src.HostName
	dst.Ip = src.Ip
	dst.Os = src.Os
	dst.Kernel = src.Kernel
	dst.CpuCores = int32(src.CpuCores)
	dst.Memory = int32(src.Memory)
	dst.Nic = src.Nic
	dst.Az = src.Az
	dst.Rack = src.Rack
	dst.Status = int32(src.Status)
	dst.Purpose = src.Purpose
	for _, disk := range src.Disks {
		dst.Disks = append(dst.Disks, Disk{
			Name:     disk.Name,
			Path:     disk.Path,
			Capacity: disk.Capacity,
			Status:   int32(disk.Status),
		})
	}
}

// ImportHost 导入主机接口
// @Summary 导入主机接口
// @Description 将给定的主机信息导入TiCP
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

	importReq := manager.ImportHostRequest{}
	importReq.Host = &manager.HostInfo{
		HostName: host.HostName,
		Ip:       host.Ip,
		Os:       host.Os,
		Kernel:   host.Kernel,
		CpuCores: host.CpuCores,
		Memory:   host.Memory,
		Az:       host.Az,
		Rack:     host.Rack,
		Nic:      host.Nic,
		Status:   manager.HostStatus(host.Status),
		Purpose:  host.Purpose,
	}
	for _, v := range host.Disks {
		importReq.Host.Disks = append(importReq.Host.Disks, &manager.Disk{
			Name:     v.Name,
			Capacity: v.Capacity,
			Status:   manager.DiskStatus(v.Status),
			Path:     v.Path,
		})
	}

	rsp, err := client.ManagerClient.ImportHost(c, &importReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, rsp.Rs.Message))
	} else {
		c.JSON(http.StatusOK, controller.Success(ImportHostRsp{HostId: rsp.HostId}))
	}
}

// ListHost 查询主机列表接口
// @Summary 查询主机列表
// @Description 展示目前所有主机
// @Tags resource
// @Accept json
// @Produce json
// @Param Token header string true "登录token"
// @Param query body ListHostCondition false "可选的查询主机的条件"
// @Success 200 {object} controller.ResultWithPage{data=[]HostInfo}
// @Router /hosts [get]
func ListHost(c *gin.Context) {
	var cond ListHostCondition
	if err := c.ShouldBindJSON(&cond); err != nil {
		_ = c.Error(err)
		return
	}

	listHostReq := manager.ListHostsRequest{
		Purpose: cond.Purpose,
		Status:  manager.HostStatus(cond.Status),
	}

	rsp, err := client.ManagerClient.ListHost(c, &listHostReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, rsp.Rs.Message))
	} else {
		var res ListHostRsp
		for _, v := range rsp.HostList {
			var host HostInfo
			CopyHostFromRsp(v, &host)
			res.Hosts = append(res.Hosts, host)
		}
		c.JSON(http.StatusOK, controller.SuccessWithPage(res, controller.Page{Page: 1, PageSize: 20, Total: len(res.Hosts)}))
	}
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
// @Router /host/ [get]
func HostDetails(c *gin.Context) {

	hostId := c.Param("hostId")

	HostDetailsReq := manager.CheckDetailsRequest{
		HostId: hostId,
	}

	rsp, err := client.ManagerClient.CheckDetails(c, &HostDetailsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, rsp.Rs.Message))
	} else {
		var res HostDetailsRsp
		CopyHostFromRsp(rsp.Details, &(res.Host))
		c.JSON(http.StatusOK, controller.Success(res))
	}
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
// @Router /host/ [delete]
func RemoveHost(c *gin.Context) {

	hostId := c.Param("hostId")

	RemoveHostReq := manager.RemoveHostRequest{
		HostId: hostId,
	}

	rsp, err := client.ManagerClient.RemoveHost(c, &RemoveHostReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, rsp.Rs.Message))
	} else {
		c.JSON(http.StatusOK, controller.Success(rsp.Rs.Message))
	}
}

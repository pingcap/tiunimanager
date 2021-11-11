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

package warehouse

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/common/resource-type"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/micro-api/controller"

	"google.golang.org/grpc/codes"
)

// GetFailureDomain godoc
// @Summary Show the resources on failure domain view
// @Description get resource info in each failure domain
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param failureDomainType query int false "failure domain type of dc/zone/rack" Enums(1, 2, 3)
// @Success 200 {object} controller.CommonResult{data=[]DomainResource}
// @Router /resources/failuredomains [get]
func GetFailureDomain(c *gin.Context) {
	var domain int
	domainStr := c.Query("failureDomainType")
	if domainStr == "" {
		domain = int(resource.ZONE)
	} else {
		domainInt, err := strconv.Atoi(domainStr) // #nosec G109
		if err != nil || domainInt > int(resource.RACK) || domainInt < int(resource.REGION) {
			errmsg := fmt.Sprintf("Input domainType [%s] Invalid: %v", domainStr, err)
			c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
			return
		}
		domain = domainInt
	}

	GetDoaminReq := clusterpb.GetFailureDomainRequest{
		FailureDomainType: int32(domain),
	}

	rsp, err := client.ClusterClient.GetFailureDomain(framework.NewMicroCtxFromGinCtx(c), &GetDoaminReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}
	res := DomainResourceRsp{
		Resources: make([]DomainResource, 0, len(rsp.FdList)),
	}
	for _, v := range rsp.FdList {
		res.Resources = append(res.Resources, DomainResource{
			ZoneName: resource.GetDomainNameFromCode(v.FailureDomain),
			ZoneCode: v.FailureDomain,
			Purpose:  v.Purpose,
			SpecName: v.Spec,
			SpecCode: v.Spec,
			Count:    v.Count,
		})
	}
	c.JSON(http.StatusOK, controller.Success(res.Resources))
}

func copyHierarchyFromRsp(root *clusterpb.Node, dst *Node) {
	dst.Code = root.Code
	dst.Prefix = root.Prefix
	dst.Name = root.Name
	if root.SubNodes == nil {
		return
	}
	dst.SubNodes = make([]Node, len(root.SubNodes))
	for i, node := range root.SubNodes {
		copyHierarchyFromRsp(node, &dst.SubNodes[i])
	}
}

// GetHierarchy godoc
// @Summary Show the resources hierarchy
// @Description get resource hierarchy-tree
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param filter query HostFilter true "resource filter"
// @Param level query int true "failure domain type of region/zone/rack/host" Enums(1, 2, 3, 4)
// @Param depth query int true "hierarchy depth"
// @Success 200 {object} controller.CommonResult{data=Node}
// @Router /resources/hierarchy [get]
func GetHierarchy(c *gin.Context) {
	var filter HostFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	if err := resource.ValidArch(filter.Arch); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	var level int
	levelStr := c.DefaultQuery("level", "1")
	levelInt, err := strconv.Atoi(levelStr) // #nosec G109
	if err != nil || levelInt > int(resource.HOST) || levelInt < int(resource.REGION) {
		errmsg := fmt.Sprintf("Input domainType [%s] invalid: %v", levelStr, err)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	level = levelInt

	var depth int
	depthStr := c.DefaultQuery("depth", "0")
	depthInt, err := strconv.Atoi(depthStr)
	if err != nil || depthInt < 0 || levelInt+depthInt > int(resource.HOST) {
		errmsg := fmt.Sprintf("Input depth [%s] invalid or is not vaild(level+depth>4) where level is [%d]: %v", depthStr, level, err)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	depth = depthInt

	GetHierarchyReq := clusterpb.GetHierarchyRequest{
		Level: int32(level),
		Depth: int32(depth),
		Filter: &clusterpb.HostFilter{
			Arch: filter.Arch,
		},
	}
	rsp, err := client.ClusterClient.GetHierarchy(framework.NewMicroCtxFromGinCtx(c), &GetHierarchyReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}

	var res GetHierarchyRsp
	copyHierarchyFromRsp(rsp.Root, &res.Root)
	c.JSON(http.StatusOK, controller.Success(res.Root))
}

func copyStocksFromRsp(src *clusterpb.Stocks, dst *Stocks) {
	dst.FreeCpuCores = src.FreeCpuCores
	dst.FreeMemory = src.FreeMemory
	dst.FreeHostCount = src.FreeHostCount
	dst.FreeDiskCount = src.FreeDiskCount
	dst.FreeDiskCapacity = src.FreeDiskCapacity
}

// GetStocks godoc
// @Summary Show the resources stocks
// @Description get resource stocks in specified conditions
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param cond query StockCondition true "query condition"
// @Success 200 {object} controller.CommonResult{data=Stocks}
// @Router /resources/stocks [get]
func GetStocks(c *gin.Context) {
	cond := StockCondition{
		StockHostCondition: StockHostCondition{
			HostStatus: int32(resource.HOST_WHATEVER),
			LoadStat:   int32(resource.HOST_STAT_WHATEVER),
		},
		StockDiskCondition: StockDiskCondition{
			DiskStatus: int32(resource.DISK_STATUS_WHATEVER),
		},
	}
	if err := c.ShouldBindQuery(&cond); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	var req clusterpb.GetStocksRequest
	req.Location = new(clusterpb.StockLocation)
	req.Location.Region = cond.Region
	req.Location.Zone = cond.Zone
	req.Location.Rack = cond.Rack
	req.Location.Host = cond.HostIp
	req.HostFilter = new(clusterpb.StockHostFilter)
	req.DiskFilter = new(clusterpb.StockDiskFilter)
	if cond.Arch != "" {
		if err := resource.ValidArch(cond.Arch); err != nil {
			c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
			return
		}
	}
	req.HostFilter.Arch = cond.Arch

	if !resource.HostStatus(cond.HostStatus).IsValidForQuery() {
		errmsg := fmt.Sprintf("input host status %d is invalid for query", cond.HostStatus)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	req.HostFilter.Status = cond.HostStatus

	if !resource.HostStat(cond.LoadStat).IsValidForQuery() {
		errmsg := fmt.Sprintf("input load stat %d is invalid for query", cond.LoadStat)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	req.HostFilter.Stat = cond.LoadStat

	if !resource.DiskStatus(cond.DiskStatus).IsValidForQuery() {
		errmsg := fmt.Sprintf("input disk status %d is invalid for query", cond.DiskStatus)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	req.DiskFilter.Status = cond.DiskStatus

	if cond.DiskType != "" {
		if err := resource.ValidDiskType(cond.DiskType); err != nil {
			c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
			return
		}
	}
	req.DiskFilter.Type = cond.DiskType

	if cond.Capacity < 0 {
		errmsg := fmt.Sprintf("input disk capacity %d is invalid for query", cond.Capacity)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	req.DiskFilter.Capacity = cond.Capacity

	rsp, err := client.ClusterClient.GetStocks(framework.NewMicroCtxFromGinCtx(c), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if rsp.Rs.Code != int32(codes.OK) {
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}

	var res GetStocksRsp
	copyStocksFromRsp(rsp.Stocks, &res.Stocks)
	c.JSON(http.StatusOK, controller.Success(res.Stocks))
}

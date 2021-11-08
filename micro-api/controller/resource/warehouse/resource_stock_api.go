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
	"time"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/common/resource-type"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
	"github.com/pingcap-inc/tiem/micro-metadb/service"

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
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "GetFailureDomain", int(status.GetCode()))

	var domain int
	domainStr := c.Query("failureDomainType")
	if domainStr == "" {
		domain = int(resource.ZONE)
	} else {
		domainInt, err := strconv.Atoi(domainStr) // #nosec G109
		if err != nil || domainInt > int(resource.RACK) || domainInt < int(resource.REGION) {
			errmsg := fmt.Sprintf("Input domainType [%s] Invalid: %v", domainStr, err)
			status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
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
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}

	status = rsp.GetStatus()
	if rsp.Rs.Code != int32(codes.OK) {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}
	res := DomainResourceRsp{
		Resources: make([]DomainResource, 0, len(rsp.FdList)),
	}
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
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "GetHierarchy", int(status.GetCode()))

	var filter HostFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	if err := resource.ValidArch(filter.Arch); err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	var level int
	levelStr := c.DefaultQuery("level", "1")
	levelInt, err := strconv.Atoi(levelStr) // #nosec G109
	if err != nil || levelInt > int(resource.HOST) || levelInt < int(resource.REGION) {
		errmsg := fmt.Sprintf("Input domainType [%s] invalid: %v", levelStr, err)
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), errmsg))
		return
	}
	level = levelInt

	var depth int
	depthStr := c.DefaultQuery("depth", "0")
	depthInt, err := strconv.Atoi(depthStr)
	if err != nil || depthInt < 0 || levelInt+depthInt > int(resource.HOST) {
		errmsg := fmt.Sprintf("Input depth [%s] invalid or is not vaild(level+depth>4) where level is [%d]: %v", depthStr, level, err)
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
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
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}

	status = rsp.GetStatus()
	if rsp.Rs.Code != int32(codes.OK) {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(int(rsp.Rs.Code), rsp.Rs.Message))
		return
	}

	var res GetHierarchyRsp
	copyHierarchyFromRsp(rsp.Root, &res.Root)
	c.JSON(http.StatusOK, controller.Success(res.Root))
}

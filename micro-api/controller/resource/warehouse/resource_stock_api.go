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
 *                                                                            *
 ******************************************************************************/

package warehouse

import (
	"fmt"
	"github.com/pingcap/tiunimanager/common/client"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/micro-api/controller"
)

func setGinContextForInvalidParam(c *gin.Context, errmsg string) {
	framework.LogWithContext(c).Error(errmsg)
	c.JSON(errors.TIUNIMANAGER_PARAMETER_INVALID.GetHttpCode(), controller.Fail(int(errors.TIUNIMANAGER_PARAMETER_INVALID), errmsg))
}

// GetHierarchy godoc
// @Summary Show the resources hierarchy
// @Description get resource hierarchy-tree
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param filter query message.GetHierarchyReq true "resource filter"
// @Success 200 {object} controller.CommonResult{data=message.GetHierarchyResp}
// @Router /resources/hierarchy [get]
func GetHierarchy(c *gin.Context) {
	var req message.GetHierarchyReq

	requestBody, ok := controller.HandleJsonRequestFromQuery(c, &req)
	if ok {
		if err := constants.ValidArchType(req.Arch); err != nil {
			setGinContextForInvalidParam(c, err.Error())
			return
		}

		if req.Level > int(constants.HOST) || req.Level < int(constants.REGION) {
			errmsg := fmt.Sprintf("Input domainType [%d] invalid, [1:Region, 2:Zone, 3:Rack, 4:Host]", req.Level)
			setGinContextForInvalidParam(c, errmsg)
			return
		}

		if req.Depth < 0 || req.Depth+req.Level > int(constants.HOST) {
			errmsg := fmt.Sprintf("Input depth [%d] invalid or is not vaild(level+depth>4) where level is [%d]", req.Depth, req.Level)
			setGinContextForInvalidParam(c, errmsg)
			return
		}

		controller.InvokeRpcMethod(c, client.ClusterClient.GetHierarchy, &message.GetHierarchyResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// GetStocks godoc
// @Summary Show the resources stocks
// @Description get resource stocks in specified conditions
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param stockFilter query message.GetStocksReq true "query condition"
// @Success 200 {object} controller.CommonResult{data=message.GetStocksResp}
// @Router /resources/stocks [get]
func GetStocks(c *gin.Context) {
	var req message.GetStocksReq
	requestBody, ok := controller.HandleJsonRequestFromQuery(c, &req)
	if ok {
		if req.Arch != "" {
			if err := constants.ValidArchType(req.Arch); err != nil {
				setGinContextForInvalidParam(c, err.Error())
				return
			}
		}

		if req.HostFilter.Status != "" {
			if !constants.HostStatus(req.HostFilter.Status).IsValidStatus() {
				errmsg := fmt.Sprintf("input host status %s is invalid for query", req.HostFilter.Status)
				setGinContextForInvalidParam(c, errmsg)
				return
			}
		}
		if req.HostFilter.Stat != "" {
			if !constants.HostLoadStatus(req.HostFilter.Stat).IsValidLoadStatus() {
				errmsg := fmt.Sprintf("input load stat %s is invalid for query", req.HostFilter.Stat)
				setGinContextForInvalidParam(c, errmsg)
				return
			}
		}

		if req.DiskFilter.DiskStatus != "" {
			if !constants.DiskStatus(req.DiskFilter.DiskStatus).IsValidStatus() {
				errmsg := fmt.Sprintf("input disk status %s is invalid for query", req.DiskFilter.DiskStatus)
				setGinContextForInvalidParam(c, errmsg)
				return
			}
		}

		if req.DiskType != "" {
			if err := constants.ValidDiskType(req.DiskType); err != nil {
				setGinContextForInvalidParam(c, err.Error())
				return
			}
		}

		if req.Capacity < 0 {
			errmsg := fmt.Sprintf("input disk capacity %d is invalid for query", req.Capacity)
			setGinContextForInvalidParam(c, errmsg)
			return
		}

		controller.InvokeRpcMethod(c, client.ClusterClient.GetStocks, &message.GetStocksResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

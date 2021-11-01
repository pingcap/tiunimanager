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

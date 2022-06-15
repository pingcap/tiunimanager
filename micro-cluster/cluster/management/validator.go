/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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

package management

import (
	"context"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message/cluster"
	"github.com/pingcap/tiunimanager/models"
)

var validator = validateCreating

func validateCreating(ctx context.Context, req *cluster.CreateClusterReq) error {
	product, versions, components, err := models.GetProductReaderWriter().GetProduct(ctx, req.Type)
	if err != nil {
		return errors.WrapError(errors.TIUNIMANAGER_UNSUPPORT_PRODUCT, "get product failed", err)
	}

	return errors.OfNullable(nil).BreakIf(func() error {
		if len(product.ProductID) == 0 || len(versions) == 0 || len(components) == 0 {
			return errors.NewErrorf(errors.TIUNIMANAGER_UNSUPPORT_PRODUCT, "product %s is not available", req.Type)
		}
		return nil
	}).BreakIf(func() error {
		for _, v := range versions {
			if v.ProductID == req.Type && v.Version == req.Version && v.Arch == req.CpuArchitecture {
				return nil
			}
		}
		return errors.NewErrorf(errors.TIUNIMANAGER_UNSUPPORT_PRODUCT, "product is not available, productID = %s, version = %s, arch = %s", req.Type, req.Version, req.CpuArchitecture)
	}).BreakIf(func() error {
		count := req.ResourceParameter.GetComponentCount(constants.ComponentIDTiKV)
		if count < int32(req.Copies) {
			return errors.NewErrorf(errors.TIUNIMANAGER_INVALID_TOPOLOGY, "count of TiKV = %d, is less than copies %d", count, req.Copies)
		}
		return nil
	}).BreakIf(func() error {
		for _, componentInfo := range components {
			count := req.ResourceParameter.GetComponentCount(constants.EMProductComponentIDType(componentInfo.ComponentID))
			if count > componentInfo.MaxInstance {
				return errors.NewErrorf(errors.TIUNIMANAGER_INVALID_TOPOLOGY, "count of components %s should be less than %d", componentInfo.ComponentID, componentInfo.MaxInstance)
			}
			if count < componentInfo.MinInstance {
				return errors.NewErrorf(errors.TIUNIMANAGER_INVALID_TOPOLOGY, "count of components %s should be more than %d", componentInfo.ComponentID, componentInfo.MinInstance)
			}
			if len(componentInfo.SuggestedInstancesCount) > 0 {
				matched := false
				for _, c := range componentInfo.SuggestedInstancesCount {
					if c == count {
						matched = true
					}
				}
				if !matched {
					return errors.NewErrorf(errors.TIUNIMANAGER_INVALID_TOPOLOGY, "the total number of %s should be in %v", componentInfo.ComponentID, componentInfo.SuggestedInstancesCount)
				}
			}
		}
		return nil
	}).If(func(err error) {
		framework.LogWithContext(ctx).Error(err.Error())
	}).Else(func() {
		framework.LogWithContext(ctx).Infof("validate creating request succeed")
	}).Present()
}

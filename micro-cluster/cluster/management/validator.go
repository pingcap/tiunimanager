/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
)

var creatingRules = []func(req *cluster.CreateClusterReq, product *structs.ProductDetail) error {
	func(req *cluster.CreateClusterReq, product *structs.ProductDetail) error {
		if _, ok := product.Versions[req.Version]; !ok {
			return errors.NewErrorf(errors.TIEM_UNSUPPORT_PRODUCT, "version %s is not supported", req.Version)
		}
		return nil
	},
	func(req *cluster.CreateClusterReq, product *structs.ProductDetail) error {
		if _, ok := product.Versions[req.Version].Arch[req.CpuArchitecture]; !ok {
			return errors.NewErrorf(errors.TIEM_UNSUPPORT_PRODUCT, "arch %s is not supported", req.CpuArchitecture)
		}
		return nil
	},
	func(req *cluster.CreateClusterReq, product *structs.ProductDetail) error {
		// TiKV
		count := req.ResourceParameter.GetComponentCount(constants.ComponentIDTiKV)
		if count < int32(req.Copies) {
			return errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "count of TiKV = %d, is less than copies %d", count, req.Copies)
		}
		return nil
	},
	func(req *cluster.CreateClusterReq, product *structs.ProductDetail) error {
		for _, property := range product.Versions[req.Version].Arch[req.CpuArchitecture] {
			count := req.ResourceParameter.GetComponentCount(constants.EMProductComponentIDType(property.ID))
			if count > property.MaxInstance {
				return errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "count of components %s should be less than %d", property.ID, property.MaxInstance)
			}
			if count < property.MinInstance {
				return errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "count of components %s should be more than %d", property.ID, property.MinInstance)
			}
			if len(property.SuggestedInstancesCount) > 0 {
				matched := false
				for _, c := range property.SuggestedInstancesCount {
					if c == count {
						matched = true
					}
				}
				if !matched {
					return errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "the total number of %s should be in %v", property.ID, property.SuggestedInstancesCount)
				}
			}
		}
		return nil
	},
}

var validator = validateCreating
func validateCreating(ctx context.Context, req *cluster.CreateClusterReq) error {
	optionalError := errors.OfNullable(nil)
	products, err := models.GetProductReaderWriter().QueryProductDetail(ctx, req.Vendor, req.Region, req.Type, constants.ProductStatusOnline, constants.EMInternalProductNo)
	if err != nil {
		errMsg := fmt.Sprintf("get product detail failed, vendor = %s, region = %s, productID = %s", req.Vendor, req.Region, req.Type)
		framework.LogWithContext(ctx).Errorf("%s, err = %s", errMsg, err.Error())
		return err
	}

	if product, ok := products[req.Type]; !ok {
		errMsg := fmt.Sprintf("product is not existed, vendor = %s, region = %s, productID = %s", req.Vendor, req.Region, req.Type)
		framework.LogWithContext(ctx).Error(errMsg)
		return errors.NewErrorf(errors.TIEM_UNSUPPORT_PRODUCT, errMsg)
	} else {
		for _, v := range creatingRules {
			optionalError.BreakIf(func() error {
				return v(req, &product)
			})
		}
	}

	return optionalError.If(func(err error) {
		framework.LogWithContext(ctx).Error(err.Error())
	}).Else(func() {
		framework.LogWithContext(ctx).Infof("validate creating request succeed")
	}).Present()
}

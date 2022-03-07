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

/*******************************************************************************
 * @File: manager.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/3/7
*******************************************************************************/

package product

import (
	"context"
	"github.com/pingcap-inc/tiem/message"
	"sync"
)

type Manager struct{}
var once sync.Once

var manager *Manager

func NewManager() *Manager {
	once.Do(func() {
		if manager == nil {
			manager = &Manager{}
		}
	})
	return manager
}

func (p *Manager) UpdateVendors(ctx context.Context, req message.UpdateVendorInfoReq) (message.UpdateVendorInfoResp, error){
	return message.UpdateVendorInfoResp{}, nil
}

func (p *Manager) QueryVendors(ctx context.Context, req message.QueryVendorInfoReq) (message.QueryVendorInfoResp, error){
	return message.QueryVendorInfoResp{}, nil
}

func (p *Manager) QueryAvailableVendors(ctx context.Context, req message.QueryAvailableVendorsReq) (message.QueryAvailableVendorsResp, error){
	return message.QueryAvailableVendorsResp{}, nil
}

func (p *Manager) UpdateProducts(ctx context.Context, req message.UpdateProductsInfoReq) (message.UpdateProductsInfoResp, error){
	return message.UpdateProductsInfoResp{}, nil
}

func (p *Manager) QueryProducts(ctx context.Context, req message.QueryProductsInfoReq) (message.QueryProductsInfoResp, error){
	return message.QueryProductsInfoResp{}, nil
}

func (p *Manager) QueryAvailableProducts(ctx context.Context, req message.QueryAvailableProductsReq) (message.QueryAvailableProductsResp, error){
	return message.QueryAvailableProductsResp{}, nil
}

func (p *Manager) QueryProductDetail(ctx context.Context, req message.QueryProductDetailReq) (message.QueryProductDetailResp, error){
	return message.QueryProductDetailResp{}, nil
}

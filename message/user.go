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
 ******************************************************************************/

/*******************************************************************************
 * @File: user.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package message

import (
	"github.com/pingcap-inc/tiem/common/structs"
)

type CreateAccountReq struct {
	*structs.Tenant
	Name string `json:"name" form:"name" example:"default"`
	Password string `json:"password" form:"password" example:"default"`
}

type CreateAccountResp struct {
	structs.Account
}

type FindAccountByNameReq struct {
	Name string `json:"name" form:"name" example:"default"`
}

type FindAccountByNameResp struct {
	structs.Account
}

type CreateTenantReq struct {
	Name string `json:"name" form:"name" example:"default"`
}

type CreateTenantResp struct {
	structs.Tenant
}

type FindTenantByNameReq struct {
	Name string `json:"name" form:"name" example:"default"`
}

type FindTenantByNameResp struct {
	structs.Tenant
}

type FindTenantByIdReq struct {
	ID string //Todo: gorm
}

type FindTenantByIdResp struct {
	structs.Tenant
}

type ProvideTokenReq struct {
	*structs.Token
}

type ProvideTokenResp struct {
	TokenString string
}

type ModifyTokenReq struct {
	*structs.Token
}

type ModifyTokenResp struct {

}

type GetTokenReq struct {
	TokenString string
}

type GetTokenResp struct {
	structs.Token
}

type LoginReq struct {
	UserName string
	Password string
}

type LoginResp struct {
	TokenString string
}

type LogoutReq struct {
	TokenString string
}

type LogoutResp struct {
	AccountName string
}

type AccessibleReq struct {
	PathType string
	Path string
	TokenString string
}

type AccessibleResp struct {
	TenantID string
	AccountID string
	AccountName string
}

type CreateTokenReq struct {
	AccountID string
	AccountName string
	TenantID string
}

type CreateTokenResp struct {
	structs.Token
}
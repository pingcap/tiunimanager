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
	"github.com/pingcap-inc/tiem/models/user/account"
	"github.com/pingcap-inc/tiem/models/user/identification"
	"github.com/pingcap-inc/tiem/models/user/tenant"
)

type CreateAccountReq struct {
	*tenant.Tenant
	Name string `json:"name" form:"name" example:"default"`
	Password string `json:"password" form:"password" example:"default"`
}

type CreateAccountResp struct {
	account.Account
}

type FindAccountByNameReq struct {
	Name string `json:"name" form:"name" example:"default"`
}

type FindAccountByNameResp struct {
	account.Account
}

type CreateTenantReq struct {
	Name string `json:"name" form:"name" example:"default"`
}

type CreateTenantResp struct {
	tenant.Tenant
}

type FindTenantByNameReq struct {
	Name string `json:"name" form:"name" example:"default"`
}

type FindTenantByNameResp struct {
	tenant.Tenant
}

type FindTenantByIdReq struct {
	ID string //Todo: gorm
}

type FindTenantByIdResp struct {
	tenant.Tenant
}

type ProvideTokenReq struct {
	*identification.Token
}

type ProvideTokenResp struct {
	TokenString string
}

type ModifyTokenReq struct {
	*identification.Token
}

type ModifyTokenResp struct {

}

type GetTokenReq struct {
	TokenString string
}

type GetTokenResp struct {
	identification.Token
}

type LoginReq struct {
	UserName string  `json:"userName" form:"userName"`
	Password string  `json:"userPassword" form:"userPassword"`
}

type LoginResp struct {
	TokenString string `json:"token" form:"token"`
	UserName    string `json:"userName" form:"userName"`
	TenantId    string `json:"tenantId" form:"tenantId"`
}

type LogoutReq struct {
	TokenString string   `json:"token" form:"token"`
}

type LogoutResp struct {
	AccountName string  `json:"accountName" form:"accountName"`
}

type AccessibleReq struct {
	PathType string 	`json:"pathType" form:"pathType"`
	Path string			`json:"path" form:"path"`
	TokenString string 	`json:"tokenString" form:"tokenString"`
}

type AccessibleResp struct {
	TenantID string 		`json:"tenantID" form:"tenantID"`
	AccountID string		`json:"accountID" form:"accountID"`
	AccountName string		`json:"accountName" form:"accountName"`
}

type CreateTokenReq struct {
	AccountID string
	AccountName string
	TenantID string
}

type CreateTokenResp struct {
	identification.Token
}
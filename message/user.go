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
 ******************************************************************************/

/*******************************************************************************
 * @File: user_api.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package message

import (
	"github.com/pingcap/tiunimanager/common/structs"
)

// LoginReq login
type LoginReq struct {
	Name     string                `json:"userName" form:"userName" validate:"required,min=5,max=32"`
	Password structs.SensitiveText `json:"userPassword" form:"userPassword" validate:"required,min=5,max=32"`
}

type LoginResp struct {
	TokenString     structs.SensitiveText `json:"token" form:"token"`
	UserID          string `json:"userId" form:"userId"`
	TenantID        string `json:"tenantId" form:"tenantId"`
	PasswordExpired bool   `json:"passwordExpired" form:"passwordExpired"`
}

// LogoutReq logout
type LogoutReq struct {
	TokenString structs.SensitiveText `json:"token" form:"token" validate:"required,min=8,max=64"`
}

type LogoutResp struct {
	UserID string `json:"userId" form:"userId"`
}

// AccessibleReq identify
type AccessibleReq struct {
	TokenString structs.SensitiveText `json:"token" form:"token" validate:"required,min=8,max=64"`
	CheckPassword bool `json:"checkPassword" form:"checkPassword"`
}

type AccessibleResp struct {
	UserID   string `json:"userId" form:"userId"`
	TenantID string `json:"tenantId" form:"tenantId"`
}

//CreateUserReq user message
type CreateUserReq struct {
	Name     string                `json:"name" validate:"required,min=5,max=32"`
	TenantID string                `json:"tenantId"`
	Email    string                `json:"email" validate:"required,email"`
	Phone    string                `json:"phone"`
	Password structs.SensitiveText `json:"password" validate:"required,min=5,max=32"`
	Nickname string                `json:"nickname"`
}

type CreateUserResp struct {
}

type DeleteUserReq struct {
	ID string `json:"id" swaggerignore:"true"`
}
type DeleteUserResp struct {
}

type GetUserReq struct {
	ID string `json:"id" swaggerignore:"true"`
}
type GetUserResp struct {
	User structs.UserInfo `json:"user"`
}

type QueryUserReq struct {
	structs.PageRequest
}
type QueryUserResp struct {
	Users map[string]structs.UserInfo `json:"users"`
}

type UpdateUserProfileReq struct {
	ID       string `json:"id" swaggerignore:"true"`
	Nickname string `json:"nickname"`
	Email    string `json:"email" validate:"email"`
	Phone    string `json:"phone"`
}
type UpdateUserProfileResp struct {
}

type UpdateUserPasswordReq struct {
	ID       string                `json:"id"`
	Password structs.SensitiveText `json:"password" validate:"required,min=5,max=32"`
}
type UpdateUserPasswordResp struct {
}

// CreateTenantReq Tenant message
type CreateTenantReq struct {
	ID               string `json:"id" validate:"required,min=5,max=32"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	OnBoardingStatus string `json:"onBoardingStatus"`
	MaxCluster       int32  `json:"maxCluster"`
	MaxCPU           int32  `json:"maxCpu"`
	MaxMemory        int32  `json:"maxMemory"`
	MaxStorage       int32  `json:"maxStorage"`
}
type CreateTenantResp struct {
}

type DeleteTenantReq struct {
	ID string `json:"id"`
}
type DeleteTenantResp struct {
}

type GetTenantReq struct {
	ID string `json:"id"`
}

type GetTenantResp struct {
	Info structs.TenantInfo `json:"info"`
}

type QueryTenantReq struct {
	structs.PageRequest
}

type QueryTenantResp struct {
	Tenants map[string]structs.TenantInfo `json:"tenants"`
}

type UpdateTenantProfileReq struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	MaxCluster int32  `json:"maxCluster"`
	MaxCPU     int32  `json:"maxCpu"`
	MaxMemory  int32  `json:"maxMemory"`
	MaxStorage int32  `json:"maxStorage"`
}
type UpdateTenantProfileResp struct {
}

type UpdateTenantOnBoardingStatusReq struct {
	ID               string `json:"id"`
	OnBoardingStatus string `json:"onBoardingStatus"`
}
type UpdateTenantOnBoardingStatusResp struct {
}

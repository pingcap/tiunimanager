/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package structs

import "time"

type UserInfo struct {
	ID              string    `json:"id"`
	DefaultTenantID string    `json:"defaultTenantId"`
	Name            []string  `json:"names"`
	Creator         string    `json:"creator"`
	TenantID        []string  `json:"tenantIds"`
	Nickname        string    `json:"nickname"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Status          string    `json:"status"`
	CreateAt        time.Time `json:"createAt"`
	UpdateAt        time.Time `json:"updateAt"`
}

type TenantInfo struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Creator          string    `json:"creator"`
	Status           string    `json:"status"`
	OnBoardingStatus string    `json:"onBoardingStatus"`
	MaxCluster       int32     `json:"maxCluster"`
	MaxCPU           int32     `json:"maxCpu"`
	MaxMemory        int32     `json:"maxMemory"`
	MaxStorage       int32     `json:"maxStorage"`
	CreateAt         time.Time `json:"createAt"`
	UpdateAt         time.Time `json:"updateAt"`
}

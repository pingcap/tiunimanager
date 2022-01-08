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
 * @File: user_api.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package constants

type TenantStatus string

//Definition tenant status information
const (
	TenantStatusNormal     TenantStatus = "Normal"
	TenantStatusDeactivate TenantStatus = "Deactivate"
)

type TenantOnBoardingStatus string
const (
	TenantOnBoarding  TenantOnBoardingStatus = "On"
	TenantOFFBoarding TenantOnBoardingStatus = "Off"
)

type UserStatus string

//Definition user status information
const (
	UserStatusNormal     UserStatus = "Normal"
	UserStatusDeactivate UserStatus = "Deactivate"
)

type TokenStatus string

//Definition token status information
const (
	TokenStatusNormal     TokenStatus = "Normal"
	TokenStatusDeactivate TokenStatus = "Deactivate"
)

type CommonStatus int

const (
	Valid              CommonStatus = 0
	Invalid            CommonStatus = 1
	Deleted            CommonStatus = 2
	UnrecognizedStatus CommonStatus = -1
)

func (s CommonStatus) IsValid() bool {
	return s == Valid
}

type TenantType int

const (
	SystemManagement  TenantType = 0
	InstanceWorkspace TenantType = 1
	PluginAccess      TenantType = 2
)
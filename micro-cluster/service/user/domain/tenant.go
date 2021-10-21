
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

package domain

// Tenant 租户
type Tenant struct {
	Name   string
	Id     string
	Type   TenantType
	Status CommonStatus
}

type TenantType int

const (
	SystemManagement  TenantType = 0
	InstanceWorkspace TenantType = 1
	PluginAccess      TenantType = 2
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

func CommonStatusFromStatus(status int32) CommonStatus {
	switch status {
	case 0:
		return Valid
	case 1:
		return Invalid
	case 2:
		return Deleted
	default:
		return UnrecognizedStatus
	}
}

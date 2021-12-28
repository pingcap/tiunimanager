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
 * @File: common.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/15 18:08
*******************************************************************************/

package parameter

import "github.com/pingcap-inc/tiem/common/structs"

const (
	contextClusterMeta             = "ClusterMeta"
	contextModifyParameters        = "ModifyParameters"
	contextApplyParameterInfo      = "ApplyParameterInfo"
	contextUpdateParameterInfo     = "UpdateParameterInfo"
	contextMaintenanceStatusChange = "maintenanceStatusChange"
)

type UpdateParameterSource int

const (
	TiUP UpdateParameterSource = iota
	SQL
	TiupAndSql
	API
)

type ParameterValueType int

const (
	Integer ParameterValueType = iota
	String
	Boolean
	Float
	Array
)

type ApplyParameterType int

const (
	ModifyApply ApplyParameterType = iota
	DirectApply
)

type ModifyParameter struct {
	Reboot bool
	Params []structs.ClusterParameterSampleInfo
}

type ClusterReboot int
const (
	NonReboot ClusterReboot = iota
	Reboot
)

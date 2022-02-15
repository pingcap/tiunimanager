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

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pingcap-inc/tiem/common/structs"
)

const (
	contextClusterMeta             = "ClusterMeta"
	contextModifyParameters        = "ModifyParameters"
	contextHasApplyParameter       = "HasApplyParameter"
	contextMaintenanceStatusChange = "MaintenanceStatusChange"
	contextClusterConfigStr        = "ClusterConfigStr"
)

var units = []string{"KB", "MB", "GB", "s", "m", "h"}

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

type ReadOnlyParameter int

const (
	ReadWriter ReadOnlyParameter = iota
	ReadOnly
)

type ModifyParameter struct {
	ClusterID    string
	ParamGroupId string
	Reboot       bool
	Params       []*ModifyClusterParameterInfo
	Nodes        []string
}

type ModifyClusterParameterInfo struct {
	ParamId        string
	Category       string
	Name           string
	InstanceType   string
	UpdateSource   int
	ReadOnly       int
	SystemVariable string
	Type           int
	Unit           string
	UnitOptions    []string
	Range          []string
	HasApply       int
	RealValue      structs.ParameterRealValue
}

type ClusterReboot int

const (
	NonReboot ClusterReboot = iota
	Reboot
)

// ValidateRange
// @Description: validate parameter value by range field
// @Parameter param
// @return bool
func ValidateRange(param *ModifyClusterParameterInfo, hasModify bool) bool {
	// Determine if range is nil or an expression, continue the loop directly
	if param.Range == nil || len(param.Range) == 0 {
		return true
	}
	// If it is a modified parameter workflow, the value empty is skipped
	if hasModify && strings.TrimSpace(param.RealValue.ClusterValue) == "" {
		return true
	}
	switch param.Type {
	case int(Integer):
		if len(param.Range) == 2 {
			// When the length is 2, then determine whether it is within the range of values
			start, err1 := strconv.ParseInt(param.Range[0], 0, 64)
			end, err2 := strconv.ParseInt(param.Range[1], 0, 64)
			clusterValue, err3 := strconv.ParseInt(param.RealValue.ClusterValue, 0, 64)
			if err1 == nil && err2 == nil && err3 == nil && clusterValue >= start && clusterValue <= end {
				return true
			}
		} else {
			// When the length is 1 or greater than 2, iterate through enumerated values to determine if they are equal
			clusterValue, err := strconv.ParseInt(param.RealValue.ClusterValue, 0, 64)
			for i := 0; i < len(param.Range); i++ {
				val, err1 := strconv.ParseInt(param.Range[i], 0, 64)
				if err == nil && err1 == nil && clusterValue == val {
					return true
				}
			}
		}
	case int(String):
		// If the parameter is a string type and the range length is 1, it means it is an expression or a fixed display value
		if len(param.Range) <= 1 {
			return true
		}
		for _, enumValue := range param.Range {
			if param.RealValue.ClusterValue == enumValue {
				return true
			}
		}
	case int(Boolean):
		_, err := strconv.ParseBool(param.RealValue.ClusterValue)
		if err == nil {
			return true
		}
	case int(Float):
		start, err1 := strconv.ParseFloat(param.Range[0], 64)
		end, err2 := strconv.ParseFloat(param.Range[1], 64)
		clusterValue, err3 := strconv.ParseFloat(param.RealValue.ClusterValue, 64)
		if err1 == nil && err2 == nil && err3 == nil && clusterValue >= start && clusterValue <= end {
			return true
		}
	case int(Array):
		return true
	}
	return false
}

// DisplayFullParameterName
// @Description: display full parameter name
// @Parameter category
// @Parameter name
// @return string
func DisplayFullParameterName(category, name string) string {
	if category == "basic" {
		return name
	}
	return fmt.Sprintf("%s.%s", category, name)
}

// UnmarshalCovertArray
// @Description: unmarshal json string convert array
// @Parameter marshal json string
// @return []string
// @return error
func UnmarshalCovertArray(marshal string) ([]string, error) {
	result := make([]string, 0)
	if len(marshal) > 0 {
		if err := json.Unmarshal([]byte(marshal), &result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

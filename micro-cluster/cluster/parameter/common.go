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
	"reflect"
	"strconv"
	"strings"

	"github.com/pingcap/tiunimanager/library/spec"

	"github.com/pingcap/tiunimanager/common/structs"
)

const (
	contextClusterMeta             = "ClusterMeta"
	contextModifyParameters        = "ModifyParameters"
	contextHasApplyParameter       = "HasApplyParameter"
	contextMaintenanceStatusChange = "MaintenanceStatusChange"
	contextClusterConfigStr        = "ClusterConfigStr"
)

// Define unit conversion
var units = map[string]int64{
	// storage unit
	"B":  1,
	"KB": 1024,
	"MB": 1024 * 1024,
	"GB": 1024 * 1024 * 1024,
	"TB": 1024 * 1024 * 1024 * 1024,

	// time unit
	"us": 1,
	"ms": 1 * 1000,
	"s":  1000 * 1000,
	"m":  60 * 1000 * 1000,
	"h":  60 * 60 * 1000 * 1000,
	"d":  24 * 60 * 60 * 1000 * 1000,
}

// Define replace unit, used to replace the result value returned by the component API
var replaceUnits = map[string]string{
	"KiB": "KB",
	"MiB": "MB",
	"GiB": "GB",
	"TiB": "TB",
}

type UpdateParameterSource int

const (
	TiUP UpdateParameterSource = iota
	SQL
	TiUPAndSQL
	API
	TiUPAndAPI
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

type RangeType int

const (
	NoneRange RangeType = iota
	ContinuousRange
	DiscreteRange
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
	RangeType      int
	HasApply       int
	RealValue      structs.ParameterRealValue
}

type ClusterReboot int

const (
	NonReboot ClusterReboot = iota
	Reboot
)

type GlobalComponentConfig struct {
	TiDBClusterComponent spec.TiDBClusterComponent
	ConfigMap            map[string]interface{}
}

// ValidateRange
// @Description: validate parameter value by range field
// @Parameter param
// @return bool
func ValidateRange(param *ModifyClusterParameterInfo, hasModify bool) bool {
	// Determine if range is nil or an expression, continue the loop directly
	if param.Range == nil || len(param.Range) == 0 || param.RangeType == int(NoneRange) {
		return true
	}
	// If it is a modified parameter workflow, the value empty is skipped
	if hasModify && strings.TrimSpace(param.RealValue.ClusterValue) == "" {
		return true
	}
	// If you are modifying a parameter group and it is a modify apply parameter type with a null value, skip directly
	if !hasModify && param.HasApply == int(ModifyApply) && strings.TrimSpace(param.RealValue.ClusterValue) == "" {
		return true
	}
	switch param.Type {
	case int(Integer):
		clusterValue, err := strconv.ParseInt(param.RealValue.ClusterValue, 0, 64)
		if param.RangeType == int(ContinuousRange) && len(param.Range) == 2 {
			// When the range type is 1, then determine whether it is within the range of values
			start, err1 := strconv.ParseInt(param.Range[0], 0, 64)
			end, err2 := strconv.ParseInt(param.Range[1], 0, 64)
			if err1 == nil && err2 == nil && err == nil && clusterValue >= start && clusterValue <= end {
				return true
			}
		} else if param.RangeType == int(DiscreteRange) {
			// When the range type is discrete, iterate through enumerated values to determine if they are equal
			for i := 0; i < len(param.Range); i++ {
				val, err1 := strconv.ParseInt(param.Range[i], 0, 64)
				if err == nil && err1 == nil && clusterValue == val {
					return true
				}
			}
		}
	case int(String):
		// When the range type is 1, then determine whether it is within the range of values
		if param.RangeType == int(ContinuousRange) && len(param.Range) == 2 && len(param.UnitOptions) > 0 {
			clusterValue, ok := convertUnitValue(param.UnitOptions, param.RealValue.ClusterValue)
			start, ok1 := convertUnitValue(param.UnitOptions, param.Range[0])
			end, ok2 := convertUnitValue(param.UnitOptions, param.Range[1])
			if ok && ok1 && ok2 && clusterValue >= start && clusterValue <= end {
				return true
			}
		} else if param.RangeType == int(DiscreteRange) {
			for _, enumValue := range param.Range {
				if param.RealValue.ClusterValue == enumValue {
					return true
				}
			}
		}
	case int(Boolean):
		_, err := strconv.ParseBool(param.RealValue.ClusterValue)
		if err == nil {
			return true
		}
	case int(Float):
		clusterValue, err := strconv.ParseFloat(param.RealValue.ClusterValue, 64)
		if param.RangeType == int(ContinuousRange) && len(param.Range) == 2 {
			// When the range type is 1, then determine whether it is within the range of values
			start, err1 := strconv.ParseFloat(param.Range[0], 64)
			end, err2 := strconv.ParseFloat(param.Range[1], 64)
			if err1 == nil && err2 == nil && err == nil && clusterValue >= start && clusterValue <= end {
				return true
			}
		} else if param.RangeType == int(DiscreteRange) {
			for _, v := range param.Range {
				rv, err1 := strconv.ParseFloat(v, 64)
				if err == nil && err1 == nil && clusterValue == rv {
					return true
				}
			}
		}
	case int(Array):
		return true
	}
	return false
}

// convertUnitValue
// @Description: convert unit value
// @Parameter unitOptions ["KB", "MB", "GB"]
// @Parameter value 512MB
// @return int64 512 * 1024
// @return bool true
func convertUnitValue(unitOptions []string, value string) (int64, bool) {
	// Compatible with multiple units for unit conversion
	for _, u := range unitOptions {
		if strings.Contains(value, u) {
			// For example, 512MB, remove the unit MB, the value is 512
			v := strings.TrimRight(value, u)
			num, err := strconv.ParseInt(v, 0, 64)
			if multiples := units[u]; multiples > 0 && err == nil {
				return num * multiples, true
			}
		}
	}
	return -1, false
}

// DisplayFullParameterName
// @Description: display full parameter name
// @Parameter category
// @Parameter name
// @return string
func DisplayFullParameterName(category, name string) string {
	if category == "" || category == "basic" {
		return name
	}
	return fmt.Sprintf("%s.%s", category, name)
}

// FlattenedParameters
// @Description: flattened parameters
// @Parameter inputs
// @return map[string]string
// @return error
func FlattenedParameters(inputs map[string]interface{}) (map[string]string, error) {
	result := make(map[string]string)
	err := flattening(inputs, "", result)
	return result, err
}

// flattening
// @Description: Deep traversal flattening to get parameter names
// @Parameter object
// @Parameter name
// @Parameter result
// @return err
func flattening(object map[string]interface{}, name string, result map[string]string) (err error) {
	for k, v := range object {
		if name != "" {
			k = name + "." + k
		}
		if v == nil {
			// value is null, skip directly
			continue
		} else if strings.HasPrefix(reflect.TypeOf(v).String(), "map") {
			// value is an interface type
			if err = flattening(v.(map[string]interface{}), k, result); err != nil {
				return err
			}
		} else if strings.HasPrefix(reflect.TypeOf(v).String(), "[]interface") {
			// value is an array type
			b, err := json.Marshal(v)
			if err != nil {
				return err
			}
			result[k] = string(b)
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return nil
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

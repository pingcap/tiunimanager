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

/**
 * @Author: guobob
 * @Description:
 * @File:  compare.go
 * @Version: 1.0.0
 * @Date: 2021/12/10 11:58
 */

package utils

import (
 "reflect"
)

//Compares whether the interface values are equal
func CompareInterface(a, b interface{}) bool {
	t1 := reflect.ValueOf(a).Type().String()
	t2 := reflect.ValueOf(b).Type().String()
	if t1 != t2 {
		return false
	}
	switch a.(type) {
	case int:
		return a.(int) == b.(int)
	case uint:
		return a.(uint) == b.(uint)
	case int32:
		return a.(int32) == b.(int32)
	case uint32:
		return a.(uint32) == b.(uint32)
	case int64:
		return a.(int64) == b.(int64)
	case uint64:
		return a.(uint64) == b.(uint64)
	case string:
		return a.(string) == b.(string)
	default:
		return false
	}
}

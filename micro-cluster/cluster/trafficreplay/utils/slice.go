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
 * @File:  slice.go
 * @Version: 1.0.0
 * @Date: 2021/12/10 11:52
 */

package utils

import (
	"sort"
)

func SortSlice(a [][]interface{}) {
	var less = func(i, j int) bool {
		m := a[i][0]
		n := a[j][0]
		if m == nil {
			return true
		}
		if n == nil {
			return false
		}
		switch m.(type) {
		case int:
			return m.(int) < n.(int)
		case uint:
			return m.(uint) < n.(uint)
		case int32:
			return m.(int32) < n.(int32)
		case uint32:
			return m.(uint32) < n.(uint32)
		case int64:
			return m.(int64) < n.(int64)
		case uint64:
			return m.(uint64) < n.(uint64)
		case string:
			return m.(string) < n.(string)
		default:
			//fmt.Println("not supported type : " + reflect.ValueOf(m).Type().String())
			//Unsupported types so far we return false
			return false
		}
	}
	sort.Slice(a, less)
}

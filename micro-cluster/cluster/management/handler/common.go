/*******************************************************************************
 * @File: common
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10
*******************************************************************************/

package handler

import (
	"reflect"
	"strings"
)

func Contain(target interface{}, list interface{}) bool {
	if reflect.TypeOf(list).Kind() == reflect.Slice || reflect.TypeOf(list).Kind() == reflect.Array {
		listValue := reflect.ValueOf(list)
		for i := 0; i < listValue.Len(); i++ {
			if target == listValue.Index(i).Interface() {
				return true
			}
		}
	}
	if reflect.TypeOf(target).Kind() == reflect.String && reflect.TypeOf(list).Kind() == reflect.String {
		return strings.Contains(list.(string), target.(string))
	}
	return false
}

/**
 * @Author: guobob
 * @Description:
 * @File:  es_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/13 09:20
 */

package es

import (
	"fmt"
	"testing"
)

func Test_ESSearch(t *testing.T){
	e := &ESSearch{
		LastPos:0,
		BulkSize:10000,
		IndexName:"tiem-tidb-replay-2021.12.13",
	}
	fmt.Println(e)
}
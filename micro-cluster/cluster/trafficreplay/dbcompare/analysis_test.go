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
 * @File:  analysis_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/10 09:54
 */

package dbcompare

import (
	"fmt"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/trafficreplay/utils"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestAddKey(t *testing.T) {
	tsr := &TemplateSQLResult{
		SQLResults: make(map[uint64][]*SQLResult),
		Mu:         new(sync.RWMutex),
	}
	type arg struct {
		key                     uint64
		SQL                     string
		ExecSQL                 uint64
		ExecSQLSucc             uint64
		ExecSQLFail             uint64
		SQLCompareSucc          uint64
		SQLCompareFail          uint64
		SQLCompareErrNoFail     uint64
		SQLCompareRowCountFail  uint64
		SQLCompareRowDetailFail uint64
		SQLExecTimePr           uint64
		SQLExecTimeRr           uint64
	}

	args := make([]arg, 0)
	names := make([]string, 0)

	name1 := "and new key 2000"
	arg1 := arg{
		key:            2000,
		SQL:            "select * from t2 where id =?",
		ExecSQL:        1,
		ExecSQLFail:    0,
		ExecSQLSucc:    1,
		SQLCompareSucc: 1,
	}

	name2 := "and new key 1000"
	arg2 := arg{
		key:            1000,
		SQL:            "select * from t1 where id >?",
		ExecSQL:        1,
		ExecSQLFail:    0,
		ExecSQLSucc:    1,
		SQLCompareSucc: 1,
	}
	name3 := "and exist key 1000 with  no hash collisions"
	arg3 := arg{
		key:            1000,
		SQL:            "select * from t1 where id >?",
		ExecSQL:        1,
		ExecSQLFail:    0,
		ExecSQLSucc:    1,
		SQLCompareSucc: 1,
	}

	name4 := "and exist key 1000 with  hash collisions"
	arg4 := arg{
		key:            1000,
		SQL:            "select * from t1 where id =?",
		ExecSQL:        1,
		ExecSQLFail:    0,
		ExecSQLSucc:    1,
		SQLCompareSucc: 1,
	}
	args = append(args, arg1, arg2, arg3, arg4)
	names = append(names, name1, name2, name3, name4)

	for k, v := range args {
		fmt.Println(names[k])
		tsr.Add(v.key, v.SQL, v.ExecSQL, v.ExecSQLSucc, v.ExecSQLFail, v.SQLCompareSucc, v.SQLCompareFail,
			v.SQLCompareErrNoFail, v.SQLCompareRowCountFail, v.SQLCompareRowDetailFail, v.SQLExecTimePr,
			v.SQLExecTimeRr)
	}

}

func Test_GetPrAvgExecTime_with_divide_zero(t *testing.T) {
	tsr := &TemplateSQLResult{
		SQLResults: make(map[uint64][]*SQLResult),
		Mu:         new(sync.RWMutex),
	}

	key := uint64(100)
	sqlrs := []*SQLResult{
		&SQLResult{
			SQLTemplate:        "select * from t1 where id =?",
			SQLExecCount:       0,
			SQLExecTimeCountPr: 10000,
		},
	}
	tsr.SQLResults[key] = sqlrs

	prAvgExecTime := tsr.GetPrAvgExecTime(key, "select * from t1 where id =?")

	assert.New(t).Equal(prAvgExecTime, uint64(10000))

}

func Test_GetPrAvgExecTime(t *testing.T) {
	tsr := &TemplateSQLResult{
		SQLResults: make(map[uint64][]*SQLResult),
		Mu:         new(sync.RWMutex),
	}

	key := uint64(100)
	sqlrs := []*SQLResult{
		&SQLResult{
			SQLTemplate:        "select * from t1 where id =?",
			SQLExecCount:       2,
			SQLExecTimeCountPr: 10000,
		},
	}
	tsr.SQLResults[key] = sqlrs

	prAvgExecTime := tsr.GetPrAvgExecTime(key, "select * from t1 where id =?")

	assert.New(t).Equal(prAvgExecTime, uint64(5000))
}

func Test_GetPrAvgExecTime_with_no_found_key(t *testing.T) {
	tsr := &TemplateSQLResult{
		SQLResults: make(map[uint64][]*SQLResult),
		Mu:         new(sync.RWMutex),
	}

	key := uint64(100)
	sqlrs := []*SQLResult{
		&SQLResult{
			SQLTemplate:        "select * from t1 where id =?",
			SQLExecCount:       2,
			SQLExecTimeCountPr: 10000,
		},
	}
	tsr.SQLResults[key] = sqlrs

	prAvgExecTime := tsr.GetPrAvgExecTime(uint64(50), "select * from t1 where id =?")

	assert.New(t).Equal(prAvgExecTime, uint64(0))
}

func Test_GetPrAvgExecTime_found_but_sql_not_equal(t *testing.T) {
	tsr := &TemplateSQLResult{
		SQLResults: make(map[uint64][]*SQLResult),
		Mu:         new(sync.RWMutex),
	}

	key := uint64(100)
	sqlrs := []*SQLResult{
		&SQLResult{
			SQLTemplate:        "select * from t1 where id =?",
			SQLExecCount:       2,
			SQLExecTimeCountPr: 10000,
		},
	}
	tsr.SQLResults[key] = sqlrs

	prAvgExecTime := tsr.GetPrAvgExecTime(key, "select * from t1 where id >?")

	assert.New(t).Equal(prAvgExecTime, uint64(0))
}

func Test_GetRrAvgExecTime_with_divide_zero(t *testing.T) {
	tsr := &TemplateSQLResult{
		SQLResults: make(map[uint64][]*SQLResult),
		Mu:         new(sync.RWMutex),
	}

	key := uint64(100)
	sqlrs := []*SQLResult{
		&SQLResult{
			SQLTemplate:        "select * from t1 where id =?",
			SQLExecCount:       0,
			SQLExecTimeCountRr: 10000,
		},
	}
	tsr.SQLResults[key] = sqlrs

	rrAvgExecTime := tsr.GetRrAvgExecTime(key, "select * from t1 where id =?")

	assert.New(t).Equal(rrAvgExecTime, uint64(10000))

}

func Test_GetRrAvgExecTime(t *testing.T) {
	tsr := &TemplateSQLResult{
		SQLResults: make(map[uint64][]*SQLResult),
		Mu:         new(sync.RWMutex),
	}

	key := uint64(100)
	sqlrs := []*SQLResult{
		&SQLResult{
			SQLTemplate:        "select * from t1 where id =?",
			SQLExecCount:       2,
			SQLExecTimeCountRr: 10000,
		},
	}
	tsr.SQLResults[key] = sqlrs

	rrAvgExecTime := tsr.GetRrAvgExecTime(key, "select * from t1 where id =?")

	assert.New(t).Equal(rrAvgExecTime, uint64(5000))
}

func Test_GetRrAvgExecTime_with_no_found_key(t *testing.T) {
	tsr := &TemplateSQLResult{
		SQLResults: make(map[uint64][]*SQLResult),
		Mu:         new(sync.RWMutex),
	}

	key := uint64(100)
	sqlrs := []*SQLResult{
		&SQLResult{
			SQLTemplate:        "select * from t1 where id =?",
			SQLExecCount:       2,
			SQLExecTimeCountRr: 10000,
		},
	}
	tsr.SQLResults[key] = sqlrs

	rrAvgExecTime := tsr.GetRrAvgExecTime(uint64(50), "select * from t1 where id =?")

	assert.New(t).Equal(rrAvgExecTime, uint64(0))
}

func Test_GetRrAvgExecTime_found_but_sql_not_equal(t *testing.T) {
	tsr := &TemplateSQLResult{
		SQLResults: make(map[uint64][]*SQLResult),
		Mu:         new(sync.RWMutex),
	}

	key := uint64(100)
	sqlrs := []*SQLResult{
		&SQLResult{
			SQLTemplate:        "select * from t1 where id =?",
			SQLExecCount:       2,
			SQLExecTimeCountRr: 10000,
		},
	}
	tsr.SQLResults[key] = sqlrs

	rrAvgExecTime := tsr.GetRrAvgExecTime(key, "select * from t1 where id >?")

	assert.New(t).Equal(rrAvgExecTime, uint64(0))
}

func TestTemplateSQLResult_Add1(t *testing.T) {
	osrs := []OneSQLResult{
		{
			SQLHashKey:            2000,
			SQL:            "select * from t2 where id =?",
			SQLTemplate:            "select * from t2 where id =?",
			SQLType: utils.EventStmtExecute,
			SQLExecSucc:        1,
			SQLExecFail:    0,
		},
		{
			SQLHashKey:            1000,
			SQL:            "select * from t1 where id >?",
			SQLTemplate: "select * from t1 where id >?",
			SQLType: utils.EventStmtExecute,
			SQLExecSucc:        1,
			SQLExecFail:    0,
		},
		{
			SQLHashKey:            1000,
			SQL:            "select * from t1 where id >?",
			SQLTemplate: "select * from t1 where id >?",
			SQLType: utils.EventStmtExecute,
			SQLExecSucc:        1,
			SQLExecFail:    0,
		},
		{
			SQLHashKey:            1000,
			SQL:            "select * from t1 where id =?",
			SQLTemplate: "select * from t1 where id >?",
			SQLType: utils.EventStmtExecute,
			SQLExecSucc:        1,
			SQLExecFail:    0,
		},
	}

	tsr := NewTemplateSQLResult()

	for _,v := range osrs{
		tsr.Add1(&v)
	}
}
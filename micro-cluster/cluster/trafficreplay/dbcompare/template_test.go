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
 * @File:  template_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/10 09:32
 */

package dbcompare

import (
	"github.com/agiledragon/gomonkey"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/trafficreplay/utils"
	"github.com/pingcap/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse_With_Succ(t *testing.T) {
	sql := "select * from t1 where id =1;"
	_, err := parse(sql)
	assert.New(t).Nil(err)
}

func TestParse_With_fail(t *testing.T) {

	sql := "select id from "
	_, err := parse(sql)

	assert.New(t).NotNil(err)
}

func TestParse_With_Result_len_zero(t *testing.T) {

	sql := " "
	_, err := parse(sql)
	assert.New(t).NotNil(err)
}

func TestParse_With_Prepare_Succ(t *testing.T) {
	sql := "select * from t where id =?;"

	_, err := parse(sql)

	assert.New(t).Nil(err)
}

func TestCheckIsSelectStmt(t *testing.T) {
	type args struct {
		sql string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "sql is select ",
			args: args{
				sql: "select * from t1 where id =10;",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "sql is  select  for update ",
			args: args{
				sql: "select * from t1 where id =10 for update;",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "sql is not select ",
			args: args{
				sql: "insert into t (id,name) values (1,'aaa');",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "insert select  ",
			args: args{
				sql: "insert into t (id,name) select id ,name from t;",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "sql is not select  ",
			args: args{
				sql: "update t set id =100 where id =10;",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "select into outfile ",
			args: args{
				sql: "select id,name from test where id >100 into outfile 'test.txt';",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "parse sql fail ",
			args: args{
				sql: "select * from ;",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckIsSelectStmt(tt.args.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckIsSelectStmt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckIsSelectStmt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GenerateSQLTemplate(t *testing.T) {
	sql := "select id,name from t where id =1"
	sqlType := utils.EventQuery
	sqlExecSucc := uint64(1)
	sqlExecFail := uint64(0)
	sqlErrNoNotEqual := uint64(0)
	sqlRowCountNotEqual := uint64(0)
	sqlRowDetailNotEqual := uint64(0)
	sqlExecTimePr := uint64(1000000)
	sqlExecTimeRr := uint64(1100000)
	osr := NewOneSQLResult(sql, sqlType, sqlExecSucc, sqlExecFail, sqlErrNoNotEqual, sqlRowCountNotEqual,
		sqlRowDetailNotEqual, sqlExecTimePr, sqlExecTimeRr)

	osr.GenerateSQLTemplate()

	assert.New(t).Equal(osr.SQLTemplate, "select `id` , `name` from `t` where `id` = ?")
}

func Test_OneSQLResultInit_EventQuery_HashString_fail(t *testing.T) {
	sql := "select id,name from t where id =1"
	sqlType := utils.EventQuery
	sqlExecSucc := uint64(1)
	sqlExecFail := uint64(0)
	sqlErrNoNotEqual := uint64(0)
	sqlRowCountNotEqual := uint64(0)
	sqlRowDetailNotEqual := uint64(0)
	sqlExecTimePr := uint64(1000000)
	sqlExecTimeRr := uint64(1100000)
	osr := NewOneSQLResult(sql, sqlType, sqlExecSucc, sqlExecFail, sqlErrNoNotEqual, sqlRowCountNotEqual,
		sqlRowDetailNotEqual, sqlExecTimePr, sqlExecTimeRr)

	err1 := errors.New("hash string fail")

	patch := gomonkey.ApplyFunc(utils.HashString, func(s string) (uint64, error) {
		return 0, err1
	})
	defer patch.Reset()

	err := osr.OneSQLResultInit()

	assert.New(t).Equal(err, err1)

}

func Test_OneSQLResultInit_EventQuery_HashString_succ(t *testing.T) {
	sql := "select id,name from t where id =1"
	sqlType := utils.EventQuery
	sqlExecSucc := uint64(1)
	sqlExecFail := uint64(0)
	sqlErrNoNotEqual := uint64(0)
	sqlRowCountNotEqual := uint64(0)
	sqlRowDetailNotEqual := uint64(0)
	sqlExecTimePr := uint64(1000000)
	sqlExecTimeRr := uint64(1100000)
	osr := NewOneSQLResult(sql, sqlType, sqlExecSucc, sqlExecFail, sqlErrNoNotEqual, sqlRowCountNotEqual,
		sqlRowDetailNotEqual, sqlExecTimePr, sqlExecTimeRr)

	err := osr.OneSQLResultInit()

	assert.New(t).Nil(err)

}

func Test_OneSQLResultInit_EventStmtExecute_HashString_fail(t *testing.T) {
	sql := "select id,name from t where id =?"
	sqlType := utils.EventStmtExecute
	sqlExecSucc := uint64(1)
	sqlExecFail := uint64(0)
	sqlErrNoNotEqual := uint64(0)
	sqlRowCountNotEqual := uint64(0)
	sqlRowDetailNotEqual := uint64(0)
	sqlExecTimePr := uint64(1000000)
	sqlExecTimeRr := uint64(1100000)
	osr := NewOneSQLResult(sql, sqlType, sqlExecSucc, sqlExecFail, sqlErrNoNotEqual, sqlRowCountNotEqual,
		sqlRowDetailNotEqual, sqlExecTimePr, sqlExecTimeRr)

	err1 := errors.New("hash string fail")

	patch := gomonkey.ApplyFunc(utils.HashString, func(s string) (uint64, error) {
		return 0, err1
	})
	defer patch.Reset()

	err := osr.OneSQLResultInit()

	assert.New(t).Equal(err, err1)

}

func Test_OneSQLResultInit_EventStmtExecute_HashString_succ(t *testing.T) {
	sql := "select id,name from t where id =?"
	sqlType := utils.EventStmtExecute
	sqlExecSucc := uint64(1)
	sqlExecFail := uint64(0)
	sqlErrNoNotEqual := uint64(0)
	sqlRowCountNotEqual := uint64(0)
	sqlRowDetailNotEqual := uint64(0)
	sqlExecTimePr := uint64(1000000)
	sqlExecTimeRr := uint64(1100000)
	osr := NewOneSQLResult(sql, sqlType, sqlExecSucc, sqlExecFail, sqlErrNoNotEqual, sqlRowCountNotEqual,
		sqlRowDetailNotEqual, sqlExecTimePr, sqlExecTimeRr)

	err := osr.OneSQLResultInit()

	assert.New(t).Nil(err)

}

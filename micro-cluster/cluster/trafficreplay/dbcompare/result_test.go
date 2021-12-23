/**
 * @Author: guobob
 * @Description:
 * @File:  result_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/10 14:07
 */

package dbcompare

import (
	"encoding/json"
	"github.com/agiledragon/gomonkey"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/trafficreplay/utils"
	"github.com/pingcap/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func Test_InitOnePairResult(t *testing.T) {
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}

	opr.InitOnePairResult()
}

func TestOnePairResult_CheckErrNoEqual(t *testing.T) {
	opr := &OnePairResult{
		PrResult:  make([][]string, 0),
		RrResult:  make([][]string, 0),
		Params:    make([]interface{}, 0),
		PrErrorNo: 0,
		RrErrorNo: 0,
	}
	res := opr.CheckErrNoEqual()
	assert.New(t).True(res)

	opr.PrErrorNo = 1062
	res = opr.CheckErrNoEqual()
	assert.New(t).False(res)
}

func TestOnePairResult_CheckRowCountEqual(t *testing.T) {
	opr := &OnePairResult{
		PrResult:  make([][]string, 0),
		RrResult:  make([][]string, 0),
		Params:    make([]interface{}, 0),
		PrErrorNo: 0,
		RrErrorNo: 0,
	}
	res := opr.CheckRowCountEqual()
	assert.New(t).True(res)

	opr.PrResult = append(opr.RrResult, []string{"result"})

	res = opr.CheckRowCountEqual()
	assert.New(t).False(res)

}

func TestOnePairResult_CompareResDetail_with_row_count_len_zero(t *testing.T) {
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}
	res, err := opr.CompareResDetail()

	assert.New(t).True(res)
	assert.New(t).Nil(err)
}

func TestOnePairResult_CompareResDetail_with_HashResDetail_PrResult_fail(t *testing.T) {

	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}

	opr.PrResult = append(opr.RrResult, []string{"test"})
	opr.RrResult = append(opr.RrResult, []string{"test"})

	err1 := errors.New("hash string fail")
	patches := gomonkey.ApplyMethod(reflect.TypeOf(opr), "HashResDetail",
		func(_ *OnePairResult, result [][]string) ([][]interface{}, error) {
			return nil, err1
		})
	defer patches.Reset()

	res, err := opr.CompareResDetail()

	assert.New(t).Equal(err, err1)
	assert.New(t).False(res)
}

func TestOnePairResult_CompareResDetail_with_HashResDetail_RrResult_fail(t *testing.T) {

	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}
	opr.PrResult = append(opr.RrResult, []string{"test"})
	opr.RrResult = append(opr.RrResult, []string{"test"})

	ret := make([][]interface{}, 0)
	ret = append(ret, []interface{}{8})
	err1 := errors.New("hash string fail")
	a := gomonkey.OutputCell{
		Values: gomonkey.Params{ret, nil},
		Times:  1,
	}

	b := gomonkey.OutputCell{
		Values: gomonkey.Params{nil, err1},
		Times:  2,
	}
	outputs := make([]gomonkey.OutputCell, 0)
	outputs = append(outputs, a, b)

	patches := gomonkey.ApplyMethodSeq(reflect.TypeOf(opr), "HashResDetail",
		outputs)
	defer patches.Reset()

	res, err := opr.CompareResDetail()

	assert.New(t).False(res)
	assert.New(t).Equal(err, err1)
}

func TestOnePairResult_CompareResDetail_with_CompareData_fail(t *testing.T) {

	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}
	opr.PrResult = append(opr.RrResult, []string{"test"})
	opr.RrResult = append(opr.RrResult, []string{"test"})

	ret := make([][]interface{}, 0)
	ret = append(ret, []interface{}{8})
	patches := gomonkey.ApplyMethod(reflect.TypeOf(opr), "HashResDetail",
		func(_ *OnePairResult, result [][]string) ([][]interface{}, error) {
			return ret, nil
		})
	defer patches.Reset()

	err1 := errors.New("compare data fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(opr), "CompareData",
		func(_ *OnePairResult, pr [][]interface{}, rr [][]interface{}) error {
			return err1
		})
	defer patches1.Reset()

	res, err := opr.CompareResDetail()

	assert.New(t).False(res)
	assert.New(t).Equal(err, err1)
}

func TestOnePairResult_CompareResDetail_succ(t *testing.T) {

	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}
	opr.PrResult = append(opr.RrResult, []string{"test"})
	opr.RrResult = append(opr.RrResult, []string{"test"})

	ret := make([][]interface{}, 0)
	ret = append(ret, []interface{}{8})
	patches := gomonkey.ApplyMethod(reflect.TypeOf(opr), "HashResDetail",
		func(_ *OnePairResult, result [][]string) ([][]interface{}, error) {
			return ret, nil
		})
	defer patches.Reset()

	res, err := opr.CompareResDetail()

	assert.New(t).True(res)
	assert.New(t).Nil(err)
}

func TestOnePairResult_UnMarshalToStruct_fail(t *testing.T) {

	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}

	var s []byte = nil
	err := opr.UnMarshalToStruct(s)

	ast := assert.New(t)
	ast.NotNil(err)

}


func TestOnePairResult_UnMarshalToStruct_succ(t *testing.T) {

	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}

	s ,_:= json.Marshal(opr)

	err := opr.UnMarshalToStruct(s)

	ast := assert.New(t)
	ast.Nil(err)
}


func TestOnePairResult_DetermineNeedCompareResult(t *testing.T) {
	type fields struct {
		Type        uint64
		StmtID      uint64
		Params      []interface{}
		DB          string
		Query       string
		PrBeginTime uint64
		PrEndTime   uint64
		PrErrorNo   uint16
		PrErrorDesc string
		PrResult    [][]string
		RrBeginTime uint64
		RrEndTime   uint64
		RrErrorNo   uint16
		RrErrorDesc string
		RrResult    [][]string
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			name:"sql event is not EventQuery ,EventStmtExecute",
			fields:fields{
				Type:utils.EventQuit,
				Query:"",
			},
			want:false,
			wantErr: false,
		},
		{
			name:"sql event is  EventQuery and sql is select  ",
			fields:fields{
				Type:utils.EventQuery,
				Query:"select * from t1 where id=1",
			},
			want:true,
			wantErr: false,
		},
		{
			name:"sql event is  EventQuery and sql is select for update ",
			fields:fields{
				Type:utils.EventQuery,
				Query:"select * from t1 where id=1 for update ",
			},
			want:true,
			wantErr: false,
		},
		{
			name:"sql event is  EventQuery and sql is insert ",
			fields:fields{
				Type:utils.EventQuery,
				Query:"insert into t1 (id,name) values (1,'aa');",
			},
			want:false,
			wantErr: false,
		},
		{
			name:"sql event is  EventQuery and sql parse fail ",
			fields:fields{
				Type:utils.EventQuery,
				Query:"insert into t1 ;",
			},
			want:false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opr := &OnePairResult{
				Type:        tt.fields.Type,
				StmtID:      tt.fields.StmtID,
				Params:      tt.fields.Params,
				DB:          tt.fields.DB,
				Query:       tt.fields.Query,
				PrBeginTime: tt.fields.PrBeginTime,
				PrEndTime:   tt.fields.PrEndTime,
				PrErrorNo:   tt.fields.PrErrorNo,
				PrErrorDesc: tt.fields.PrErrorDesc,
				PrResult:    tt.fields.PrResult,
				RrBeginTime: tt.fields.RrBeginTime,
				RrEndTime:   tt.fields.RrEndTime,
				RrErrorNo:   tt.fields.RrErrorNo,
				RrErrorDesc: tt.fields.RrErrorDesc,
				RrResult:    tt.fields.RrResult,
			}
			got, err := opr.DetermineNeedCompareResult()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetermineNeedCompareResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetermineNeedCompareResult() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func  TestOnePairResult_GetPrExecTime( t *testing.T){
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}
	opr.PrEndTime = uint64(time.Now().Unix())
	opr.PrBeginTime = uint64(time.Now().Add(5 *time.Second).Unix())

	prExecTime := opr.GetPrExecTime()
	assert.New(t).Equal(prExecTime,uint64(0))

	opr.PrBeginTime = uint64(time.Now().Unix())
	opr.PrEndTime = uint64(time.Now().Add(5 * time.Second).Unix())

	prExecTime = opr.GetPrExecTime()

	assert.New(t).Greater(prExecTime,uint64(0))

}

func  TestOnePairResult_GetRrExecTime( t *testing.T){
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}
	opr.RrEndTime = uint64(time.Now().Unix())
	opr.RrBeginTime = uint64(time.Now().Add(5 *time.Second).Unix())

	rrExecTime := opr.GetRrExecTime()
	assert.New(t).Equal(rrExecTime,uint64(0))

	opr.RrBeginTime = uint64(time.Now().Unix())
	opr.RrEndTime = uint64(time.Now().Add(5 * time.Second).Unix())

	rrExecTime = opr.GetRrExecTime()

	assert.New(t).Greater(rrExecTime,uint64(0))

}

func  TestOnePairResult_HashResDetail_succ (t *testing.T ){
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}
	var result  [][]string
	rowRes1 := make([]string, 3)
	rowRes1 = append(rowRes1, "hello", "word", "program")
	rowRes2 := make([]string, 3)
	rowRes2 = append(rowRes2, "hello1", "word1", "program1")
	result = append(result,rowRes1,rowRes2)

	res, err := opr.HashResDetail(result)

	ast := assert.New(t)
	ast.GreaterOrEqual(len(res), 0)
	ast.Nil(err)
}

func  TestOnePairResult_HashResDetail_fail (t *testing.T ){
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}
	var result  [][]string
	rowRes1 := make([]string, 3)
	rowRes1 = append(rowRes1, "hello", "word", "program")
	rowRes2 := make([]string, 3)
	rowRes2 = append(rowRes2, "hello1", "word1", "program1")
	result = append(result,rowRes1,rowRes2)

	err1 := errors.New("hash string fail")

	patch := gomonkey.ApplyFunc(utils.HashString, func(s string) (uint64, error) {
		return 0, err1
	})
	defer patch.Reset()

	res, err := opr.HashResDetail(result)

	ast := assert.New(t)
	ast.Nil(res)
	ast.Equal(err, err1)
}

func TestOnePairResult_CompareData_fail(t *testing.T) {
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}
	var a , b [][]interface{}

	a = [][]interface{}{{"test"}}
	b = [][]interface{}{{"test"}}

	patch := gomonkey.ApplyFunc(utils.CompareInterface, func(a interface{}, b interface{}) bool{
		return false
	})
	defer patch.Reset()

	err := opr.CompareData(a,b)

	assert.New(t).NotNil(err)
}

func TestOnePairResult_CompareData_succ(t *testing.T) {
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}

	var a , b [][]interface{}
	a = [][]interface{}{{"test"}}
	b = [][]interface{}{{"test"}}

	err := opr.CompareData(a,b)

	assert.New(t).Nil(err)
}

func TestOnePairResult_CompareRes_CheckErrNoEqual_fail(t *testing.T) {
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}

	patches := gomonkey.ApplyMethod(reflect.TypeOf(opr), "CheckErrNoEqual",
		func(_ *OnePairResult) bool {
			return false
		})
	defer patches.Reset()

	scr := opr.CompareRes()
	assert.New(t).Equal(utils.ErrorNoNotEqual,scr.ErrCode)

}

func TestOnePairResult_CompareRes_CheckRowCountEqual_fail(t *testing.T) {
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}

	patches := gomonkey.ApplyMethod(reflect.TypeOf(opr), "CheckRowCountEqual",
		func(_ *OnePairResult) bool {
			return false
		})
	defer patches.Reset()

	scr := opr.CompareRes()
	assert.New(t).Equal(utils.ResultRowCountNotEqual,scr.ErrCode)

}

func TestOnePairResult_CompareRes_CompareResDetail_fail(t *testing.T) {
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}

	patches := gomonkey.ApplyMethod(reflect.TypeOf(opr), "CompareResDetail",
		func(_ *OnePairResult) (bool,error) {
			return false,nil
		})
	defer patches.Reset()

	scr := opr.CompareRes()
	assert.New(t).Equal(utils.ResultRowDetailNotEqual,scr.ErrCode)

}

func TestOnePairResult_CompareRes_succ(t *testing.T) {
	opr := &OnePairResult{
		PrResult: make([][]string, 0),
		RrResult: make([][]string, 0),
		Params:   make([]interface{}, 0),
	}

	patches := gomonkey.ApplyMethod(reflect.TypeOf(opr), "CompareResDetail",
		func(_ *OnePairResult) (bool,error) {
			return true,nil
		})
	defer patches.Reset()

	scr := opr.CompareRes()
	assert.New(t).Equal(utils.Success,scr.ErrCode)

}
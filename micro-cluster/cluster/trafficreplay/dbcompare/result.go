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
 * @File:  result.go
 * @Version: 1.0.0
 * @Date: 2021/12/10 11:25
 */

package dbcompare

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/trafficreplay/utils"
	"github.com/pingcap/errors"
	"strings"
)

type SqlCompareRes struct {
	Type     string        `json:"sqltype"`
	Sql      string        `json:"sql"`
	Values   []interface{} `json:"values"`
	ErrCode  int           `json:"errcode"`
	ErrDesc  string        `json:"errdesc"`
	RrValues [][]string    `json:"rrValues"`
	PrValues [][]string    `json:"prValues"`
}

type OnePairResult struct {
	Type   uint64        `json:"type"`
	StmtID uint64        `json:"stmtID,omitempty"`
	Params []interface{} `json:"params,omitempty"`
	DB     string        `json:"db,omitempty"`
	Query  string        `json:"query,omitempty"`
	//read from packet
	PrBeginTime uint64     `json:"pr-begin-time"`
	PrEndTime   uint64     `json:"pr-end-time"`
	PrErrorNo   uint16     `json:"pr-error-no"`
	PrErrorDesc string     `json:"pr-error-desc"`
	PrResult    [][]string `json:"pr-result"`
	//read from replay server
	RrBeginTime uint64     `json:"rr-begin-time"`
	RrEndTime   uint64     `json:"rr-end-time"`
	RrErrorNo   uint16     `json:"rr-error-no"`
	RrErrorDesc string     `json:"rr-error-desc"`
	RrResult    [][]string `json:"rr-result"`
}

func (opr *OnePairResult) InitOnePairResult() {
	opr.Type = 0
	opr.StmtID = 0
	opr.Params = opr.Params[:0]
	opr.DB = ""
	opr.Query = ""
	opr.PrBeginTime = 0
	opr.PrEndTime = 0
	opr.PrErrorNo = 0
	opr.PrErrorDesc = ""
	opr.PrResult = opr.PrResult[:0][:0]
	opr.RrBeginTime = 0
	opr.RrEndTime = 0
	opr.RrErrorNo = 0
	opr.RrErrorDesc = ""
	opr.RrResult = opr.RrResult[:0][:0]
}

func (opr *OnePairResult) UnMarshalToStruct(s []byte) error {
	err := json.Unmarshal(s, opr)
	if err != nil {
		return err
	}
	return nil
}

//DetermineNeedCompareResult :The current version only compares Select statements for
//EventQuery and EventStmtExecute events
func (opr *OnePairResult) DetermineNeedCompareResult() (bool, error) {
	if opr.Type == utils.EventQuery || opr.Type == utils.EventStmtExecute {
		isSelect, err := CheckIsSelectStmt(opr.Query)
		if err != nil {
			return false, err
		}
		if isSelect == false {
			return false, nil
		}
	} else {
		return false, nil
	}
	return true, nil
}

func (opr *OnePairResult)GetPrExecTime() uint64{
	var prSqlExecTime uint64
	if opr.PrBeginTime < opr.PrEndTime {
		prSqlExecTime = opr.PrEndTime - opr.PrBeginTime
	} else {
		prSqlExecTime = 0
	}
	return prSqlExecTime
}

func (opr *OnePairResult)GetRrExecTime() uint64{
	var rrSqlExecTime uint64
	if opr.RrBeginTime < opr.RrEndTime {
		rrSqlExecTime = opr.RrEndTime - opr.RrBeginTime
	} else {
		rrSqlExecTime = 0
	}
	return rrSqlExecTime
}

func(opr *OnePairResult)CheckErrNoEqual()bool{
	return opr.PrErrorNo == opr.RrErrorNo
}

func (opr *OnePairResult)CheckRowCountEqual()bool{
	return len(opr.PrResult) == len(opr.RrResult)
}

func (opr *OnePairResult) HashResDetail(result [][]string) ([][]interface{}, error) {
	//var rowStr string
	res := make([][]interface{}, 0,len(result))
	for i := range result {
		rowStr := strings.Join(result[i], "")
		rowv := make([]interface{}, 0)
		v, err := utils.HashString(rowStr)
		if err != nil {
			return nil, err
		}
		rowv = append(rowv, v, uint64(i))
		res = append(res, rowv)
	}
	return res, nil
}

func (opr *OnePairResult) CompareData(a, b [][]interface{}) error {
	for i := range a {
		ok := utils.CompareInterface(a[i][0], b[i][0])
		if !ok {
			err := errors.New("compare row string hash value not equal ")
			return err
		}
	}
	return nil
}

func (opr *OnePairResult) CompareResDetail() (bool, error) {
	if len(opr.RrResult) == 0{
		return true , nil
	}
	pr, err := opr.HashResDetail(opr.PrResult)
	if err != nil {
		return false, err
	}
	utils.SortSlice(pr)
	rr, err := opr.HashResDetail(opr.RrResult)
	if err != nil {
		return false, err
	}
	utils.SortSlice(rr)

	err = opr.CompareData(pr, rr)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (opr *OnePairResult) CompareRes() *SqlCompareRes {
	scr := new(SqlCompareRes)
	scr.Type = utils.TypeString(opr.Type)
	scr.Sql = opr.Query
	scr.Values = opr.Params
	if !opr.CheckErrNoEqual(){
		scr.ErrCode = utils.ErrorNoNotEqual
		scr.ErrDesc=fmt.Sprintf("sql exec on product and simulation tidb server error no  not equal %v-%v",
			opr.PrErrorNo,opr.RrErrorNo)
		return scr
	}
	if !opr.CheckRowCountEqual(){
		scr.ErrCode=utils.ResultRowCountNotEqual
		scr.ErrDesc=fmt.Sprintf("sql exec on product and simulation tidb server  result row count not equal %v-%v",
			len(opr.PrResult),len(opr.RrResult))
		return scr
	}

	//TODO handle err
	ok , _ := opr.CompareResDetail()
	if !ok {
		scr.ErrCode=utils.ResultRowDetailNotEqual
		scr.ErrDesc=fmt.Sprintf("sql exec on product and simulation tidb server  result row detail not equal %v-%v",
			opr.PrResult,opr.RrResult)
		return scr
	}

	return scr
}

func (opr *OnePairResult) CheckExecTime(ctx context.Context,rrAvgExecTime uint64) {
	//In the current version of the same sql template,
	//if the playback time in the simulation environment is
	//more than twice the average of the execution time in the
	//production environment, we consider the performance degradation
	//TODO : Performance judgement ratio configurable
	if rrAvgExecTime > opr.GetPrExecTime() * 2 {
		framework.LogWithContext(ctx).Errorf("sql %v with params %v exec on simulation environment performance degradation",
			opr.Query,opr.Params)
	}
}

func (opr *OnePairResult) AddTemplate(ctx context.Context,tsr *TemplateSQLResult,
	sqlExecSucc, sqlExecFail, sqlErrNoNotEqual,
	sqlRowCountNotEqual, sqlRowDetailNotEqual uint64) {
	osr := NewOneSQLResult(opr.Query, opr.Type, sqlExecSucc, sqlExecFail, sqlErrNoNotEqual,
		sqlRowCountNotEqual, sqlRowDetailNotEqual, opr.GetPrExecTime(), opr.GetRrExecTime())
	err := osr.OneSQLResultInit()
	if err != nil {
		fmt.Println(err)
		return
	}
	tsr.Add1(osr)

	//prAvgExecTime:= tsr.GetPrAvgExecTime(osr.SQLHashKey,osr.SQLTemplate)
	rrAvgExecTIme:= tsr.GetRrAvgExecTime(osr.SQLHashKey,osr.SQLTemplate)
	opr.CheckExecTime(ctx,rrAvgExecTIme)
}


func (opr *OnePairResult) CompareResAddTemplate(ctx context.Context,tsr *TemplateSQLResult)  {

	var sqlExecSucc, sqlExecFail, sqlErrNoNotEqual,
	sqlRowCountNotEqual, sqlRowDetailNotEqual uint64

	if !opr.CheckErrNoEqual(){
		ErrCode := utils.ErrorNoNotEqual
		ErrDesc :=fmt.Sprintf("sql exec on product and simulation tidb server error no  not equal %v-%v",
			opr.PrErrorNo,opr.RrErrorNo)
		framework.LogWithContext(ctx).Errorf("compare result fail ,errcode:%v,errdesc:%v",
			ErrCode,ErrDesc)

		sqlExecFail = 1
		sqlErrNoNotEqual = 1
		opr.AddTemplate(ctx,tsr,
			sqlExecSucc, sqlExecFail, sqlErrNoNotEqual,
			sqlRowCountNotEqual, sqlRowDetailNotEqual)
		return
	}
	if !opr.CheckRowCountEqual(){
		ErrCode :=utils.ResultRowCountNotEqual
		ErrDesc :=fmt.Sprintf("sql exec on product and simulation tidb server  result row count not equal %v-%v",
			len(opr.PrResult),len(opr.RrResult))
		framework.LogWithContext(ctx).Errorf("compare result fail ,errcode:%v,errdesc:%v",
			ErrCode,ErrDesc)

		sqlExecSucc = 1
		sqlRowCountNotEqual = 1
		opr.AddTemplate(ctx,tsr,
			sqlExecSucc, sqlExecFail, sqlErrNoNotEqual,
			sqlRowCountNotEqual, sqlRowDetailNotEqual)
		return
	}

	//TODO handle err
	ok , err := opr.CompareResDetail()
	if err !=nil{
		framework.LogWithContext(ctx).Errorf("compare result detail fail ,%v",err)
	}
	if !ok {
		ErrCode:=utils.ResultRowDetailNotEqual
		ErrDesc:=fmt.Sprintf("sql exec on product and simulation tidb server  result row detail not equal %v-%v",
			opr.PrResult,opr.RrResult)
		framework.LogWithContext(ctx).Errorf("compare result fail ,errcode:%v,errdesc:%v",
			ErrCode,ErrDesc)

		sqlExecSucc = 1
		sqlRowDetailNotEqual = 1
		opr.AddTemplate(ctx,tsr,
			sqlExecSucc, sqlExecFail, sqlErrNoNotEqual,
			sqlRowCountNotEqual, sqlRowDetailNotEqual)
		return
	}

	sqlExecSucc = 1
	opr.AddTemplate(ctx,tsr,
		sqlExecSucc, sqlExecFail, sqlErrNoNotEqual,
		sqlRowCountNotEqual, sqlRowDetailNotEqual)
}

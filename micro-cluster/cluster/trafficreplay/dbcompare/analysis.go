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
 * @File:  analysis.go
 * @Version: 1.0.0
 * @Date: 2021/12/9 21:21
 */

package dbcompare

import "sync"

// Store the results of a sql template
type SQLResult struct {
	//sql template string
	SQLTemplate string `json:"sql"`
	//current sql template execution times
	SQLExecCount uint64 `json:"sql_exec_count"`
	//current sql template execution success times
	SQLExecSuccCount uint64 `json:"sql_exec_succ_count"`
	//current sql template execution fail times
	SQLExecFailCount uint64 `json:"sql_exec_fail_count"`
	//current sql result compare success count
	SQLCompareSuccCount uint64 `json:"sql_compare_succ_count"`
	//current sql result compare fail count
	SQLCompareFailCount uint64 `json:"sql_compare_fail_count"`
	//current sql result compare error no fail count
	SQLCompareErrNoFailCount uint64 `json:"sql_compare_errno_fail_count"`
	//current sql result compare row count fail count
	SQLCompareRowCountFailCount uint64 `json:"sql_compare_row_count_fail_count"`
	//current sql result compare row detail fail count
	SQLCompareRowDetailFailCount uint64 `json:"sql_compare_row_detail_fail_count"`
	//sum of current sql execution time on production environment
	SQLExecTimeCountPr uint64 `json:"sql_exec_time_count_pr"`
	//sum of current sql execution time on simulation environment
	SQLExecTimeCountRr uint64 `json:"sql_exec_time_count_rr"`
	//statistics on the number of inconsistent execution times of sql statements
	//in the production and simulation environments
	SQLExecTimeCompareFail uint64 `json:"sql_exec_time_compare_fail"`
	//current sql execution time deterioration statistics min
	SQLExecTimeStandard uint64 `json:"sql_exec_time_standard"`
}

type TemplateSQLResult struct {
	// map key : SQLTemplates hash value
	// map value : slice SQLResult, preventing hash collisions used to cache all SQL template execution results
	SQLResults map[uint64][]*SQLResult
	//for protecting SQL template storage structures
	Mu *sync.RWMutex
}


func NewTemplateSQLResult() *TemplateSQLResult{
	t := new(TemplateSQLResult)
	t.Mu= new(sync.RWMutex)
	t.SQLResults = make(map[uint64][]*SQLResult)
	return t
}

//add key to map
func (tsr *TemplateSQLResult) Add(key uint64, SQL string, ExecSQL, ExecSQLSucc, ExecSQLFail, SQLCompareSucc,
	SQLCompareFail, SQLCompareErrNoFail, SQLCompareRowCountFail,
	SQLCompareRowDetailFail, SQLExecTimePr, SQLExecTimeRr uint64) {
	tsr.Mu.Lock()
	defer tsr.Mu.Unlock()
	if v, ok := tsr.SQLResults[key]; ok {
		var found = false
		for i := range v {
			if SQL == (*tsr.SQLResults[key][i]).SQLTemplate {
				(*tsr.SQLResults[key][i]).SQLExecCount += ExecSQL
				(*tsr.SQLResults[key][i]).SQLExecSuccCount += ExecSQLSucc
				(*tsr.SQLResults[key][i]).SQLExecFailCount += ExecSQLFail
				(*tsr.SQLResults[key][i]).SQLCompareSuccCount += SQLCompareSucc
				(*tsr.SQLResults[key][i]).SQLCompareFailCount += SQLCompareFail
				(*tsr.SQLResults[key][i]).SQLCompareErrNoFailCount += SQLCompareErrNoFail
				(*tsr.SQLResults[key][i]).SQLCompareRowCountFailCount += SQLCompareRowCountFail
				(*tsr.SQLResults[key][i]).SQLCompareRowDetailFailCount += SQLCompareRowDetailFail
				(*tsr.SQLResults[key][i]).SQLExecTimeCountPr += SQLExecTimePr
				(*tsr.SQLResults[key][i]).SQLExecTimeCountRr += SQLExecTimeRr
				found = true
				break
			}
		}
		if !found {
			tsr.SQLResults[key] = append(tsr.SQLResults[key], &SQLResult{
				SQLTemplate:                  SQL,
				SQLExecCount:                 ExecSQL,
				SQLExecSuccCount:             ExecSQLSucc,
				SQLExecFailCount:             ExecSQLFail,
				SQLCompareSuccCount:          SQLCompareSucc,
				SQLCompareFailCount:          SQLCompareFail,
				SQLCompareErrNoFailCount:     SQLCompareErrNoFail,
				SQLCompareRowCountFailCount:  SQLCompareRowCountFail,
				SQLCompareRowDetailFailCount: SQLCompareRowDetailFail,
				SQLExecTimeCountPr:           SQLExecTimePr,
				SQLExecTimeCountRr:           SQLExecTimeRr,
			})
		}
	} else {
		sliceSQLResult := make([]*SQLResult, 0)
		sliceSQLResult = append(sliceSQLResult, &SQLResult{
			SQLTemplate:                  SQL,
			SQLExecCount:                 ExecSQL,
			SQLExecSuccCount:             ExecSQLSucc,
			SQLExecFailCount:             ExecSQLFail,
			SQLCompareSuccCount:          SQLCompareSucc,
			SQLCompareFailCount:          SQLCompareFail,
			SQLCompareErrNoFailCount:     SQLCompareErrNoFail,
			SQLCompareRowCountFailCount:  SQLCompareRowCountFail,
			SQLCompareRowDetailFailCount: SQLCompareRowDetailFail,
			SQLExecTimeCountPr:           SQLExecTimePr,
			SQLExecTimeCountRr:           SQLExecTimeRr,
		})
		tsr.SQLResults[key] = sliceSQLResult
	}

	return
}

func (tsr *TemplateSQLResult) Add1(osr *OneSQLResult) {
	key := osr.SQLHashKey
	SQL := osr.SQLTemplate
	ExecSQL := uint64(1)
	ExecSQLSucc := osr.SQLExecSucc
	ExecSQLFail := osr.SQLExecFail
	SQLCompareSucc := 1 - osr.SQLErrNoNotEqual - osr.SQLRowCountNotEqual - osr.SQLRowDetailNotEqual
	SQLCompareFail := 1 - SQLCompareSucc
	SQLCompareErrNoFail := osr.SQLErrNoNotEqual
	SQLCompareRowCountFail := osr.SQLRowCountNotEqual
	SQLCompareRowDetailFail := osr.SQLRowDetailNotEqual
	SQLExecTimePr := osr.SQLExecTimePr
	SQLExecTimeRr := osr.SQLExecTimeRr

	tsr.Mu.Lock()
	defer tsr.Mu.Unlock()
	if v, ok := tsr.SQLResults[key]; ok {
		var found = false
		for i := range v {
			if SQL == (*tsr.SQLResults[key][i]).SQLTemplate {
				(*tsr.SQLResults[key][i]).SQLExecCount += ExecSQL
				(*tsr.SQLResults[key][i]).SQLExecSuccCount += ExecSQLSucc
				(*tsr.SQLResults[key][i]).SQLExecFailCount += ExecSQLFail
				(*tsr.SQLResults[key][i]).SQLCompareSuccCount += SQLCompareSucc
				(*tsr.SQLResults[key][i]).SQLCompareFailCount += SQLCompareFail
				(*tsr.SQLResults[key][i]).SQLCompareErrNoFailCount += SQLCompareErrNoFail
				(*tsr.SQLResults[key][i]).SQLCompareRowCountFailCount += SQLCompareRowCountFail
				(*tsr.SQLResults[key][i]).SQLCompareRowDetailFailCount += SQLCompareRowDetailFail
				(*tsr.SQLResults[key][i]).SQLExecTimeCountPr += SQLExecTimePr
				(*tsr.SQLResults[key][i]).SQLExecTimeCountRr += SQLExecTimeRr
				found = true
				break
			}
		}
		if !found {
			tsr.SQLResults[key] = append(tsr.SQLResults[key], &SQLResult{
				SQLTemplate:                  SQL,
				SQLExecCount:                 ExecSQL,
				SQLExecSuccCount:             ExecSQLSucc,
				SQLExecFailCount:             ExecSQLFail,
				SQLCompareSuccCount:          SQLCompareSucc,
				SQLCompareFailCount:          SQLCompareFail,
				SQLCompareErrNoFailCount:     SQLCompareErrNoFail,
				SQLCompareRowCountFailCount:  SQLCompareRowCountFail,
				SQLCompareRowDetailFailCount: SQLCompareRowDetailFail,
				SQLExecTimeCountPr:           SQLExecTimePr,
				SQLExecTimeCountRr:           SQLExecTimeRr,
			})
		}
	} else {
		sliceSQLResult := make([]*SQLResult, 0)
		sliceSQLResult = append(sliceSQLResult, &SQLResult{
			SQLTemplate:                  SQL,
			SQLExecCount:                 ExecSQL,
			SQLExecSuccCount:             ExecSQLSucc,
			SQLExecFailCount:             ExecSQLFail,
			SQLCompareSuccCount:          SQLCompareSucc,
			SQLCompareFailCount:          SQLCompareFail,
			SQLCompareErrNoFailCount:     SQLCompareErrNoFail,
			SQLCompareRowCountFailCount:  SQLCompareRowCountFail,
			SQLCompareRowDetailFailCount: SQLCompareRowDetailFail,
			SQLExecTimeCountPr:           SQLExecTimePr,
			SQLExecTimeCountRr:           SQLExecTimeRr,
		})
		tsr.SQLResults[key] = sliceSQLResult
	}

	return
}

func (tsr *TemplateSQLResult) GetPrAvgExecTime(key uint64, SQLTemplate string) uint64 {

	if v, ok := tsr.SQLResults[key]; !ok {
		return 0
	} else {
		for i := range v {
			if v[i].SQLTemplate == SQLTemplate {
				//Prevent divide zero
				sqlExecCount := v[i].SQLExecCount
				if sqlExecCount <= 0 {
					return v[i].SQLExecTimeCountPr
				}
				return v[i].SQLExecTimeCountPr / sqlExecCount
			}
		}
	}

	return 0
}

func (tsr *TemplateSQLResult) GetRrAvgExecTime(key uint64, SQLTemplate string) uint64 {

	if v, ok := tsr.SQLResults[key]; !ok {
		return 0
	} else {
		for i := range v {
			if v[i].SQLTemplate == SQLTemplate {
				//Prevent divide zero
				sqlExecCount := v[i].SQLExecCount
				if sqlExecCount <= 0 {
					return v[i].SQLExecTimeCountRr
				}
				return v[i].SQLExecTimeCountRr / sqlExecCount
			}
		}
	}

	return 0
}


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
 * @File:  es.go
 * @Version: 1.0.0
 * @Date: 2021/12/13 09:06
 */

package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/log"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/trafficreplay/dbcompare"
	"github.com/pingcap/errors"
)

type ESSearch struct {
	LastPos  int
	BulkSize int
	IndexName string
}

func (e *ESSearch)Search(ctx context.Context) (*esapi.Response, error) {
	var buf bytes.Buffer
	//get elasticsearch client
	es := framework.Current.GetElasticsearchClient()
	if es ==nil{
		err:=errors.New("get elasticsearch client fail,elasticsearch client is nil")
		framework.LogWithContext(ctx).Errorf(err.Error())
		return nil,err
	}

	return  es.Search(e.IndexName,&buf,e.LastPos,e.BulkSize)
}

func (e *ESSearch)AnalysisResult(ctx context.Context,resp *esapi.Response) ([]dbcompare.OnePairResult,error){
	if resp.IsError() || resp.StatusCode != 200 {
		errs := fmt.Sprintf("search %v failed! response [%v]", e.IndexName,resp.String())
		framework.LogWithContext(ctx).Errorf(errs)
		return nil,errors.New(errs)
	}
	var esResult log.ElasticSearchVo
	if err := json.NewDecoder(resp.Body).Decode(&esResult); err != nil {
		framework.LogWithContext(ctx).Errorf("Error parsing the response body: %v", err)
		return nil,err
	}

	//save db-replay output
	osrs := make([]dbcompare.OnePairResult,len(esResult.Hits.Hits))

	for k, hit := range esResult.Hits.Hits {
		var osr  dbcompare.OnePairResult
		marshal, err := json.Marshal(hit.Source)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("search db-replay error: %v", err)
			return nil,err
		}

		//Parsing the db-replay output into struct
		err = json.Unmarshal(marshal, &osr)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("search db-replay error: %v", err)
			return nil,err
		}
		osrs[k]=osr
	}
	return osrs,nil
}
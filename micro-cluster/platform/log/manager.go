/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

/*******************************************************************************
 * @File: manager.go
 * @Description: platform log implements
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/3/10 16:24
*******************************************************************************/

package log

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/pingcap-inc/tiunimanager/message"

	"github.com/pingcap-inc/tiunimanager/micro-cluster/cluster/log"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/pingcap-inc/tiunimanager/util/convert"

	"github.com/pingcap-inc/tiunimanager/common/errors"
	"github.com/pingcap-inc/tiunimanager/library/framework"
	"github.com/pingcap-inc/tiunimanager/proto/clusterservices"
)

//  search log index prefix
const logIndexPrefix = "em-system-logs-*"

type Manager struct{}

var manager *Manager
var once sync.Once

func NewManager() *Manager {
	once.Do(func() {
		if manager == nil {
			manager = &Manager{}
		}
	})
	return manager
}

// QueryPlatformLog
// @Description: query platform log
// @Receiver m
// @Parameter ctx
// @Parameter req
// @return resp
// @return page
// @return err
func (m Manager) QueryPlatformLog(ctx context.Context, req message.QueryPlatformLogReq) (resp message.QueryPlatformLogResp, page *clusterservices.RpcPage, err error) {
	buf, err := prepareSearchParams(ctx, req)
	if err != nil {
		return resp, page, err
	}
	esResp, err := framework.Current.GetElasticsearchClient().Search(logIndexPrefix, &buf, (req.Page-1)*req.PageSize, req.PageSize)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query em platform log, search es err: %v", err)
		return resp, page, errors.NewErrorf(errors.TIUNIMANAGER_LOG_QUERY_FAILED, errors.TIUNIMANAGER_LOG_QUERY_FAILED.Explain(), err)
	}
	return handleResult(ctx, req, esResp)
}

func prepareSearchParams(ctx context.Context, req message.QueryPlatformLogReq) (buf bytes.Buffer, err error) {
	buf = bytes.Buffer{}
	query, err := buildSearchPlatformReqParams(req)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query em platform log, build search params err: %v", err)
		return buf, errors.NewErrorf(errors.TIUNIMANAGER_LOG_QUERY_FAILED, errors.TIUNIMANAGER_LOG_QUERY_FAILED.Explain(), err)
	}
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return buf, errors.NewErrorf(errors.TIUNIMANAGER_LOG_QUERY_FAILED, "query em platform log, prepare search param error: %v", err)
	}
	return buf, nil
}

func handleResult(ctx context.Context, req message.QueryPlatformLogReq, esResp *esapi.Response) (resp message.QueryPlatformLogResp, page *clusterservices.RpcPage, err error) {
	if esResp.IsError() || esResp.StatusCode != 200 {
		framework.LogWithContext(ctx).Errorf("query em platform log, search es err: %v", err)
		return resp, page, errors.NewErrorf(errors.TIUNIMANAGER_LOG_QUERY_FAILED, errors.TIUNIMANAGER_LOG_QUERY_FAILED.Explain(), esResp.String())
	}
	var esResult log.ElasticSearchResult
	if err = json.NewDecoder(esResp.Body).Decode(&esResult); err != nil {
		framework.LogWithContext(ctx).Errorf("query em platform log, decoder err: %v", err)
		return resp, page, errors.NewErrorf(errors.TIUNIMANAGER_LOG_QUERY_FAILED, errors.TIUNIMANAGER_LOG_QUERY_FAILED.Explain(), err)
	}

	resp = message.QueryPlatformLogResp{
		Took: esResult.Took,
	}
	resp.Results = make([]message.PlatformLogItem, 0)
	for _, hit := range esResult.Hits.Hits {
		var hitItem message.SearchPlatformLogSourceItem
		if err := convert.ConvertObj(hit.Source, &hitItem); err != nil {
			return resp, page, errors.NewErrorf(errors.TIUNIMANAGER_CONVERT_OBJ_FAILED, "query em platform log, convert obj error: %v", err)
		}

		resp.Results = append(resp.Results, message.PlatformLogItem{
			Index:       hit.Index,
			Id:          hit.Id,
			Level:       hitItem.Level,
			TraceId:     hitItem.TraceId,
			MicroMethod: hitItem.MicroMethod,
			Message:     hitItem.Msg,
			Timestamp:   time.Unix(hitItem.Timestamp.Unix(), 0).Format(log.DateFormat),
		})
	}

	page = &clusterservices.RpcPage{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Total:    int32(esResult.Hits.Total.Value),
	}
	return resp, page, nil
}

// buildSearchPlatformReqParams build request elasticsearch parameters
//response e.g.:
//{
//  "query": {
//    "bool": {
//      "filter": [
//        {
//          "term": {
//            "level": "info"
//          }
//        },
//        {
//          "match": {
//            "msg": "to do something"
//          }
//        },
//        {
//          "range": {
//            "@timestamp": {
//              "gte": 1631840244000,
//              "lte": 1631876244000
//            }
//          }
//        }
//      ]
//    }
//  }
//}
func buildSearchPlatformReqParams(req message.QueryPlatformLogReq) (map[string]interface{}, error) {
	filters := make([]interface{}, 0)

	// option parameters: traceId
	if req.TraceId != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"Em-X-Trace-Id": req.TraceId,
			},
		})
	}
	// option parameters: level
	if req.Level != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"level": req.Level,
			},
		})
	}
	// option parameters: msg
	if req.Message != "" {
		filters = append(filters, map[string]interface{}{
			"match": map[string]interface{}{
				"msg": req.Message,
			},
		})
	}
	// option parameters: startTime, endTime
	tsFilter, err := log.FilterTimestamp(req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	if len(tsFilter) > 0 {
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"@timestamp": tsFilter,
			},
		})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filters,
			},
		},
	}
	return query, nil
}

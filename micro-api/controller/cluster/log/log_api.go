
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
 *                                                                            *
 ******************************************************************************/

package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pingcap-inc/tiem/micro-api/controller"
	"google.golang.org/grpc/codes"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/framework"
)

const (
	dateFormat         = "2006-01-02 15:04:05"
	tidbLogIndexPrefix = "tiem-tidb-cluster-*"
)

// SearchTiDBLog
// @Summary search tidb log
// @Description search tidb log
// @Tags logs
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Param searchReq query SearchTiDBLogReq false "query request"
// @Success 200 {object} controller.ResultWithPage{data=SearchTiDBLogRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /logs/tidb/{clusterId} [get]
func SearchTiDBLog(c *gin.Context) {
	clusterId := c.Param("clusterId")

	var reqParams SearchTiDBLogReq
	err := c.ShouldBindQuery(&reqParams)
	if err != nil {
		framework.Log().Error(err)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}

	var buf bytes.Buffer
	query, err := buildSearchTiDBReqParams(clusterId, reqParams)
	if err != nil {
		framework.Log().Error(err)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		framework.Log().Errorf("Error encoding query: %s", err)
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}

	// default value valid
	if reqParams.Page <= 1 {
		reqParams.Page = 1
	}
	if reqParams.PageSize <= 0 {
		reqParams.PageSize = 10
	}
	from := (reqParams.Page - 1) * reqParams.PageSize
	esClient := framework.Current.GetElasticsearchClient()
	resp, err := esClient.Search(tidbLogIndexPrefix, &buf, from, reqParams.PageSize)
	if err != nil {
		framework.Log().Errorf("search tidb log error: %v", err)
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}
	if resp.IsError() || resp.StatusCode != 200 {
		framework.Log().Errorf("search tidb failed! response [%v]", resp.String())
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal),
			fmt.Sprintf("search tidb failed! response: [%v]", resp.String())))
		return
	}
	var esResult ElasticSearchVo
	if err = json.NewDecoder(resp.Body).Decode(&esResult); err != nil {
		framework.Log().Fatalf("Error parsing the response body: %s", err)
		c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
		return
	}

	results := SearchTiDBLogRsp{
		Took: esResult.Took,
	}
	results.Results = make([]SearchTiDBLogDetail, 0)
	for _, hit := range esResult.Hits.Hits {
		var hitDetail SearchTiDBLogSourceDetail
		marshal, err := json.Marshal(hit.Source)
		if err != nil {
			framework.Log().Errorf("search tidb log error: %v", err)
			c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
			return
		}
		err = json.Unmarshal(marshal, &hitDetail)
		if err != nil {
			framework.Log().Errorf("search tidb log error: %v", err)
			c.JSON(http.StatusInternalServerError, controller.Fail(int(codes.Internal), err.Error()))
			return
		}

		results.Results = append(results.Results, SearchTiDBLogDetail{
			Index:      hit.Index,
			Id:         hit.Id,
			Level:      hitDetail.Log.Level,
			SourceLine: hitDetail.Log.Logger + ":" + hitDetail.Log.Origin.File.Line,
			Message:    hitDetail.Message,
			Ip:         hitDetail.Ip,
			ClusterId:  hitDetail.ClusterId,
			Module:     hitDetail.Fileset.Name,
			Ext:        hitDetail.Tidb,
			Timestamp:  time.Unix(hitDetail.Timestamp.Unix(), 0).Format(dateFormat),
		})
	}

	page := &controller.Page{
		Page:     reqParams.Page,
		PageSize: len(results.Results),
		Total:    esResult.Hits.Total.Value,
	}
	c.JSON(http.StatusOK, controller.BuildResultWithPage(0, "OK", page, results))
}

// buildSearchTiDBReqParams build request elasticsearch parameters
//response e.g.:
//{
//  "query": {
//    "bool": {
//      "filter": [
//        {
//          "term": {
//            "fileset.name": "pd"
//          }
//        },
//        {
//          "match": {
//            "message": "gc worker"
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
func buildSearchTiDBReqParams(clusterId string, reqParams SearchTiDBLogReq) (map[string]interface{}, error) {
	filters := make([]interface{}, 0)
	// required parameters: clusterId
	filters = append(filters, map[string]interface{}{
		"term": map[string]interface{}{
			"clusterId": clusterId,
		},
	})
	// option parameters: module
	if reqParams.Module != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"fileset.name": strings.ToLower(reqParams.Module),
			},
		})
	}
	// option parameters: level
	if reqParams.Level != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"log.level": strings.ToUpper(reqParams.Level),
			},
		})
	}
	// option parameters: ip
	if reqParams.Ip != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"ip": reqParams.Ip,
			},
		})
	}
	// option parameters: message
	if reqParams.Message != "" {
		filters = append(filters, map[string]interface{}{
			"match": map[string]interface{}{
				"message": reqParams.Message,
			},
		})
	}
	// option parameters: startTime, endTime
	tsFilter, err := filterTimestamp(reqParams)
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

// filterTimestamp search tidb log by @timestamp
func filterTimestamp(reqParams SearchTiDBLogReq) (map[string]interface{}, error) {
	tsFilter := map[string]interface{}{}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	var startTime time.Time
	if reqParams.StartTime != "" {
		startTime, err = time.ParseInLocation(dateFormat, reqParams.StartTime, loc)
		if err != nil {
			return nil, err
		}
	}
	var endTime time.Time
	if reqParams.EndTime != "" {
		endTime, err = time.ParseInLocation(dateFormat, reqParams.EndTime, loc)
		if err != nil {
			return nil, err
		}
	}
	if reqParams.StartTime != "" && reqParams.EndTime != "" {
		if startTime.After(endTime) {
			return nil, errors.New("illegal parameters, startTime after endTime")
		}
		tsFilter["gte"] = startTime.Unix() * 1000
		tsFilter["lte"] = endTime.Unix() * 1000
	} else if reqParams.StartTime != "" && reqParams.EndTime == "" {
		tsFilter["gte"] = startTime.Unix() * 1000
	} else if reqParams.StartTime == "" && reqParams.EndTime != "" {
		tsFilter["lte"] = endTime.Unix() * 1000
	}
	return tsFilter, nil
}

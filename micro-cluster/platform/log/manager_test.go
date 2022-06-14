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

/*******************************************************************************
 * @File: manager_test.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/17 14:47
*******************************************************************************/

package log

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/pingcap/tiunimanager/message"

	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/golang/mock/gomock"

	"github.com/alecthomas/assert"
	"github.com/pingcap/tiunimanager/common/structs"
)

func TestManager_prepareSearchParams_Success(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		buf, err := prepareSearchParams(context.TODO(), message.QueryPlatformLogReq{
			TraceId:   "123",
			Level:     "info",
			Message:   "hello",
			StartTime: 1630468800,
			EndTime:   1638331200,
			PageRequest: structs.PageRequest{
				Page:     1,
				PageSize: 10,
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, buf)
	})

	t.Run("success for end time", func(t *testing.T) {
		_, err := prepareSearchParams(context.TODO(), message.QueryPlatformLogReq{
			StartTime: 0,
			EndTime:   1638331200,
		})
		assert.NoError(t, err)
	})

	t.Run("success for start time", func(t *testing.T) {
		_, err := prepareSearchParams(context.TODO(), message.QueryPlatformLogReq{
			StartTime: 1630468800,
			EndTime:   0,
		})
		assert.NoError(t, err)
	})

	t.Run("error start time grate then end time", func(t *testing.T) {
		_, err := prepareSearchParams(context.TODO(), message.QueryPlatformLogReq{
			StartTime: 1638331200,
			EndTime:   1630468800,
		})
		assert.Error(t, err)
	})
}

var esResult = `
{
  "took" : 965,
  "timed_out" : false,
  "_shards" : {
    "total" : 75,
    "successful" : 75,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 1,
      "relation" : "eq"
    },
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "em-system-logs-2022.03.11",
        "_type" : "_doc",
        "_id" : "1n5reH8BvbYAsaKgyQFM",
        "_score" : 1.0,
        "_source" : {
          "@timestamp" : "2022-03-11T10:00:03.170Z",
          "type" : "logs",
          "input" : {
            "type" : "log"
          },
          "agent" : {
            "ephemeral_id" : "6e010327-7a29-4c9e-a81f-1da67eb889c5",
            "id" : "aea264c2-cfda-4bb5-90a3-76646da8b3f3",
            "name" : "CentOS76_VM",
            "type" : "filebeat",
            "version" : "8.0.0"
          },
          "ecs" : {
            "version" : "1.11.0"
          },
          "log" : {
            "offset" : 8166163,
            "file" : {
              "path" : "/root/tiunimanager/logs/cluster-server.log"
            }
          },
          "msg" : "WeekDay Friday, Hour: 18 need do auto backup for 0 clusters",
          "host" : {
            "name" : "CentOS76_VM"
          },
          "level" : "info",
          "time" : "2022-03-11T18:00:00+08:00"
        }
      }
    ]
  }
}
`

func TestManager_handleResult_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		resp, page, err := handleResult(context.TODO(), message.QueryPlatformLogReq{
			Level: "info",
		}, &esapi.Response{
			StatusCode: 200,
			Header:     nil,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(esResult))),
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.EqualValues(t, 1, page.Total)
	})

	t.Run("response error", func(t *testing.T) {
		_, _, err := handleResult(context.TODO(), message.QueryPlatformLogReq{
			Level: "info",
		}, &esapi.Response{
			StatusCode: 400,
			Header:     nil,
			Body:       nil,
		})
		assert.Error(t, err)
	})

	t.Run("json decode error", func(t *testing.T) {
		_, _, err := handleResult(context.TODO(), message.QueryPlatformLogReq{
			Level: "info",
		}, &esapi.Response{
			StatusCode: 200,
			Header:     nil,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`json decode error`))),
		})
		assert.Error(t, err)
	})
}

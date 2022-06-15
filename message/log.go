/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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
 * @File: log.go
 * @Description: system log
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/3/9 10:26
*******************************************************************************/

package message

import (
	"time"

	"github.com/pingcap/tiunimanager/common/structs"
)

// QueryPlatformLogReq Messages that query platform log information can be filtered based on query criteria
type QueryPlatformLogReq struct {
	TraceId   string `form:"traceId" example:"UNe7K1uERa-2fwSxGJ6CFQ"`
	Level     string `form:"level" example:"warn"`
	Message   string `form:"message" example:"some do something"`
	StartTime int64  `form:"startTime" example:"1630468800"`
	EndTime   int64  `form:"endTime" example:"1638331200"`
	structs.PageRequest
}

// QueryPlatformLogResp Reply message for querying platform log information
type QueryPlatformLogResp struct {
	Took    int               `json:"took" example:"10"`
	Results []PlatformLogItem `json:"results"`
}

// PlatformLogItem Query results item
type PlatformLogItem struct {
	Index       string `json:"index" example:"em-system-logs-2021.12.30"`
	Id          string `json:"id" example:"zvadfwf"`
	Level       string `json:"level" example:"warn"`
	TraceId     string `json:"traceId" example:"UNe7K1uERa-2fwSxGJ6CFQ"`
	MicroMethod string `json:"microMethod" example:"em.cluster.ClusterService.GetSystemInfo"`
	Message     string `json:"message"  example:"some do something"`
	Timestamp   string `json:"timestamp" example:"2021-09-23 14:23:10"`
}

type SearchPlatformLogSourceItem struct {
	Level       string    `json:"level"`
	TraceId     string    `json:"Em-X-Trace-Id"`
	MicroMethod string    `json:"micro-method"`
	Msg         string    `json:"msg"`
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"@timestamp"`
}

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
 * @File: common.go
 * @Description: cluster log common object
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/17 14:09
*******************************************************************************/

package log

import (
	"time"

	"github.com/pingcap-inc/tiunimanager/common/errors"
)

const (
	contextClusterMeta = "ClusterMeta"
)

const (
	DateFormat     = "2006-01-02 15:04:05"
	logIndexPrefix = "em-database-cluster-*"
)

type ElasticSearchResult struct {
	Took     int      `json:"took"`
	TimedOut bool     `json:"timed_out"`
	Hits     HitsInfo `json:"hits"`
}

type HitsInfo struct {
	Total    HitsTotal  `json:"total"`
	MaxScore float64    `json:"max_score"`
	Hits     []HitsItem `json:"hits"`
}

type HitsTotal struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

type HitsItem struct {
	Index  string                 `json:"_index"`
	Id     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

type SearchClusterLogSourceItem struct {
	Tidb      map[string]interface{} `json:"tidb"`
	Log       LogItem                `json:"log"`
	Ip        string                 `json:"ip"`
	ClusterId string                 `json:"clusterId"`
	Message   string                 `json:"message"`
	Fileset   FilesetItem            `json:"fileset"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"@timestamp"`
}

type LogItem struct {
	Offset int    `json:"offset"`
	Level  string `json:"level"`
	Logger string `json:"logger"`
}

type FilesetItem struct {
	Name string `json:"name"`
}

// CollectorClusterLogConfig
// @Description: collector tidb log config struct
type CollectorClusterLogConfig struct {
	Module  string                `json:"module" yaml:"module"`
	TiDB    CollectorModuleDetail `json:"tidb" yaml:"tidb"`
	PD      CollectorModuleDetail `json:"pd" yaml:"pd"`
	TiKV    CollectorModuleDetail `json:"tikv" yaml:"tikv"`
	TiFlash CollectorModuleDetail `json:"tiflash" yaml:"tiflash"`
	CDC     CollectorModuleDetail `json:"ticdc" yaml:"ticdc"`
}

type CollectorModuleDetail struct {
	Enabled bool                 `json:"enabled" yaml:"enabled"`
	Var     CollectorModuleVar   `json:"var" yaml:"var"`
	Input   CollectorModuleInput `json:"input" yaml:"input"`
}

type CollectorModuleVar struct {
	Paths []string `json:"paths" yaml:"paths"`
}

type CollectorModuleInput struct {
	Fields          CollectorModuleFields `json:"fields" yaml:"fields"`
	FieldsUnderRoot bool                  `json:"fields_under_root" yaml:"fields_under_root"`
	IncludeLines    []string              `json:"include_lines" yaml:"include_lines"`
	ExcludeLines    []string              `json:"exclude_lines" yaml:"exclude_lines"`
}

type CollectorModuleFields struct {
	Type      string `json:"type" yaml:"type"`
	ClusterId string `json:"clusterId" yaml:"clusterId"`
	Ip        string `json:"ip" yaml:"ip"`
}

// FilterTimestamp search tidb log by @timestamp
func FilterTimestamp(startTime, endTime int64) (map[string]interface{}, error) {
	tsFilter := map[string]interface{}{}

	if startTime > 0 && endTime > 0 {
		if startTime > endTime {
			return nil, errors.NewErrorf(errors.TIUNIMANAGER_LOG_TIME_AFTER, errors.TIUNIMANAGER_LOG_TIME_AFTER.Explain())
		}
		tsFilter["gte"] = startTime * 1000
		tsFilter["lte"] = endTime * 1000
	} else if startTime > 0 && endTime <= 0 {
		tsFilter["gte"] = startTime * 1000
	} else if startTime <= 0 && endTime > 0 {
		tsFilter["lte"] = endTime * 1000
	}
	return tsFilter, nil
}

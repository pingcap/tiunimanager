
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
	"time"
)

type ElasticSearchVo struct {
	Took     int      `json:"took"`
	TimedOut bool     `json:"timed_out"`
	Hits     HitsInfo `json:"hits"`
}

type HitsInfo struct {
	Total    HitsTotal    `json:"total"`
	MaxScore float64      `json:"max_score"`
	Hits     []HitsDetail `json:"hits"`
}

type HitsTotal struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

type HitsDetail struct {
	Index  string                 `json:"_index"`
	Id     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

type SearchTiDBLogSourceDetail struct {
	Tidb      map[string]interface{} `json:"tidb"`
	Log       LogDetail              `json:"log"`
	Ip        string                 `json:"ip"`
	ClusterId string                 `json:"clusterId"`
	Message   string                 `json:"message"`
	Fileset   FilesetDetail          `json:"fileset"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"@timestamp"`
}

type LogDetail struct {
	Offset int          `json:"offset"`
	Level  string       `json:"level"`
	Logger string       `json:"logger"`
	Origin OriginDetail `json:"origin"`
}

type OriginDetail struct {
	File FileDetail `json:"file"`
}

type FileDetail struct {
	Line string `json:"line"`
}

type FilesetDetail struct {
	Name string `json:"name"`
}

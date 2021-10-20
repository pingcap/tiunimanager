
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

type SearchTiDBLogRsp struct {
	Took    int                   `json:"took" example:"10"`
	Results []SearchTiDBLogDetail `json:"results"`
}

type SearchTiDBLogDetail struct {
	Index      string                 `json:"index" example:"tiem-tidb-cluster-2021.09.23"`
	Id         string                 `json:"id" example:"zvadfwf"`
	Level      string                 `json:"level" example:"warn"`
	SourceLine string                 `json:"sourceLine" example:"main.go:210"`
	Message    string                 `json:"message"  example:"tidb log"`
	Ip         string                 `json:"ip" example:"127.0.0.1"`
	ClusterId  string                 `json:"clusterId" example:"abc"`
	Module     string                 `json:"module" example:"tidb"`
	Ext        map[string]interface{} `json:"ext"`
	Timestamp  string                 `json:"timestamp" example:"2021-09-23 14:23:10"`
}

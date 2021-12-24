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
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package structs

type Usage struct {
	Total     float32 `json:"total"`
	Used      float32 `json:"used"`
	UsageRate float32 `json:"usageRate"`
}

// RegionInfo Information about the physical location of the cluster, the meaning of Region is the same as that of Cloud
type RegionInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ZoneInfo Information about the physical location of the cluster, the meaning of Zone is the same as that of Cloud
type ZoneInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ProductSpecInfo Enterprise Manager product specification description information, mainly will be described by computing, storage are specifications
type ProductSpecInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Page struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

// ResultWithoutPage Public structure for replying to http messages
type ResultWithoutPage struct {
	Code    int         `json:"code"`    //http error code
	Message string      `json:"message"` //http error msg
	Data    interface{} `json:"data"`    //The returned content, storage in JSON format
}

// ResultWithPage Public structure for replying to http messages
type ResultWithPage struct {
	Code    int         `json:"code"`    //http error code
	Message string      `json:"message"` //http error msg
	Page    Page        `json:"page"`    //Use when query results are returned multiple times
	Data    interface{} `json:"data"`    //The returned content, storage in JSON format
}

// PageRequest The paging information parameters passed in by the client when requesting the server
type PageRequest struct {
	Page     int `json:"page" form:"page"`         //Current page location
	PageSize int `json:"pageSize" form:"pageSize"` //Number of this request
}

func (p PageRequest) GetOffset() int {
	if p.Page == 0 {
		p.Page = 1
	}

	return (p.Page - 1) * p.PageSize
}

// Calculate the Page's offset in DB query sql
func (p *PageRequest) CalcOffset() (offset int) {
	if p.Page > 1 {
		offset = (p.Page - 1) * p.PageSize
	} else {
		offset = 0
	}
	return
}

// AsyncTaskWorkFlowInfo Public information returned by asynchronous tasks
type AsyncTaskWorkFlowInfo struct {
	WorkFlowID string `json:"workFlowId"` // Asynchronous task workflow ID
}

//
// Index common struct for statistical indicators
// @Description:
//
type Index struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
	Unit        string      `json:"unit"`
}

// Version Identifies the version number of the software
type Version struct {
	Version   string `json:"version"`
	GitHash   string `json:"gitHash"`
	GitBranch string `json:"gitBranch"`
	BuildTime string `json:"buildTime"`
}

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
 * @File: workflow.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package structs

import "time"

type WorkFlowInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	BizID      string    `json:"bizId"`
	BizType    string    `json:"bizType"`
	Status     string    `json:"status" enums:"Initializing,Processing,Finished,Error,Canceled"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
	DeleteTime time.Time `json:"deleteTime"`
}

type WorkFlowNodeInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Parameters string    `json:"parameters"`
	Result     string    `json:"result"`
	Status     string    `json:"status" enums:"Initializing,Processing,Finished,Error,Canceled"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
}

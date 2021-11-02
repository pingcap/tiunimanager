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

package flowtask

import "github.com/pingcap-inc/tiem/micro-api/controller"

type FlowWorkDisplayInfo struct {
	Id           uint   `json:"id"`
	FlowWorkName string `json:"flowWorkName"`
	ClusterId    string `json:"clusterId"`
	ClusterName  string `json:"clusterName"`
	controller.StatusInfo
	controller.Operator
}

type FlowWorkDetailInfo struct {
	FlowWorkDisplayInfo
	Tasks []FlowWorkTaskInfo `json:"tasks"`
}

type FlowWorkTaskInfo struct {
	Id         uint   `json:"id"`
	TaskName   string `json:"taskName"`
	Parameters string `json:"taskParameters"`
	Result     string `json:"result"`
	Status     int    `json:"taskStatus"`
}

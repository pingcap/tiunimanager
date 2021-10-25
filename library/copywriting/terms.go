
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

package copywriting

var DefaultLanguage = "cn"

func DisplayByDefault(key string) string {
	return CopyWritingContainer[DefaultLanguage][key]
}

func Display(key string, language string) string {
	return CopyWritingContainer[language][key]
}

// CopyWritingContainer map[language]map[copyWritingKey]copyWritingValue
var CopyWritingContainer = map[string]map[string]string{
	"cn":{
		CWClusterStatusUnlined: "未上线",
		CWClusterStatusOnline:  "运行中",
		CWClusterStatusOffline: "已下线",
		CWClusterStatusDeleted: "已删除",
		CWFlowCreateCluster:    "创建中",
		CWFlowDeleteCluster:    "删除中",
		CWFlowBackupCluster:    "备份中",
		CWFlowRecoverCluster:   "恢复中",
		CWFlowModifyParameters: "参数修改中",
		CWFlowExportData: 		"数据导出中",
		CWFlowImportData: 		"数据导出中",
		CWTaskStatusInit:       "未开始",
		CWTaskStatusProcessing: "进行中",
		CWTaskStatusFinished:   "完成",
		CWTaskStatusError:      "失败",
	},
}

var CWClusterStatusUnlined = "CW_ClusterStatusUnlined"
var CWClusterStatusOnline = "CW_ClusterStatusOnline"
var CWClusterStatusOffline = "CW_ClusterStatusOffline"
var CWClusterStatusDeleted = "CW_ClusterStatusDeleted"

var CWTaskStatusInit = "CW_TaskStatusInit"
var CWTaskStatusProcessing = "CW_TaskStatusProcessing"
var CWTaskStatusFinished = "CW_TaskStatusFinished"
var CWTaskStatusError = "CW_TaskStatusError"

var CWFlowCreateCluster = "CW_FlowCreateCluster"
var CWFlowTakeoverCluster = "CW_FlowTakeoverCluster"
var CWFlowDeleteCluster = "CW_FlowDeleteCluster"

var CWFlowBackupCluster = "CW_FlowBackupCluster"
var CWFlowRecoverCluster = "CW_FlowRecoverCluster"
var CWFlowModifyParameters = "CW_FlowModifyParameters"

var CWFlowExportData = "CW_FlowExportData"
var CWFlowImportData = "CW_FlowImportData"
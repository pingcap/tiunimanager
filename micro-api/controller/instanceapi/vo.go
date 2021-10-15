
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package instanceapi

import (
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"time"
)

type ParamItem struct {
	Definition   knowledge.Parameter 	`json:"definition"`
	CurrentValue ParamInstance	`json:"currentValue"`
}

type ParamInstance struct {
	Name 		string 			`json:"name"`
	Value  		interface{} 	`json:"value"`
}

type BackupRecord struct {
	ID 				int64		`json:"id"`
	ClusterId 		string		`json:"clusterId"`
	StartTime 		time.Time	`json:"startTime"`
	EndTime 		time.Time	`json:"endTime"`
	BackupType		string		`json:"backupType"`		// 全量/增量
	BackupMethod 	string		`json:"backupMethod"`	// 物理/逻辑
	BackupMode 		string 		`json:"backupMode"`		// 手动/自动
	Operator 		controller.Operator	`json:"operator"`
	Size 			uint64		`json:"size"`
	Status 			controller.StatusInfo	`json:"status"`
	FilePath 		string		`json:"filePath"`
}

/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
 * @File: datatransfer.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package structs

import "time"

type DataImportExportRecordInfo struct {
	RecordID      string    `json:"recordId"`
	ClusterID     string    `json:"clusterId"`
	TransportType string    `json:"transportType"`
	FilePath      string    `json:"filePath"`
	ZipName       string    `json:"zipName"`
	StorageType   string    `json:"storageType"`
	Comment       string    `json:"comment"`
	Status        string    `json:"status"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	CreateTime    time.Time `json:"createTime"`
	UpdateTime    time.Time `json:"updateTime"`
	DeleteTime    time.Time `json:"deleteTime"`
}

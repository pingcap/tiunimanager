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

package importexport

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"time"
)

type DataExportResp struct {
	RecordId int64 `json:"recordId"`
}

type DataImportResp struct {
	RecordId int64 `json:"recordId"`
}

type DataTransportInfo struct {
	RecordId      int64                 `json:"recordId"`
	ClusterId     string                `json:"clusterId"`
	TransportType string                `json:"transportType"`
	StartTime     time.Time             `json:"startTime"`
	EndTime       time.Time             `json:"endTime"`
	Status        controller.StatusInfo `json:"status"`
	FilePath      string                `json:"filePath"`
	StorageType   string                `json:"storageType"`
}

type DataTransportRecordQueryResp struct {
	TransportRecords []*DataTransportInfo `json:"transportRecords"`
}

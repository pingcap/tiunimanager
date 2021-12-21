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
 * @File: datatransfer.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package message

import (
	"github.com/pingcap-inc/tiem/common/structs"
)

type DataExportReq struct {
	ClusterID       string `json:"clusterId"`
	UserName        string `json:"userName"`
	Password        string `json:"password"`
	FileType        string `json:"fileType"`
	Filter          string `json:"filter"`
	Sql             string `json:"sql"`
	StorageType     string `json:"storageType"`
	ZipName         string `json:"zipName"`
	EndpointUrl     string `json:"endpointUrl"`
	BucketUrl       string `json:"bucketUrl"`
	AccessKey       string `json:"accessKey"`
	SecretAccessKey string `json:"secretAccessKey"`
	Comment         string `json:"comment"`
}

type DataExportResp struct {
	structs.AsyncTaskWorkFlowInfo
	RecordID string `json:"recordId"`
}

type DataImportReq struct {
	ClusterID       string `json:"clusterId"`
	UserName        string `json:"userName"`
	Password        string `json:"password"`
	RecordId        string `json:"recordId"`
	StorageType     string `json:"storageType"`
	EndpointUrl     string `json:"endpointUrl"`
	BucketUrl       string `json:"bucketUrl"`
	AccessKey       string `json:"accessKey"`
	SecretAccessKey string `json:"secretAccessKey"`
	Comment         string `json:"comment"`
}

type DataImportResp struct {
	RecordID string `json:"recordId"`
	structs.AsyncTaskWorkFlowInfo
}

type QueryDataImportExportRecordsReq struct {
	structs.PageRequest
	RecordID  string `json:"recordId" form:"recordId"`
	ClusterID string `json:"clusterId" form:"clusterId"`
	ReImport  bool   `json:"reImport" form:"reImport"`
	StartTime int64  `json:"startTime" form:"startTime"`
	EndTime   int64  `json:"endTime" form:"endTime"`
}

type QueryDataImportExportRecordsResp struct {
	Records []*structs.DataImportExportRecordInfo `json:"transportRecords"`
}

type DeleteImportExportRecordReq struct {
	RecordID string `json:"recordId" form:"recordId" swaggerignore:"true"`
}

type DeleteImportExportRecordResp struct {
	RecordID string `json:"recordId" form:"recordId"`
}

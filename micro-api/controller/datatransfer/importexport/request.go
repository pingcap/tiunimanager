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

import "github.com/pingcap-inc/tiem/micro-api/controller"

type DataExportReq struct {
	ClusterId       string `json:"clusterId"`
	UserName        string `json:"userName"`
	Password        string `json:"password"`
	FileType        string `json:"fileType"`
	Filter          string `json:"filter"`
	Sql             string `json:"sql"`
	StorageType     string `json:"storageType"`
	EndpointUrl     string `json:"endpointUrl"`
	BucketUrl       string `json:"bucketUrl"`
	BucketRegion    string `json:"bucketRegion"`
	AccessKey       string `json:"accessKey"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type DataImportReq struct {
	ClusterId       string `json:"clusterId"`
	UserName        string `json:"userName"`
	Password        string `json:"password"`
	StorageType     string `json:"storageType"`
	EndpointUrl     string `json:"endpointUrl"`
	BucketUrl       string `json:"bucketUrl"`
	AccessKey       string `json:"accessKey"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type DataTransportQueryReq struct {
	controller.PageRequest
	RecordId int64 `json:"recordId" form:"recordId"`
}

type DataTransportDeleteReq struct {
	ClusterId string `json:"clusterId"`
	RecordId  int64  `json:"recordId"`
}

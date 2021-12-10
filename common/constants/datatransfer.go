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

package constants

//Definition export & import data constants
type TransportType string

const (
	DefaultImportDir    string        = "/home/em/import"
	DefaultExportDir    string        = "/home/em/export"
	DefaultZipName      string        = "data.zip"
	TransportTypeExport TransportType = "export"
	TransportTypeImport TransportType = "import"
)

type DataExportStatus string

//Definition data export status information
const (
	DataExportInitializing DataExportStatus = "Initializing"
	DataExportProcessing   DataExportStatus = "Processing"
	DataExportFinished     DataExportStatus = "Finished"
	DataExportFailed       DataExportStatus = "Failed"
)

type DataImportStatus string

//Definition data import status information
const (
	DataImportInitializing DataImportStatus = "Initializing"
	DataImportProcessing   DataImportStatus = "Processing"
	DataImportFinished     DataImportStatus = "Finished"
	DataImportFailed       DataImportStatus = "Failed"
)

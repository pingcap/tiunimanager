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

package importexport

import (
	"context"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/message"
)

type ImportExportService interface {
	// ExportData
	// @Description: export data
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *message.DataExportResp
	// @Return error
	ExportData(ctx context.Context, request *message.DataExportReq) (*message.DataExportResp, error)

	// ImportData
	// @Description: import data
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *message.DataImportResp
	// @Return error
	ImportData(ctx context.Context, request *message.DataImportReq) (*message.DataImportResp, error)

	// QueryDataTransportRecords
	// @Description: query data import & export records by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *message.QueryDataImportExportRecordsResp
	// @Return *structs.Page
	// @Return error
	QueryDataTransportRecords(ctx context.Context, request *message.QueryDataImportExportRecordsReq) (*message.QueryDataImportExportRecordsResp, *structs.Page, error)

	// DeleteDataTransportRecord
	// @Description: delete data import & export records by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return *message.DeleteImportExportRecordResp
	// @Return error
	DeleteDataTransportRecord(ctx context.Context, request *message.DeleteImportExportRecordReq) (*message.DeleteImportExportRecordResp, error)
}

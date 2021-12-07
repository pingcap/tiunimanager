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

package models

import (
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
	changefeed2 "github.com/pingcap-inc/tiem/models/cluster/changefeed"
	"github.com/pingcap-inc/tiem/models/datatransfer/importexport"
	"github.com/pingcap-inc/tiem/models/workflow"
	"gorm.io/gorm"
)

var defaultDb database

type database struct {
	base                     *gorm.DB
	ChangeFeedReaderWriter   changefeed2.ReaderWriter
	WorkFlowReaderWriter     workflow.ReaderWriter
	ImportExportReaderWriter importexport.ReaderWriter
	BRReaderWriter           backuprestore.ReaderWriter
}

func open() {
	// todo
}

func initReaderWriter() {
	defaultDb.ChangeFeedReaderWriter = changefeed2.NewGormChangeFeedReadWrite(defaultDb.base)
	defaultDb.WorkFlowReaderWriter = workflow.NewFlowReadWrite(defaultDb.base)
	defaultDb.ImportExportReaderWriter = importexport.NewImportExportReadWrite(defaultDb.base)
	defaultDb.BRReaderWriter = backuprestore.NewBRReadWrite(defaultDb.base)
}

func addTable() {
	// todo
}

func GetChangeFeedReaderWriter() changefeed2.ReaderWriter {
	return defaultDb.ChangeFeedReaderWriter
}

func SetChangeFeedReaderWriter(rw changefeed2.ReaderWriter) {
	defaultDb.ChangeFeedReaderWriter = rw
}

func GetWorkFlowReaderWriter() workflow.ReaderWriter {
	return defaultDb.WorkFlowReaderWriter
}

func SetWorkFlowReaderWriter(rw workflow.ReaderWriter) {
	defaultDb.WorkFlowReaderWriter = rw
}

func GetImportExportReaderWriter() importexport.ReaderWriter {
	return defaultDb.ImportExportReaderWriter
}

func SetImportExportReaderWriter(rw importexport.ReaderWriter) {
	defaultDb.ImportExportReaderWriter = rw
}

func GetBRReaderWriter() backuprestore.ReaderWriter {
	return defaultDb.BRReaderWriter
}

func SetBRReaderWriter(rw backuprestore.ReaderWriter) {
	defaultDb.BRReaderWriter = rw
}

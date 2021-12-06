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

var DefaultDb database

type database struct {
	base                     *gorm.DB
	ChangeFeedReaderWriter   changefeed2.ReaderWriter
	WorkFlowReaderWriter     workflow.ReaderWriter
	ImportExportReaderWriter importexport.ReaderWriter
	BrReaderWriter           backuprestore.ReaderWriter
}

func open() {
	// todo
}

func initReaderWriter() {
	DefaultDb.ChangeFeedReaderWriter = changefeed2.NewGormChangeFeedReadWrite(DefaultDb.base)
	DefaultDb.WorkFlowReaderWriter = workflow.NewFlowReadWrite(DefaultDb.base)
	DefaultDb.ImportExportReaderWriter = importexport.NewImportExportReadWrite(DefaultDb.base)
	DefaultDb.BrReaderWriter = backuprestore.NewBRReadWrite(DefaultDb.base)
}

func addTable() {
	// todo
}

func GetChangeFeedReaderWriter() changefeed2.ReaderWriter {
	return DefaultDb.ChangeFeedReaderWriter
}

func GetWorkFlowReaderWriter() workflow.ReaderWriter {
	return DefaultDb.WorkFlowReaderWriter
}

func GetImportExportReaderWriter() importexport.ReaderWriter {
	return DefaultDb.ImportExportReaderWriter
}

func GetBrReaderWriter() backuprestore.ReaderWriter {
	return DefaultDb.BrReaderWriter
}

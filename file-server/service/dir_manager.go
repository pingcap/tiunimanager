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

package service

import (
	"fmt"
	"path/filepath"
)

//todo: get from config center
var importDir string = "/tmp/tiem/import"
var exportDir string = "/tmp/tiem/export"
var DirMgr DirManager

type DirManager struct {
	ImportDir string
	ExportDir string
}

func InitDirManager() *DirManager {
	importAbsDir, err := filepath.Abs(importDir)
	if err != nil {
		getLogger().Fatalf("import dir %s is not vaild", importDir)
		return nil
	}
	exportAbsDir, err := filepath.Abs(exportDir)
	if err != nil {
		getLogger().Fatalf("export dir %s is not vaild", exportDir)
		return nil
	}
	DirMgr = DirManager{
		ImportDir: importAbsDir,
		ExportDir: exportAbsDir,
	}
	return &DirMgr
}

func (mgr *DirManager) GetImportPath(clusterId string) string {
	return fmt.Sprintf("%s/%s", mgr.ImportDir, clusterId)
}

func (mgr *DirManager) GetExportPath(clusterId string) string {
	return fmt.Sprintf("%s/%s", mgr.ExportDir, clusterId)
}
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
	"fmt"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/handler"
)

// DataImportConfig data import toml config for lightning https://docs.pingcap.com/zh/tidb/dev/tidb-lightning-configuration
type DataImportConfig struct {
	Lightning    LightningCfg    `toml:"lightning"`
	TiKVImporter TiKVImporterCfg `toml:"tikv-importer"`
	MyDumper     MyDumperCfg     `toml:"mydumper"`
	TiDB         TiDBCfg         `toml:"tidb"`
}

type LightningCfg struct {
	Level             string `toml:"level"`              //lightning log level
	File              string `toml:"file"`               //lightning log path
	CheckRequirements bool   `toml:"check-requirements"` //lightning pre-check
}

// BackendLocal tidb-lightning backend https://docs.pingcap.com/zh/tidb/stable/tidb-lightning-backends#tidb-lightning-backend
const (
	BackendLocal string = "local"
	//BackendImport string = "importer"
	//BackendTiDB   string = "tidb"
)

type TiKVImporterCfg struct {
	Backend     string `toml:"backend"`       //backend mode
	SortedKvDir string `toml:"sorted-kv-dir"` //temp store path
}

type MyDumperCfg struct {
	DataSourceDir string `toml:"data-source-dir"` //import data filepath
}

type TiDBCfg struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	StatusPort int    `toml:"status-port"` //table information from tidb status port
	PDAddr     string `toml:"pd-addr"`
}

type ImportInfo struct {
	ClusterId   string
	UserName    string
	Password    string
	FilePath    string
	RecordId    string
	StorageType string
	ConfigPath  string
}

type ExportInfo struct {
	ClusterId   string
	UserName    string
	Password    string
	FileType    string
	RecordId    string
	FilePath    string
	Filter      string
	Sql         string
	StorageType string
}

func NewDataImportConfig(meta *handler.ClusterMeta, info *ImportInfo) *DataImportConfig {
	if meta == nil {
		return nil
	}
	//todo: get from meta
	tidbServerHost := ""
	pdServerHost := ""
	tidbServerPort := 0
	tidbStatusPort := 0
	pdClientPort := 0

	if tidbServerPort == 0 {
		tidbServerPort = common.DefaultTidbPort
	}
	if tidbStatusPort == 0 {
		tidbStatusPort = common.DefaultTidbStatusPort
	}
	if pdClientPort == 0 {
		pdClientPort = common.DefaultPDClientPort
	}

	/*
	 * todo: sorted-kv-dir and data-source-dir in the same disk, may slow down import performance,
	 *  and check-requirements = true can not pass lightning pre-check
	 *  in real environment, config data-source-dir = user nfs storage, sorted-kv-dir = other disk, turn on pre-check
	 */
	config := &DataImportConfig{
		Lightning: LightningCfg{
			Level:             "info",
			File:              fmt.Sprintf("%s/tidb-lightning.log", info.ConfigPath),
			CheckRequirements: false, //todo: TBD
		},
		TiKVImporter: TiKVImporterCfg{
			Backend:     BackendLocal,
			SortedKvDir: info.ConfigPath, //todo: TBD
		},
		MyDumper: MyDumperCfg{
			DataSourceDir: info.FilePath,
		},
		TiDB: TiDBCfg{
			Host:       tidbServerHost,
			Port:       tidbServerPort,
			User:       info.UserName,
			Password:   info.Password,
			StatusPort: tidbStatusPort,
			PDAddr:     fmt.Sprintf("%s:%d", pdServerHost, pdClientPort),
		},
	}
	return config
}

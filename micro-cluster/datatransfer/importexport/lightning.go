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

// DataImportConfig data import toml config for lightning https://docs.pingcap.com/zh/tidb/dev/tidb-lightning-configuration
type DataImportConfig struct {
	Lightning    LightningCfg    `toml:"lightning"`
	TikvImporter TikvImporterCfg `toml:"tikv-importer"`
	MyDumper     MyDumperCfg     `toml:"mydumper"`
	Tidb         TidbCfg         `toml:"tidb"`
}

type LightningCfg struct {
	Level             string `toml:"level"`              //lightning log level
	File              string `toml:"file"`               //lightning log path
	CheckRequirements bool   `toml:"check-requirements"` //lightning pre check
}

// BackendLocal tidb-lightning backend https://docs.pingcap.com/zh/tidb/stable/tidb-lightning-backends#tidb-lightning-backend
const (
	BackendLocal string = "local"
	//BackendImport string = "importer"
	//BackendTidb   string = "tidb"
)

type TikvImporterCfg struct {
	Backend     string `toml:"backend"`       //backend mode: local/normal
	SortedKvDir string `toml:"sorted-kv-dir"` //temp store path
}

type MyDumperCfg struct {
	DataSourceDir string `toml:"data-source-dir"` //import data filepath
}

type TidbCfg struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	StatusPort int    `toml:"status-port"` //table infomation from tidb status port
	PdAddr     string `toml:"pd-addr"`
}

func NewDataImportConfig() *DataImportConfig {
	//todo
	return &DataImportConfig{}
}

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
	"fmt"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
)

// DataImportConfig data import toml config for lightning https://docs.pingcap.com/zh/tidb/dev/tidb-lightning-configuration
type DataImportConfig struct {
	Lightning    LightningCfg    `toml:"lightning"`
	TiKVImporter TiKVImporterCfg `toml:"tikv-importer"`
	MyDumper     MyDumperCfg     `toml:"mydumper"`
	TiDB         TiDBCfg         `toml:"tidb"`
	CheckPoint   CheckPointCfg   `toml:"checkpoint"`
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

const (
	DriverFile  string = "file"
	DriverMysql string = "mysql"
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

type CheckPointCfg struct {
	Enable bool   `toml:"enable"`
	Driver string `toml:"driver"`
	Dsn    string `toml:"dsn"`
}

func NewDataImportConfig(ctx context.Context, meta *meta.ClusterMeta, info *importInfo) *DataImportConfig {
	if meta == nil || meta.Cluster == nil {
		framework.LogWithContext(ctx).Errorf("input meta is invalid!")
		return nil
	}
	tidbAddress := meta.GetClusterConnectAddresses()
	if len(tidbAddress) == 0 {
		framework.LogWithContext(ctx).Errorf("get tidb address from meta failed, empty address")
		return nil
	}
	framework.LogWithContext(ctx).Infof("get cluster %s tidb address from meta, %+v", meta.Cluster.ID, tidbAddress)
	pdAddress := meta.GetPDClientAddresses()
	if len(pdAddress) == 0 {
		framework.LogWithContext(ctx).Errorf("get pd address from meta failed, empty address")
		return nil
	}
	framework.LogWithContext(ctx).Infof("get cluster %s pd address from meta, %+v", meta.Cluster.ID, pdAddress)
	tidbStatusAddress := meta.GetClusterStatusAddress()
	if len(tidbStatusAddress) == 0 {
		framework.LogWithContext(ctx).Errorf("get tidb status address from meta failed, empty address")
		return nil
	}
	framework.LogWithContext(ctx).Infof("get cluster %s tidb status address from meta, %+v", meta.Cluster.ID, tidbStatusAddress)

	tidbServerHost := tidbAddress[0].IP
	tidbServerPort := tidbAddress[0].Port
	tidbStatusPort := tidbStatusAddress[0].Port
	pdServerHost := pdAddress[0].IP
	pdClientPort := pdAddress[0].Port

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
		CheckPoint: CheckPointCfg{
			Enable: true,
			Driver: DriverFile,
			Dsn:    fmt.Sprintf("%s:tidb_lightning_checkpoint.pb", info.ConfigPath),
		},
	}
	return config
}

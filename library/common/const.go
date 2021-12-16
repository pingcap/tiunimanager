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

package common

// micro service default port
const (
	DefaultMicroMetaDBPort  int = 4100
	DefaultMicroClusterPort int = 4110
	DefaultMicroApiPort     int = 4116
	DefaultMicroFilePort    int = 4118
	DefaultMetricsPort      int = 4121
)

// tidb component default port
const (
	DefaultTidbPort       int = 4000
	DefaultTidbStatusPort int = 10080
	DefaultPDClientPort   int = 2379
	DefaultAlertPort      int = 9093
	DefaultGrafanaPort    int = 3000
)

const (
	TiEM          string = "tiem"
	LogDirPrefix  string = "/logs/"
	CertDirPrefix string = "/cert/"
	DBDirPrefix   string = "/"

	SqliteFileName   string = "tiem.sqlite.db"
	DatabaseFileName string = "em.db"

	CrtFileName string = "server.crt"
	KeyFileName string = "server.key"

	LocalAddress string = "0.0.0.0"
)

const (
	LogFileSystem      = "system"
	LogFileSecondParty = "2nd"
	LogFileLibTiUP     = "libTiUP"

	LogFileAccess = "access"
	LogFileAudit  = "audit"
)

const (
	RegistryMicroServicePrefix = "/micro/registry/"
	HttpProtocol               = "http://"
)

var (
	TemplateFileName = "hostInfo_template.xlsx"
	TemplateFilePath = "./etc"
)

type TransportType string

const (
	DefaultImportDir    string        = "/tmp/tiem/import"
	DefaultExportDir    string        = "/tmp/tiem/export"
	DefaultZipName      string        = "data.zip"
	NfsStorageType      string        = "nfs"
	S3StorageType       string        = "s3"
	TransportTypeExport TransportType = "export"
	TransportTypeImport TransportType = "import"
)

const SlowSqlThreshold = 100

type ClusterRelationType uint32

const (
	SlaveTo ClusterRelationType = iota + 1
	StandBy
	CloneFrom
	RecoverFrom
)

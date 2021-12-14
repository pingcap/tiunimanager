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
 * @File: common.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package constants

const (
	Sunday    string = "Sunday"
	Monday    string = "Monday"
	Tuesday   string = "Tuesday"
	Wednesday string = "Wednesday"
	Thursday  string = "Thursday"
	Friday    string = "Friday"
	Saturday  string = "Saturday"
)

var WeekDayMap = map[string]int{
	Sunday:    0,
	Monday:    1,
	Tuesday:   2,
	Wednesday: 3,
	Thursday:  4,
	Friday:    5,
	Saturday:  6}

//System log-related constants
const (
	LogFileSystem      string = "system"
	LogFileSecondParty string = "2rd"
	LogFileLibTiUP     string = "libTiUP"
	LogFileAccess      string = "access"
	LogFileAudit       string = "audit"
	LogDirPrefix       string = "/logs/"
)

// Enterprise Manager database constants
const (
	DBDirPrefix    string = "/"
	SqliteFileName string = "em.db"
)

// Enterprise Manager Certificates constants
const (
	CertDirPrefix string = "/cert/"
	CertFileName  string = "server.crt"
	KeyFileName   string = "server.key"
)

// micro service default port
const (
	DefaultMicroMetaDBPort  int = 4100
	DefaultMicroClusterPort int = 4110
	DefaultMicroApiPort     int = 4116
	DefaultMicroFilePort    int = 4118
	DefaultMetricsPort      int = 4121
)

type EMProductIDType string

//Definition of product ID provided by Enterprise manager
const (
	EMProductIDTiDB              EMProductIDType = "TiDB"
	EMProductIDDataMigration     EMProductIDType = "DataMigration"
	EMProductIDEnterpriseManager EMProductIDType = "EnterpriseManager"
)

type EMProductComponentIDType string

//Definition of product component ID provided by Enterprise manager
const (
	ComponentIDTiDB    EMProductComponentIDType = "TiDB"
	ComponentIDTiKV    EMProductComponentIDType = "TiKV"
	ComponentIDPD      EMProductComponentIDType = "PD"
	ComponentIDTiFlash EMProductComponentIDType = "TiFlash"
	ComponentIDTiCDC   EMProductComponentIDType = "CDC"

	ComponentIDGrafana          EMProductComponentIDType = "Grafana"
	ComponentIDPrometheus       EMProductComponentIDType = "Prometheus"
	ComponentIDAlertManger      EMProductComponentIDType = "AlertManger"
	ComponentIDNodeExporter     EMProductComponentIDType = "NodeExporter"
	ComponentIDBlackboxExporter EMProductComponentIDType = "BlackboxExporter"

	ComponentIDClusterServer EMProductComponentIDType = "cluster-server"
	ComponentIDOpenAPIServer EMProductComponentIDType = "openapi-server"
	ComponentIDFileServer    EMProductComponentIDType = "file-server"
)

type EMProductComponentNameType string

//Consistent names and IDs for some components, and only define components that are inconsistent

const (
	ComponentNameTiDB        EMProductComponentIDType = "Compute Engine"
	ComponentNameTiKV        EMProductComponentIDType = "Storage Engine"
	ComponentNamePD          EMProductComponentIDType = "Schedule Engine"
	ComponentNameTiFlash     EMProductComponentIDType = "Column Storage Engine"
	ComponentNameGrafana     EMProductComponentIDType = "Monitor GUI"
	ComponentNamePrometheus  EMProductComponentIDType = "Monitor"
	ComponentNameAlertManger EMProductComponentIDType = "Alter GUI"
)

type EMInternalProduct int

//Definition of internal product, only Enterprise Manager' value is EMInternalProductYes
const (
	EMInternalProductNo  = 0
	EMInternalProductYes = 1
)

type ProductStatus string

//Definition product status information
const (
	ProductStatusOnline  ProductStatus = "Online"
	ProductStatusOffline ProductStatus = "Offline"
)

type ProductComponentStatus string

//Definition product component status information
const (
	ProductComponentStatusOnline  ProductComponentStatus = "Online"
	ProductComponentStatusOffline ProductComponentStatus = "Offline"
)

type ProductSpecStatus string

//Definition product spec status information
const (
	ProductSpecStatusOnline  ProductSpecStatus = "Online"
	ProductSpecStatusOffline ProductSpecStatus = "Offline"
)

type ProductUpgradePathStatus string

//Definition product spec status information
const (
	ProductUpgradePathAvailable   ProductUpgradePathStatus = "Available"
	ProductUpgradePathUnAvailable ProductUpgradePathStatus = "UnAvailable"
)

// MaxBatchQueryDataNumber Batch querying data from an array with maximum conditions
const (
	MaxBatchQueryDataNumber int = 512
)

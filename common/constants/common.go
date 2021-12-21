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

import (
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

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
	DBDirPrefix      string = "/"
	DatabaseFileName string = "em.db"
)

// Enterprise Manager Certificates constants
const (
	CertDirPrefix string = "/cert/"
	CertFileName  string = "server.crt"
	KeyFileName   string = "server.key"
)

// micro service default port
const (
	DefaultMicroMetaDBPort  int = 4099
	DefaultMicroClusterPort int = 4101
	DefaultMicroApiPort     int = 4100
	DefaultMicroFilePort    int = 4102
	DefaultMetricsPort      int = 4103
)

type EMProductIDType string

//Definition of product ID provided by Enterprise manager
const (
	EMProductIDTiDB              EMProductIDType = "TiDB"
	EMProductIDDataMigration     EMProductIDType = "DataMigration"
	EMProductIDEnterpriseManager EMProductIDType = "EnterpriseManager"
)

func ValidProductID(p string) error {
	if p == string(EMProductIDTiDB) || p == string(EMProductIDDataMigration) || p == string(EMProductIDEnterpriseManager) {
		return nil
	}
	return framework.NewTiEMErrorf(common.TIEM_RESOURCE_INVALID_PRODUCT_NAME, "valid product name: [%s|%s|%s]", string(EMProductIDTiDB), string(EMProductIDDataMigration), string(EMProductIDEnterpriseManager))
}

type EMProductComponentIDType string

var ParasiteComponentIDs = []EMProductComponentIDType{
	ComponentIDGrafana,
	ComponentIDPrometheus,
	ComponentIDAlertManger,
}

//Definition of product component ID provided by Enterprise manager
const (
	ComponentIDTiDB    EMProductComponentIDType = spec.ComponentTiDB
	ComponentIDTiKV    EMProductComponentIDType = spec.ComponentTiKV
	ComponentIDPD      EMProductComponentIDType = spec.ComponentPD
	ComponentIDTiFlash EMProductComponentIDType = spec.ComponentTiFlash
	ComponentIDTiCDC   EMProductComponentIDType = spec.ComponentCDC

	ComponentIDGrafana          EMProductComponentIDType = spec.ComponentGrafana
	ComponentIDPrometheus       EMProductComponentIDType = spec.ComponentPrometheus
	ComponentIDAlertManager     EMProductComponentIDType = spec.ComponentAlertmanager
	ComponentIDNodeExporter     EMProductComponentIDType = spec.ComponentNodeExporter
	ComponentIDBlackboxExporter EMProductComponentIDType = spec.ComponentBlackboxExporter

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

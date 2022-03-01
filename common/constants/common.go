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
	"time"

	"github.com/pingcap-inc/tiem/common/errors"
)

const (
	EM string = "em"
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
	SqliteFileName   string = "sqlite.db"
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
	return errors.NewErrorf(errors.TIEM_RESOURCE_INVALID_PRODUCT_NAME, "valid product name: [%s|%s|%s]", string(EMProductIDTiDB), string(EMProductIDDataMigration), string(EMProductIDEnterpriseManager))
}

type ProvidedVendor string

// Definition of vendors provided by Enterprise manager
const (
	Local ProvidedVendor = "Local"
	AWS   ProvidedVendor = "AWS"
)

func ValidProvidedVendor(p string) error {
	if p == string(Local) || p == string(AWS) {
		return nil
	}
	return errors.NewErrorf(errors.TIEM_RESOURCE_INVALID_VENDOR_NAME, "valid vendor name: [%s|%s]", string(Local), string(AWS))
}

type EMProductComponentIDType string

var ParasiteComponentIDs = []EMProductComponentIDType{
	ComponentIDGrafana,
	ComponentIDPrometheus,
	ComponentIDAlertManger,
}

//Definition of product component ID provided by Enterprise manager
const (
	ComponentIDTiDB    EMProductComponentIDType = "TiDB"
	ComponentIDTiKV    EMProductComponentIDType = "TiKV"
	ComponentIDPD      EMProductComponentIDType = "PD"
	ComponentIDTiFlash EMProductComponentIDType = "TiFlash"
	ComponentIDCDC     EMProductComponentIDType = "CDC"

	ComponentIDGrafana          EMProductComponentIDType = "Grafana"
	ComponentIDPrometheus       EMProductComponentIDType = "Prometheus"
	ComponentIDAlertManger      EMProductComponentIDType = "AlertManger"
	ComponentIDNodeExporter     EMProductComponentIDType = "NodeExporter"
	ComponentIDBlackboxExporter EMProductComponentIDType = "BlackboxExporter"

	ComponentIDClusterServer EMProductComponentIDType = "cluster-server"
	ComponentIDOpenAPIServer EMProductComponentIDType = "openapi-server"
	ComponentIDFileServer    EMProductComponentIDType = "file-server"
)

// SuggestedNodeCount
// @Description: get suggested node count
// @Receiver p
// @return []int32
func (p EMProductComponentIDType) SuggestedNodeCount() []int32 {
	switch p {
	case ComponentIDPD:
		return []int32{1, 3, 5, 7}
	default:
		return []int32{}
	}
}

func (p EMProductComponentIDType) SortWeight() int {
	switch p {
	case ComponentIDPD:
		return 19900
	case ComponentIDTiDB:
		return 19800
	case ComponentIDTiKV:
		return 19700
	case ComponentIDTiFlash:
		return 19600
	case ComponentIDCDC:
		return 19500
	case ComponentIDGrafana:
		return 19400
	case ComponentIDPrometheus:
		return 19300
	case ComponentIDAlertManger:
		return 19200
	case ComponentIDNodeExporter:
		return 18900
	case ComponentIDBlackboxExporter:
		return 18800
	case ComponentIDClusterServer:
		return 9900
	case ComponentIDOpenAPIServer:
		return 9800
	case ComponentIDFileServer:
		return 9700
	default:
		return 0
	}
}

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

const DefaultTokenValidPeriod time.Duration = 4 * time.Hour

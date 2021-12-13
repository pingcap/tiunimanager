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

type EMProductNameType string

//Definition of product names provided by Enterprise manager
const (
	EMProductNameTiDB              EMProductNameType = "TiDB"
	EMProductNameDataMigration     EMProductNameType = "DataMigration"
	EMProductNameEnterpriseManager EMProductNameType = "EnterpriseManager"
)

type ProductStatus string

//Definition product status information
const (
	ProductStatusOnline    ProductStatus = "Online"
	ProductStatusOffline   ProductStatus = "Offline"
	ProductStatusException ProductStatus = "Exception" // only TiDB Enterprise Manager
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

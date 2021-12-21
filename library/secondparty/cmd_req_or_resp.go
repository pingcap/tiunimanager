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

package secondparty

import (
	"github.com/pingcap-inc/tiem/library/spec"
	"github.com/pingcap-inc/tiem/models/workflow/secondparty"
	spec2 "github.com/pingcap/tiup/pkg/cluster/spec"
)

type CmdDeployReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	InstanceName  string
	Version       string
	ConfigStrYaml string
	TimeoutS      int
	TiUPPath      string
	Flags         []string
}

type CmdScaleOutReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	InstanceName  string
	ConfigStrYaml string
	TimeoutS      int
	TiUPPath      string
	Flags         []string
}

type CmdScaleInReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	InstanceName  string
	NodeId        string
	TimeoutS      int
	TiUPPath      string
	Flags         []string
}

type CmdStartReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	InstanceName  string
	TimeoutS      int
	TiUPPath      string
	Flags         []string
}

type CmdListReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	TimeoutS      int
	TiUPPath      string
	Flags         []string
}

type CmdDestroyReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	InstanceName  string
	TimeoutS      int
	TiUPPath      string
	Flags         []string
}

type CmdListResp struct {
	ListRespStr string
}

type CmdGetAllTaskStatusResp struct {
	Stats []TaskStatusMember
}

type CmdGetAllOperationStatusResp struct {
	Stats []OperationStatusMember
}

type CmdDumplingReq struct {
	TaskID   uint64
	TimeoutS int
	TiUPPath string
	Flags    []string
}

type CmdLightningReq struct {
	TaskID   uint64
	TimeoutS int
	TiUPPath string
	Flags    []string
}

type CmdDisplayReq struct {
	TiUPComponent TiUPComponentTypeStr
	InstanceName  string
	TimeoutS      int
	TiUPPath      string
	Flags         []string
}

type CmdDisplayResp struct {
	DisplayRespString string
}

type CmdBackUpReq struct {
	TaskID            uint64
	DbName            string
	TableName         string
	FilterDbTableName string // used in br command, pending for use in SQL command
	StorageAddress    string
	DbConnParameter   DbConnParam // only for SQL command, not used in br command
	RateLimitM        string
	Concurrency       string   // only for SQL command, not used in br command
	CheckSum          string   // only for SQL command, not used in br command
	LogFile           string   // used in br command, pending for use in SQL command
	TimeoutS          int      // used in br command, pending for use in SQL command
	Flags             []string // used in br command, pending for use in SQL command
}

type CmdBrResp struct {
	Destination   string
	Size          uint64
	BackupTS      uint64
	QueueTime     string
	ExecutionTime string
}

type CmdShowBackUpInfoReq struct {
	TaskID          uint64
	DbConnParameter DbConnParam
}

type CmdShowBackUpInfoResp struct {
	Destination   string
	Size          uint64
	BackupTS      uint64
	State         string
	Progress      float32
	QueueTime     string
	ExecutionTime string
	FinishTime    *string
	Connection    string
	ErrorStr      string
}

type CmdRestoreReq struct {
	TaskID            uint64
	DbName            string
	TableName         string
	FilterDbTableName string // used in br command, pending for use in SQL command
	StorageAddress    string
	DbConnParameter   DbConnParam // only for SQL command, not used in br command
	RateLimitM        string
	Concurrency       string   // only for SQL command, not used in br command
	CheckSum          string   // only for SQL command, not used in br command
	LogFile           string   // used in br command, pending for use in SQL command
	TimeoutS          int      // used in br command, pending for use in SQL command
	Flags             []string // used in br command, pending for use in SQL command
}

type CmdShowRestoreInfoReq struct {
	TaskID          uint64
	DbConnParameter DbConnParam
}

type CmdShowRestoreInfoResp struct {
	Destination   string
	Size          uint64
	BackupTS      uint64
	State         string
	Progress      float32
	QueueTime     string
	ExecutionTime string
	FinishTime    *string
	Connection    string
	ErrorStr      string
}

type CmdTransferReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	InstanceName  string
	CollectorYaml string
	RemotePath    string
	TimeoutS      int
	TiUPPath      string
	Flags         []string
}

type CmdUpgradeReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	InstanceName  string
	Version       string
	TimeoutS      int
	TiUPPath      string
	Flags         []string
}

type CmdShowConfigReq struct {
	TiUPComponent TiUPComponentTypeStr
	InstanceName  string
	TimeoutS      int
	Flags         []string
}

type CmdShowConfigResp struct {
	TiDBClusterTopo *spec2.Specification
}

type GlobalComponentConfig struct {
	TiDBClusterComponent spec.TiDBClusterComponent
	ConfigMap            map[string]interface{}
}

type ClusterComponentConfig struct {
	TiDBClusterComponent spec.TiDBClusterComponent
	InstanceAddr         string
	ConfigKey            string
	ConfigValue          string
}

type CmdEditGlobalConfigReq struct {
	TiUPComponent          TiUPComponentTypeStr
	InstanceName           string
	GlobalComponentConfigs []GlobalComponentConfig
	TimeoutS               int
	Flags                  []string
}

type CmdEditInstanceConfigReq struct {
	TiUPComponent        TiUPComponentTypeStr
	InstanceName         string
	TiDBClusterComponent spec.TiDBClusterComponent
	Host                 string
	Port                 int
	ConfigMap            map[string]interface{}
	TimeoutS             int
	Flags                []string
}

type CmdEditConfigReq struct {
	TiUPComponent TiUPComponentTypeStr
	InstanceName  string
	NewTopo       *spec2.Specification
	TimeoutS      int
	Flags         []string
}

type CmdReloadConfigReq struct {
	TiUPComponent TiUPComponentTypeStr
	InstanceName  string
	TimeoutS      int
	Flags         []string
}

type CmdClusterExecReq struct {
	TiUPComponent TiUPComponentTypeStr
	InstanceName  string
	TimeoutS      int
	Flags         []string
}

type ApiEditConfigReq struct {
	TiDBClusterComponent spec.TiDBClusterComponent
	InstanceHost         string
	InstancePort         uint
	Headers              map[string]string
	ConfigMap            map[string]interface{}
}

type ClusterEditConfigReq struct {
	DbConnParameter  DbConnParam
	ComponentConfigs []ClusterComponentConfig
}

type ClusterEditConfigResp struct {
	Message string
}

type ShowWarningsResp struct {
	Level   string
	Code    string
	Message string
}

type GetOperationStatusResp struct {
	Status   secondparty.OperationStatus
	Result   string
	ErrorStr string
}

type ClusterSetDbPswReq struct {
	DbConnParameter DbConnParam
}

type ClusterSetDbPswResp struct {
	Message string
}

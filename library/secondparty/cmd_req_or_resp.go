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

type CmdDeployReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	InstanceName  string
	Version       string
	ConfigStrYaml string
	TimeoutS      int
	TiupPath      string
	Flags         []string
}

type CmdStartReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	InstanceName  string
	TimeoutS      int
	TiupPath      string
	Flags         []string
}

type CmdListReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	TimeoutS      int
	TiupPath      string
	Flags         []string
}

type CmdDestroyReq struct {
	TiUPComponent TiUPComponentTypeStr
	TaskID        uint64
	InstanceName  string
	TimeoutS      int
	TiupPath      string
	Flags         []string
}

type CmdListResp struct {
	ListRespStr string
}

type CmdGetAllTaskStatusResp struct {
	Stats []TaskStatusMember
}

type CmdDumplingReq struct {
	TaskID   uint64
	TimeoutS int
	TiupPath string
	Flags    []string
}

type CmdLightningReq struct {
	TaskID   uint64
	TimeoutS int
	TiupPath string
	Flags    []string
}

type CmdDisplayReq struct {
	TiUPComponent TiUPComponentTypeStr
	InstanceName  string
	TimeoutS      int
	TiupPath      string
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
	Destination    string
	Size           uint64
	BackupTS       uint64
	Queue_time     string
	Execution_Time string
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
	TiupPath      string
	Flags         []string
}

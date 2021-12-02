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

package workflow

import (
	"github.com/pingcap-inc/tiem/library/copywriting"
)

type TaskStatus int

const (
	TaskStatusInit       TaskStatus = 0
	TaskStatusProcessing TaskStatus = 1
	TaskStatusFinished   TaskStatus = 2
	TaskStatusError      TaskStatus = 3
	TaskStatusCanceled   TaskStatus = 4
)

func (s TaskStatus) Display() string {

	switch s {
	case TaskStatusInit:
		return copywriting.DisplayByDefault(copywriting.CWTaskStatusInit)
	case TaskStatusProcessing:
		return copywriting.DisplayByDefault(copywriting.CWTaskStatusProcessing)
	case TaskStatusFinished:
		return copywriting.DisplayByDefault(copywriting.CWTaskStatusFinished)
	case TaskStatusError:
		return copywriting.DisplayByDefault(copywriting.CWTaskStatusError)
	case TaskStatusCanceled:
		return copywriting.DisplayByDefault(copywriting.CWTaskStatusError)
	}

	panic("Unknown task status")
}

func (s TaskStatus) Finished() bool {
	return TaskStatusFinished == s || TaskStatusError == s || TaskStatusCanceled == s
}

var allTaskStatus = []TaskStatus{
	TaskStatusInit,
	TaskStatusProcessing,
	TaskStatusFinished,
	TaskStatusError,
	TaskStatusCanceled,
}

type TaskReturnType string

const (
	SyncFuncTask TaskReturnType = "SyncFuncTask"
	PollingTasK  TaskReturnType = "PollingTasK"
)

const (
	FlowCreateCluster    = "CreateCluster"
	FlowDeleteCluster    = "DeleteCluster"
	FlowBackupCluster    = "BackupCluster"
	FlowRecoverCluster   = "RecoverCluster"
	FlowModifyParameters = "ModifyParameters"
	FlowExportData       = "ExportData"
	FlowImportData       = "ImportData"
	FlowRestartCluster   = "RestartCluster"
	FlowStopCluster      = "StopCluster"
	FlowTakeoverCluster  = "TakeoverCluster"
	FlowBuildLogConfig   = "BuildLogConfig"
	FlowScaleOutCluster  = "ScaleOutCluster"
	FlowScaleInCluster   = "ScaleInCluster"
)

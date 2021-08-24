package domain

import (
	copywriting2 "github.com/pingcap-inc/tiem/library/copywriting"
)

type ClusterStatus int

const (
	ClusterStatusUnlined ClusterStatus = 0
	ClusterStatusOnline  ClusterStatus = 1
	ClusterStatusOffline ClusterStatus = 2
	ClusterStatusDeleted ClusterStatus = 3
)

var allClusterStatus = []ClusterStatus{
	ClusterStatusUnlined,
	ClusterStatusOnline,
	ClusterStatusOffline,
	ClusterStatusDeleted,
}

func ClusterStatusFromValue(v int) ClusterStatus {
	for _, s := range allClusterStatus {
		if int(s) == v {
			return s
		}
	}
	return -1
}

// Display todo
func (s ClusterStatus) Display() string {

	switch s {
	case ClusterStatusUnlined: return copywriting2.DisplayByDefault(copywriting2.CWClusterStatusUnlined)
	case ClusterStatusOnline: return copywriting2.DisplayByDefault(copywriting2.CWClusterStatusOnline)
	case ClusterStatusOffline: return copywriting2.DisplayByDefault(copywriting2.CWClusterStatusOffline)
	case ClusterStatusDeleted: return copywriting2.DisplayByDefault(copywriting2.CWClusterStatusDeleted)
	}

	panic("Unknown cluster status")
}

type TaskStatus int

const (
	TaskStatusInit       TaskStatus = 0
	TaskStatusProcessing TaskStatus = 1
	TaskStatusFinished   TaskStatus = 2
	TaskStatusError      TaskStatus = 3
)

func (s TaskStatus) Display() string {

	switch s {
	case TaskStatusInit: return copywriting2.DisplayByDefault(copywriting2.CWTaskStatusInit)
	case TaskStatusProcessing: return copywriting2.DisplayByDefault(copywriting2.CWTaskStatusProcessing)
	case TaskStatusFinished: return copywriting2.DisplayByDefault(copywriting2.CWTaskStatusFinished)
	case TaskStatusError: return copywriting2.DisplayByDefault(copywriting2.CWTaskStatusError)
	}

	panic("Unknown task status")
}

func (s TaskStatus) Finished() bool {
	return TaskStatusFinished == s || TaskStatusError == s
}

var allTaskStatus = []TaskStatus{
	TaskStatusInit,
	TaskStatusProcessing,
	TaskStatusFinished,
	TaskStatusError,
}

func TaskStatusFromValue(v int) TaskStatus {
	for _, s := range allTaskStatus {
		if int(s) == v {
			return s
		}
	}
	return -1
}

type CronStatus int

const (
	CronStatusValid 		CronStatus 	= 0
	CronStatusInvalid 		CronStatus 	= 1
	CronStatusDeleted 		CronStatus 	= 2
)

type TaskReturnType int8

const (
	UserTask     TaskReturnType = 1
	SyncFuncTask TaskReturnType = 2
	CallbackTask TaskReturnType = 3
	PollingTasK  TaskReturnType = 4
)

const (
	FlowCreateCluster = "CreateCluster"
	FlowDeleteCluster = "DeleteCluster"
	FlowBackupCluster = "BackupCluster"
	FlowRecoverCluster = "RecoverCluster"
	FlowModifyParameters = "ModifyParameters"
	FlowExportData = "ExportData"
	FlowImportData = "ImportData"
)

type CronTaskType int8

const (
	CronMaintainStart CronTaskType = 1
	CronMaintainEnd   CronTaskType = 2
	CronBackup 		  CronTaskType = 3
)


type BackupRange 		int
type BackupType int

const (
	BackupRangeWhole BackupRange = 0
	BackupRangeIncrement = 1
)

const (
	BackupTypeLogic BackupType = 0
	BackupTypePhysics = 1
)
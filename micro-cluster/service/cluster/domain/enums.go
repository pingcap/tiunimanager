package domain

import "github.com/pingcap/ticp/copywriting"

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
	case ClusterStatusUnlined: return copywriting.DisplayByDefault(copywriting.CWClusterStatusUnlined)
	case ClusterStatusOnline: return copywriting.DisplayByDefault(copywriting.CWClusterStatusOnline)
	case ClusterStatusOffline: return copywriting.DisplayByDefault(copywriting.CWClusterStatusOffline)
	case ClusterStatusDeleted: return copywriting.DisplayByDefault(copywriting.CWClusterStatusDeleted)
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
)

type CronTaskType int8

const (
	CronMaintainStart CronTaskType = 1
	CronMaintainEnd   CronTaskType = 2
	CronBackup 		  CronTaskType = 3
)
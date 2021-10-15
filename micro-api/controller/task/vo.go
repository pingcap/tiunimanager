package taskapi

import "github.com/pingcap-inc/tiem/micro-api/controller"

type FlowWorkDisplayInfo struct {
	Id 					uint		`json:"id"`
	FlowWorkName 		string   	`json:"flowWorkName"`
	ClusterId			string 		`json:"clusterId"`
	ClusterName			string 		`json:"clusterName"`
	controller.StatusInfo
	controller.Operator
}

type FlowWorkDetailInfo struct {
	FlowWorkDisplayInfo
	Tasks 				[]FlowWorkTaskInfo 	`json:"tasks"`
}

type FlowWorkTaskInfo struct {
	Id 					uint		`json:"id"`
	TaskName 			string   	`json:"taskName"`
}
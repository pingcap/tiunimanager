package databaseapi

import "time"

type DataTransportResp struct {
	FlowId        string     `json:"flowId"`//返回flow流程的id
}

type TransportType int32
const (
	TransportTypeExport TransportType = 0
	TransportTypeImport TransportType = 1
)

type DataTransportTaskInfo struct {
	FlowId        string           `json:"flowId"`
	ClusterId     string           `json:"clusterId"`
	TransportType TransportType    `json:"transportType"`
	CreatedTime   time.Time        `json:"createdTime"`
	Result        bool             `json:"result"`
	DataDir       string           `json:"dataDir"`
}

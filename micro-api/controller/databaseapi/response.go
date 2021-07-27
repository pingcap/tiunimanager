package databaseapi

import "time"

type DataExportResp struct {
	FlowId        uint32     `json:"flowId"`//返回flow流程的id
}

type DataImportResp struct {
	FlowId        uint32     `json:"flowId"`//返回flow流程的id
}

type TransportType uint32
const (
	TransportTypeExport TransportType = 0
	TransportTypeImport TransportType = 1
)

type DataTransportTaskInfo struct {
	FlowId        uint32           `json:"flowId"`
	ClusterId     string           `json:"clusterId"`
	TransportType TransportType    `json:"transportType"`
	CreatedTime   time.Time        `json:"createdTime"`
	Result        bool             `json:"result"`
	DataDir       string           `json:"dataDir"`
}

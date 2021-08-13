package instanceapi

import (
	"github.com/pingcap/tiem/micro-api/controller"
)

type ParamQueryReq struct {
	controller.PageRequest
}

type ParamUpdateReq struct {
	Values []ParamInstance `json:"values"`
}

type BackupRecordQueryReq struct {
	//StartTime time.Time
	//EndTime time.Time
	controller.PageRequest
}

type BackupStrategyUpdateReq struct {
	BackupStrategy
}

type BackupReq struct {
	ClusterId string `json:"clusterId"`
}
type BackupRecoverReq struct {
	ClusterId string `json:"clusterId"`
}

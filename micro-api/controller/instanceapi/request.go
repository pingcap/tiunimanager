package instanceapi

import (
	"github.com/pingcap/tiem/micro-api/controller"
)

type ParamQueryReq struct {
	controller.PageRequest
}

type ParamUpdateReq struct {
	ClusterId 		string				`json:"clusterId"`
	Values			[]ParamInstance		`json:"values"`
}

type BackupRecordQueryReq struct {
	//StartTime time.Time
	//EndTime time.Time
	controller.PageRequest
}

type BackupStrategyUpdateReq struct {
	ClusterId string	`json:"clusterId"`
	BackupStrategy
}

type BackupRecoverReq struct {
	ClusterId 			string	`json:"clusterId"`
	BackupRecordId	 	int64	`json:"backupRecordId"`
}

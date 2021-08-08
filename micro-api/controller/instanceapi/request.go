package instanceapi

import (
	"github.com/pingcap/tiem/micro-api/controller"
)

type ParamQueryReq struct {
	controller.PageRequest
}

type ParamUpdateReq struct {
	ClusterId 		string
	Values			[]ParamInstance
}

type BackupRecordQueryReq struct {
	//StartTime time.Time
	//EndTime time.Time
	controller.PageRequest
}

type BackupStrategyUpdateReq struct {
	ClusterId string
	BackupStrategy
}

type BackupRecoverReq struct {
	ClusterId 			string
	BackupRecordId	 	int64
}

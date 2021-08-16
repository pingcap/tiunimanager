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

type BackupStrategy struct {
	ClusterId 	string 		`json:"clusterId"`
	BackupDate	string		`json:"backupDate"`
	FilePath	string 		`json:"filePath"`
	BackupRange string		`json:"backupRange"`
	BackupType 	string		`json:"backupType"`
	Period		string 		`json:"period"`
}

type BackupStrategyUpdateReq struct {
	strategy BackupStrategy		`json:"strategy"`
}

type BackupReq struct {
	ClusterId string `json:"clusterId"`
}
type BackupRecoverReq struct {
	ClusterId string `json:"clusterId"`
}

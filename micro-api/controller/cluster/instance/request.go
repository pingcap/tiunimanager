package instanceapi

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"time"
)

type ParamQueryReq struct {
	controller.PageRequest
}

type ParamUpdateReq struct {
	Values []ParamInstance `json:"values"`
}

type BackupRecordQueryReq struct {
	controller.PageRequest
	ClusterId string `json:"clusterId" form:"clusterId"`
	StartTime int64  `json:"startTime" form:"startTime"`
	EndTime   int64  `json:"endTime" form:"endTime"`
}

type BackupDeleteReq struct {
	ClusterId string `json:"clusterId"`
}

type BackupStrategy struct {
	ClusterId      string    `json:"clusterId"`
	BackupDate     string    `json:"backupDate"`
	Period         string    `json:"period"`
	NextBackupTime time.Time `json:"nextBackupTime"`
}

type BackupStrategyUpdateReq struct {
	Strategy BackupStrategy `json:"strategy"`
}

type BackupReq struct {
	ClusterId   	string `json:"clusterId"`
	BackupType  	string `json:"backupType"`
	BackupMethod 	string `json:"backupMethod"`
	FilePath    	string `json:"filePath"`
}
type BackupRecoverReq struct {
	ClusterId string `json:"clusterId"`
}

package instanceapi

import (
	"github.com/pingcap/ticp/knowledge/models"
	"github.com/pingcap/ticp/micro-api/controller"
	"time"
)

type ParamItem struct {
	Definition   models.Parameter
	CurrentValue ParamInstance
}

type ParamInstance struct {
	Name 		string
	Value  		interface{}
}

type BackupRecord struct {
	ID 				string
	ClusterId 		string
	StartTime 		time.Time
	EndTime 		time.Time
	Range 			BackupRange
	Way 			BackupWay
	Operator 		controller.Operator
	Size 			float32
	Status 			controller.StatusInfo
	FilePath 		string
}

type BackupRange 		int
type BackupWay 			int

type BackupStrategy struct {
	CronString 			string
}

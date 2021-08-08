package instanceapi

import (
	"github.com/pingcap/tiem/library/knowledge"
	"github.com/pingcap/tiem/micro-api/controller"
	"time"
)

type ParamItem struct {
	Definition   knowledge.Parameter
	CurrentValue ParamInstance
}

type ParamInstance struct {
	Name 		string 			`json:"name"`
	Value  		interface{} 	`json:"value"`
}

type BackupRecord struct {
	ID 				int64
	ClusterId 		string
	StartTime 		time.Time
	EndTime 		time.Time
	Range 			int
	Way 			int
	Operator 		controller.Operator
	Size 			float32
	Status 			controller.StatusInfo
	FilePath 		string
}

type BackupStrategy struct {
	CronString 			string
}

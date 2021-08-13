package instanceapi

import (
	"github.com/pingcap/tiem/library/knowledge"
	"github.com/pingcap/tiem/micro-api/controller"
	"time"
)

type ParamItem struct {
	Definition   knowledge.Parameter 	`json:"definition"`
	CurrentValue ParamInstance	`json:"currentValue"`
}

type ParamInstance struct {
	Name 		string 			`json:"name"`
	Value  		interface{} 	`json:"value"`
}

type BackupRecord struct {
	ID 				int64	`json:"id"`
	ClusterId 		string	`json:"clusterId"`
	StartTime 		time.Time	`json:"startTime"`
	EndTime 		time.Time	`json:"endTime"`
	Range 			int	`json:"range"`
	Way 			int	`json:"way"`
	Operator 		controller.Operator	`json:"operator"`
	Size 			float32	`json:"size"`
	Status 			controller.StatusInfo	`json:"status"`
	FilePath 		string	`json:"filePath"`
}

type BackupStrategy struct {
	CronString 			string	`json:"cronString"`
}

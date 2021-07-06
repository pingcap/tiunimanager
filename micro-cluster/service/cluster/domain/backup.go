package domain

import (
	"github.com/pingcap/ticp/micro-api/controller"
	"time"
)

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
	ValidityPeriod  	int64
	CronString 			string
}

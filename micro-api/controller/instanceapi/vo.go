package instanceapi

import (
	"github.com/pingcap/ticp/micro-api/controller"
	"time"
)

type ParamItem struct {
	Definition   Parameter
	CurrentValue ParamValue
}

type Parameter struct {
	Name			string
	Type 			ParamType
	Unit  			ParamUnit
	DefaultValue 	interface{}
	NeedRestart 	bool
	Constraints		[]ParamValueConstraint
	Desc 			string
}

type ParamValueConstraint struct {
	Type	ConstraintType
	ContrastValue	interface{}
}

type ParamValue struct {
	Name 		string
	Value  		interface{}
}

type ConstraintType string

const (
	ConstraintTypeLT = "lt"
	ConstraintTypeLTE = "gte"

	ConstraintTypeGT = "gt"
	ConstraintTypeGTE = "gte"

	ConstraintTypeNE = "ne"
	ConstraintTypeONEOF = "oneOf"

	ConstraintTypeMIN = "min"
	ConstraintTypeMAX = "max"
)

type ParamType int

const (
	ParamTypeInteger ParamType = 1
	ParamTypeFloat ParamType = 2
	ParamTypeString ParamType = 3
	ParamTypeIntegers ParamType = 4
	ParamTypeFloats ParamType = 5
	ParamTypeStrings ParamType = 6
	ParamTypeBoolean ParamType = 7
)

type ParamUnit string

const (
	ParamUnitNil ParamUnit = ""
	ParamUnitGb ParamUnit = "gb"
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

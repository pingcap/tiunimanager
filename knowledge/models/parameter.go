package models

type Parameter struct {
	Name         string
	Type         ParamType
	Unit         ParamUnit
	DefaultValue interface{}
	NeedRestart  bool
	Constraints  []ParamValueConstraint
	Desc         string
}

type ParamValueConstraint struct {
	Type          ConstraintType
	ContrastValue interface{}
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
	ParamTypeInteger  ParamType = 1
	ParamTypeFloat    ParamType = 2
	ParamTypeString   ParamType = 3
	ParamTypeIntegers ParamType = 4
	ParamTypeFloats   ParamType = 5
	ParamTypeStrings  ParamType = 6
	ParamTypeBoolean  ParamType = 7
)

type ParamUnit string

const (
	ParamUnitNil ParamUnit = ""
	ParamUnitGb  ParamUnit = "gb"
)

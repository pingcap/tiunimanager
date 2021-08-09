package knowledge

type Parameter struct {
	Name         string				`json:"name"`
	Type         ParamType			`json:"type"`
	Unit         ParamUnit			`json:"unit"`
	DefaultValue interface{}		`json:"defaultValue"`
	NeedRestart  bool				`json:"needRestart"`
	Constraints  []ParamValueConstraint	`json:"constraints"`
	Desc         string				`json:"desc"`
}

type ParamValueConstraint struct {
	Type          ConstraintType	`json:"type"`
	ContrastValue interface{}		`json:"contrastValue"`
}

type ConstraintType string

const (
	ConstraintTypeLT = "lt"
	ConstraintTypeLTE = "lte"

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
	ParamUnitKb  ParamUnit = "kb"
	ParamUnitMb  ParamUnit = "mb"
	ParamUnitGb  ParamUnit = "gb"
)

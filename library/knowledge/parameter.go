
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

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

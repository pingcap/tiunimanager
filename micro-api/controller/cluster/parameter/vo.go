package parameter

import (
	"github.com/pingcap-inc/tiem/library/knowledge"
)

type ParamItem struct {
	Definition   knowledge.Parameter `json:"definition"`
	CurrentValue ParamInstance       `json:"currentValue"`
}

type ParamInstance struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

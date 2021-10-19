package parameter

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

type ParamQueryReq struct {
	controller.PageRequest
}

type ParamUpdateReq struct {
	Values []ParamInstance `json:"values"`
}

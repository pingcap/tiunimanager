package instanceapi

import "github.com/pingcap/ticp/micro-api/controller"

type ParamQueryReq struct {
	ClusterId 		string
	Page 			controller.Page
}

type ParamUpdateReq struct {
	ClusterId 		string
	Values			[]ParamValue
}
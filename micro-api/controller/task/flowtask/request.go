package flowtask

import "github.com/pingcap-inc/tiem/micro-api/controller"

type QueryReq struct {
	controller.PageRequest
	Status    int    `json:"status" form:"status"`
	Keyword   string `json:"keyword" form:"keyword"`
	ClusterId string `json:"clusterId" form:"clusterId"`
}

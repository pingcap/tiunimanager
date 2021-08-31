package taskapi

import "github.com/pingcap-inc/tiem/micro-api/controller"

type QueryReq struct {
	controller.PageRequest

	Status 		int		`json:"status"`
	Keyword 	string	`json:"keyword"`
	ClusterId 	string 	`json:"clusterId"`
}
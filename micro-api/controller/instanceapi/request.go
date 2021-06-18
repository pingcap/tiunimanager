package instanceapi

import (
	"github.com/pingcap/ticp/micro-api/controller"
)

type InstanceQuery struct {
	controller.PageRequest
}

type InstanceCreate struct {
	InstanceName 		string 	`json:"instanceName"`
	InstanceVersion 	int 	`json:"instanceVersion"`
	DBPassword 			string 	`json:"dbPassword"`
	PDCount 			int 	`json:"pdCount"`
	TiDBCount 			int 	`json:"tiDBCount"`
	TiKVCount 			int 	`json:"tiKVCount"`
}

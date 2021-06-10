package instanceapi

import (
	"github.com/pingcap/ticp/api"
)

type InstanceQuery struct {
	api.PageRequest
}

type InstanceCreate struct {
	InstanceName 		string 	`json:"InstanceName"`
	InstanceVersion 	int 	`json:"instanceVersion"`
	DBPassword 			int 	`json:"dbPassword"`
	PDCount 			int 	`json:"pdCount"`
	TiDBCount 			int 	`json:"tiDBCount"`
	TiKVCount 			int 	`json:"tiKVCount"`
}

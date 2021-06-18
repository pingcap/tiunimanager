package service

import (
	manager "github.com/pingcap/ticp/micro-manager/proto"
)

var TiCPManagerServiceName = "go.micro.ticp.manager"

var SuccessResponseStatus = &manager.ResponseStatus {Code:0}

type ManagerServiceHandler struct {}

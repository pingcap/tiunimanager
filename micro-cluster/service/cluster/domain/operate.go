package domain

import (
	"time"
)

type Operator struct {
	Id 			string
	Name 		string
	TenantId 	string
}

type OperateRecord struct {
	Id 				uint
	TraceId	 		string
	Operator    	Operator
	OperateTime 	time.Time
	OperateType 	int
	FlowWorkId 		string
}

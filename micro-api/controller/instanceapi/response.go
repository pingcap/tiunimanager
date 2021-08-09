package instanceapi

type ParamUpdateRsp struct {
	Status 			string	`json:"status"`
	ClusterId 		string	`json:"clusterId"`
	TaskId			uint	`json:"taskId"`
}
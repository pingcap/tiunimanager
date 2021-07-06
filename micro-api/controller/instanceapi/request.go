package instanceapi

type ParamUpdateReq struct {
	ClusterId 		string
	Values			[]ParamInstance
}

type BackupStrategyUpdateReq struct {
	ClusterId string
	BackupStrategy
}

type BackupRecoverReq struct {
	ClusterId 			string
	BackupRecordId	 	string
}

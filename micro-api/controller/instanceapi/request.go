package instanceapi

type ParamUpdateReq struct {
	ClusterId 		string
	Values			[]ParamValue
}

type BackupStrategyUpdateReq struct {
	ClusterId string
	BackupStrategy
}

type BackupRecoverReq struct {
	ClusterId 			string
	BackupRecordId	 	string
}

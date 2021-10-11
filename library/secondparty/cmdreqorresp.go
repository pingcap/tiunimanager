package secondparty

type CmdDeployReq struct {
	TaskID        uint64
	InstanceName  string
	Version       string
	ConfigStrYaml string
	TimeoutS      int
	TiupPath      string
	Flags         []string
}

type CmdStartReq struct {
	TaskID       uint64
	InstanceName string
	TimeoutS     int
	TiupPath     string
	Flags        []string
}

type CmdListReq struct {
	TaskID   uint64
	TimeoutS int
	TiupPath string
	Flags    []string
}

type CmdDestroyReq struct {
	TaskID       uint64
	InstanceName string
	TimeoutS     int
	TiupPath     string
	Flags        []string
}

type CmdListResp struct {
	ListRespStr  string
}

type CmdGetAllTaskStatusResp struct {
	Stats []TaskStatusMember
}

type CmdDumplingReq struct {
	TaskID   uint64
	TimeoutS int
	TiupPath string
	Flags []string
}

type CmdLightningReq struct {
	TaskID   uint64
	TimeoutS int
	TiupPath string
	Flags    []string
}

type CmdClusterDisplayReq struct {
	ClusterName string
	TimeoutS    int
	TiupPath    string
	Flags       []string
}

type CmdClusterDisplayResp struct {
	DisplayRespString string
}

type CmdBackUpReq struct {
	TaskID            uint64
	DbName            string
	TableName         string
	FilterDbTableName string // used in br command, pending for use in SQL command
	StorageAddress    string
	DbConnParameter   DbConnParam // only for SQL command, not used in br command
	RateLimitM        string
	Concurrency       string   // only for SQL command, not used in br command
	CheckSum          string   // only for SQL command, not used in br command
	LogFile           string   // used in br command, pending for use in SQL command
	TimeoutS          int      // used in br command, pending for use in SQL command
	Flags             []string // used in br command, pending for use in SQL command
}

type CmdBrResp struct {
	Destination    string
	Size           uint64
	BackupTS       uint64
	Queue_time     string
	Execution_Time string
}

type CmdShowBackUpInfoReq struct {
	TaskID          uint64
	DbConnParameter DbConnParam
}

type CmdShowBackUpInfoResp struct {
	Destination    string
	Size           uint64
	BackupTS       uint64
	State          string
	Progress       float32
	Queue_time     string
	Execution_Time string
	Finish_Time    *string
	Connection     string
	ErrorStr       string
}

type CmdRestoreReq struct {
	TaskID            uint64
	DbName            string
	TableName         string
	FilterDbTableName string // used in br command, pending for use in SQL command
	StorageAddress    string
	DbConnParameter   DbConnParam // only for SQL command, not used in br command
	RateLimitM        string
	Concurrency       string   // only for SQL command, not used in br command
	CheckSum          string   // only for SQL command, not used in br command
	LogFile           string   // used in br command, pending for use in SQL command
	TimeoutS          int      // used in br command, pending for use in SQL command
	Flags             []string // used in br command, pending for use in SQL command
}

type CmdShowRestoreInfoReq struct {
	TaskID          uint64
	DbConnParameter DbConnParam
}

type CmdShowRestoreInfoResp struct {
	Destination    string
	Size           uint64
	BackupTS       uint64
	State          string
	Progress       float32
	Queue_time     string
	Execution_Time string
	Finish_Time    *string
	Connection     string
	ErrorStr       string
}
package common

// micro service default port
const (
	DefaultMicroMetaDBPort  int = 4100
	DefaultMicroClusterPort     = 4110
	DefaultMicroApiPort         = 4116
)

const (
	LogDirPrefix  string = "/logs/"
	CertDirPrefix string = "/cert/"
	DBDirPrefix	  string = "/"

	SqliteFileName string = "tiem.sqlite.db"

	CrtFileName string = "server.crt"
	KeyFileName string = "server.key"

)

const (
	LOG_FILE_SYSTEM = "system"
	LOG_FILE_TIUP_MGR = "tiupmgr"
	LOG_FILE_BR_MGR = "tiupmgr"
	LOG_FILE_LIB_TIUP = "libtiup"
	LOG_FILE_LIB_BR = "tiupmgr"

	LOG_FILE_ACCESS = "access"
	LOG_FILE_AUDIT = "audit"
)
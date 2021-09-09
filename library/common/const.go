package common

// micro service default port
const (
	DefaultMicroMetaDBPort  int = 4100
	DefaultMicroClusterPort     = 4110
	DefaultMicroApiPort         = 4116
	DefaultMetricsPort          = 4121
)

const (
	TiEM          string = "tiem"
	LogDirPrefix  string = "/logs/"
	CertDirPrefix string = "/cert/"
	DBDirPrefix   string = "/"

	SqliteFileName string = "tiem.sqlite.db"

	CrtFileName string = "server.crt"
	KeyFileName string = "server.key"

	LocalAddress string = "0.0.0.0"
)

const (
	LogFileSystem  = "system"
	LogFileTiupMgr = "tiupmgr"
	LogFileBrMgr   = "tiupmgr"
	LogFileLibTiup = "libtiup"
	LogFileLibBr   = "tiupmgr"

	LogFileAccess = "access"
	LogFileAudit  = "audit"
)

const (
	RegistryMicroServicePrefix = "/micro/registry/"
	HttpProtocol               = "http://"
)

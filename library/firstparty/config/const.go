package config

// micro service default port
const (
	DefaultMicroMetaDBPort  int = 4100
	DefaultMicroClusterPort     = 4110
	DefaultMicroManagerPort     = 4111
	DefaultMicroApiPort         = 4115
	DefaultRestPort             = 4116
)

const (
	LogDirPrefix  string = "/logs/"
	CertDirPrefix string = "/cert/"

	SqliteFileName string = "tiem.sqlite.db"

	CrtFileName string = "server.crt"
	KeyFileName string = "server.key"

	MicroApiLogFileName     string = "micro-api.log"
	MicroClusterLogFileName string = "micro-cluster.log"
	MicroMetaDBLogFileName  string = "micro-metadb.log"
	TiUPLogFileName         string = "utils-tiup.log"
	DefaultLogFileName      string = "tiem.log"
)

package config

func InitForMonolith() {
	LocalConfig = make(map[Key]Instance)

	LocalConfig[KEY_REGISTRY_ADDRESS] = Instance{
		KEY_REGISTRY_ADDRESS,
		[]string{"127.0.0.1:2379"},
		0,
	}

	LocalConfig[KEY_CERTIFICATES] = Instance{
		KEY_CERTIFICATES,
		Certificates{
			CrtFilePath: "library/firstparty/config/example/server.crt",
			KeyFilePath: "library/firstparty/config/example/server.key",
		},
		0,
	}

	LocalConfig[KEY_SQLITE_FILE_PATH] = CreateInstance(KEY_SQLITE_FILE_PATH, "./tiem.sqlite.db")

	LocalConfig[KEY_API_LOG] = CreateInstance(KEY_API_LOG, Log {
		//LogLevel      string
		LogLevel:      "debug",
		LogOutput:     "console,file",
		LogFilePath:   "../logs/micro-api.log",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
		RecordSysName: "tiem",
		RecordModName: "example",
	})

	LocalConfig[KEY_CLUSTER_LOG] = CreateInstance(KEY_CLUSTER_LOG, Log{
		//LogLevel      string
		LogLevel:      "debug",
		LogOutput:     "console,file",
		LogFilePath:   "../logs/micro-cluster.log",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
		RecordSysName: "tiem",
		RecordModName: "example",
	})

	LocalConfig[KEY_METADB_LOG] = CreateInstance(KEY_METADB_LOG, Log{
		//LogLevel      string
		LogLevel:      "debug",
		LogOutput:     "console,file",
		LogFilePath:   "../logs/micro-metadb.log",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
		RecordSysName: "tiem",
		RecordModName: "example",
	})

	LocalConfig[KEY_API_PORT] = CreateInstance(KEY_API_PORT, 443)
	LocalConfig[KEY_CLUSTER_PORT] = CreateInstance(KEY_CLUSTER_PORT, 444)
	LocalConfig[KEY_METADB_PORT] = CreateInstance(KEY_API_PORT, 443)
}

func InitForTiUPCluster() {
	//
}

func GetLogConfig() Log {
	// todo : get from LocalConfig
	return  Log {
		LogLevel: "debug",
		LogOutput: "console,file",
		LogFilePath: "../logs/tiem.log",
		LogMaxSize: 512,
		LogMaxAge: 30,
		LogMaxBackups: 0,
		LogLocalTime: true,
		LogCompress: true,
		RecordSysName: "tiem",
		RecordModName:"example",
	}
}

func GetSqliteFilePath() string {
	return GetStringWithDefault(KEY_SQLITE_FILE_PATH, "")
}

func GetCertificateCrtFilePath() string {
	c, err := Get(KEY_CERTIFICATES)
	if err == nil {
		return c.(Certificates).CrtFilePath
	} else {
		// todo
		return ""
	}
}

func GetCertificateKeyFilePath() string {
	c, err := Get(KEY_CERTIFICATES)
	if err == nil {
		return c.(Certificates).KeyFilePath
	} else {
		// todo
		return ""
	}
}

func GetRegistryAddress() []string {
	v, err := Get(KEY_REGISTRY_ADDRESS)
	if err != nil {
		panic("registry empty")
	}

	return v.([]string)
}

// Log Corresponding to [Log] in cfg.toml configuration
type Log struct {
	LogLevel      string
	LogOutput     string
	LogFilePath   string
	LogMaxSize    int
	LogMaxAge     int
	LogMaxBackups int
	LogLocalTime  bool
	LogCompress   bool

	RecordSysName string
	RecordModName string
}

type Certificates struct {
	CrtFilePath string
	KeyFilePath string
}

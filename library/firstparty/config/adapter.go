package config

import "strconv"

func InitConfigForDev(mod Mod) {
	LocalConfig = make(map[Key]Instance)

	LocalConfig[KEY_REGISTRY_ADDRESS] = Instance{
		KEY_REGISTRY_ADDRESS,
		[]string{"127.0.0.1:2379"},
		0,
	}

	LocalConfig[KEY_TRACER_ADDRESS] = Instance{KEY_TRACER_ADDRESS, "127.0.0.1:6831", 0}

	LocalConfig[KEY_CERTIFICATES] = Instance{
		KEY_CERTIFICATES,
		Certificates{
			CrtFilePath: "../library/firstparty/config/example/server.crt",
			KeyFilePath: "../library/firstparty/config/example/server.key",
		},
		0,
	}

	LocalConfig[KEY_SQLITE_FILE_PATH] = CreateInstance(KEY_SQLITE_FILE_PATH, "./tiem.sqlite.db")

	LocalConfig[KEY_API_LOG] = CreateInstance(KEY_API_LOG, Log{
		LogLevel:      "debug",
		LogOutput:     "console,file",
		LogFilePath:   "../logs/micro-api.log",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
		RecordSysName: "tiem",
		RecordModName: "openapi",
	})

	LocalConfig[KEY_CLUSTER_LOG] = CreateInstance(KEY_CLUSTER_LOG, Log{
		LogLevel:      "debug",
		LogOutput:     "console,file",
		LogFilePath:   "../logs/micro-cluster.log",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
		RecordSysName: "tiem",
		RecordModName: "cluster",
	})

	LocalConfig[KEY_METADB_LOG] = CreateInstance(KEY_METADB_LOG, Log{
		LogLevel:      "debug",
		LogOutput:     "console,file",
		LogFilePath:   "../logs/micro-metadb.log",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
		RecordSysName: "tiem",
		RecordModName: "metadb",
	})

	LocalConfig[KEY_TIUPLIB_LOG] = CreateInstance(KEY_TIUPLIB_LOG, Log{
		LogLevel:      "debug",
		LogOutput:     "file",
		LogFilePath:   "../logs/utils-tiup.log",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
		RecordSysName: "tiem",
		RecordModName: "tiuplib",
	})

	LocalConfig[KEY_BRLIB_LOG] = CreateInstance(KEY_BRLIB_LOG, Log{
		//LogLevel      string
		LogLevel:      "debug",
		LogOutput:     "file",
		LogFilePath:   "../logs/utils-br.log",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
		RecordSysName: "tiem",
		RecordModName: "brlib",
	})

	LocalConfig[KEY_DEFAULT_LOG] = CreateInstance(KEY_DEFAULT_LOG, Log{
		LogLevel:      "debug",
		LogOutput:     "console,file",
		LogFilePath:   "../logs/tiem.log",
		LogMaxSize:    512,
		LogMaxAge:     30,
		LogMaxBackups: 0,
		LogLocalTime:  true,
		LogCompress:   true,
		RecordSysName: "tiem",
		RecordModName: "default",
	})

	LocalConfig[KEY_API_PORT] = CreateInstance(KEY_API_PORT, DefaultMicroApiPort)
	LocalConfig[KEY_CLUSTER_PORT] = CreateInstance(KEY_CLUSTER_PORT, DefaultMicroClusterPort)
	LocalConfig[KEY_MANAGER_PORT] = CreateInstance(KEY_MANAGER_PORT, DefaultMicroManagerPort)
	LocalConfig[KEY_METADB_PORT] = CreateInstance(KEY_METADB_PORT, DefaultMicroMetaDBPort)

	// Init config for client args
	InitForClientArgs(mod)
}

func InitForTiUPCluster() {
	//
}

func GetApiServiceAddress() string {
	return ":" + strconv.Itoa(GetIntegerWithDefault(KEY_API_PORT, DefaultMicroApiPort))
}

func GetClusterServiceAddress() string {
	return ":" + strconv.Itoa(GetIntegerWithDefault(KEY_CLUSTER_PORT, DefaultMicroClusterPort))
}

func GetManagerServiceAddress() string {
	return ":" + strconv.Itoa(GetIntegerWithDefault(KEY_MANAGER_PORT, DefaultMicroManagerPort))
}

func GetMetaDBServiceAddress() string {
	return ":" + strconv.Itoa(GetIntegerWithDefault(KEY_METADB_PORT, DefaultMicroMetaDBPort))
}

func GetLogConfig(key Key) Log {
	switch key {
	case KEY_API_LOG:
		return LocalConfig[KEY_API_LOG].Value.(Log)
	case KEY_CLUSTER_LOG:
		return LocalConfig[KEY_CLUSTER_LOG].Value.(Log)
	case KEY_METADB_LOG:
		return LocalConfig[KEY_METADB_LOG].Value.(Log)
	case KEY_TIUPLIB_LOG:
		return LocalConfig[KEY_TIUPLIB_LOG].Value.(Log)
	case KEY_BRLIB_LOG:
		return LocalConfig[KEY_BRLIB_LOG].Value.(Log)
	default:
		return LocalConfig[KEY_DEFAULT_LOG].Value.(Log)
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

func GetTracerAddress() string {
	v, err := Get(KEY_TRACER_ADDRESS)
	if err != nil {
		panic("tracer empty")
	}
	return v.(string)
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

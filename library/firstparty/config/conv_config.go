package config

import (
	"strings"
)

// Mod Define configuration module type
type Mod string

const (
	MicroMetaDBMod  Mod = "micro-metadb"
	MicroClusterMod Mod = "micro-cluster"
	MicroApiMod     Mod = "micro-api"
	TiUPInternalMod Mod = "tiup"
)

// InitForClientArgs Init config for client args
func InitForClientArgs(mod Mod) {
	// Prefer client args
	registryAddress := clientArgs.RegistryAddress
	convInitRegistryAddressConfig(registryAddress)

	deployDir := clientArgs.DeployDir
	convInitCertConfig(deployDir)

	port := clientArgs.Port
	dataDir := clientArgs.DataDir

	// log level
	logLevel := clientArgs.LogLevel

	switch mod {
	case MicroMetaDBMod:
		convInitMicroMetaDBConfig(port, dataDir, logLevel)
		break
	case MicroClusterMod:
		convInitMicroClusterConfig(port, dataDir, logLevel)
		break
	case MicroApiMod:
		convInitMicroApiConfig(port, dataDir, logLevel)
		break
	case TiUPInternalMod:
		// update log config
		log := LocalConfig[KEY_TIUPLIB_LOG].Value.(Log)
		if dataDir != "" {
			log.LogFilePath = clientArgs.DataDir + LogDirPrefix + TiUPLogFileName
		}
		if logLevel != "" {
			log.LogLevel = logLevel
		}
		UpdateLocalConfig(KEY_TIUPLIB_LOG, log, 1)
		break
	default:
		// update log config
		log := LocalConfig[KEY_DEFAULT_LOG].Value.(Log)
		if dataDir != "" {
			log.LogFilePath = clientArgs.DataDir + LogDirPrefix + DefaultLogFileName
		}
		if logLevel != "" {
			log.LogLevel = logLevel
		}
		UpdateLocalConfig(KEY_DEFAULT_LOG, log, 1)
	}
}

// convert client-args init cert file config
func convInitCertConfig(deployDir string) {
	if deployDir != "" {
		UpdateLocalConfig(KEY_CERTIFICATES, Certificates{
			CrtFilePath: clientArgs.DeployDir + CertDirPrefix + CrtFileName,
			KeyFilePath: clientArgs.DeployDir + CertDirPrefix + KeyFileName,
		}, 1)
	}
}

// convert client-args init registry address config
func convInitRegistryAddressConfig(registryAddress string) {
	if registryAddress != "" {
		addresses := strings.Split(clientArgs.RegistryAddress, ",")
		registryAddresses := make([]string, len(addresses))
		for i, addr := range addresses {
			registryAddresses[i] = addr
		}
		UpdateLocalConfig(KEY_REGISTRY_ADDRESS, registryAddresses, 1)
	}
}

// convert client-args init micro-api config
func convInitMicroApiConfig(port int, dataDir, logLevel string) {
	if port > 0 {
		// micro-api server port
		UpdateLocalConfig(KEY_API_PORT, port, 1)
	}
	// update micro-api log config
	log := LocalConfig[KEY_API_LOG].Value.(Log)
	if dataDir != "" {
		log.LogFilePath = clientArgs.DataDir + LogDirPrefix + MicroApiLogFileName
	}
	if logLevel != "" {
		log.LogLevel = logLevel
	}
	UpdateLocalConfig(KEY_API_LOG, log, 1)
}

// convert client-args init micro-cluster config
func convInitMicroClusterConfig(port int, dataDir, logLevel string) {
	if port > 0 {
		// micro-cluster server port
		UpdateLocalConfig(KEY_CLUSTER_PORT, port, 1)
	}

	// update micro-cluster log config
	log := LocalConfig[KEY_CLUSTER_LOG].Value.(Log)
	if dataDir != "" {
		log.LogFilePath = clientArgs.DataDir + LogDirPrefix + MicroClusterLogFileName
	}
	if logLevel != "" {
		log.LogLevel = logLevel
	}
	UpdateLocalConfig(KEY_CLUSTER_LOG, log, 1)
}

// convert client-args init micro-metadb config
func convInitMicroMetaDBConfig(port int, dataDir, logLevel string) {
	if port > 0 {
		// micro-metadb server port
		UpdateLocalConfig(KEY_METADB_PORT, port, 1)
	}

	// micro-metadb log config
	log := LocalConfig[KEY_METADB_LOG].Value.(Log)
	if dataDir != "" {
		// sqlite db file path config
		UpdateLocalConfig(KEY_SQLITE_FILE_PATH, clientArgs.DataDir+"/"+SqliteFileName, 1)

		log.LogFilePath = clientArgs.DataDir + LogDirPrefix + MicroMetaDBLogFileName
	}
	if logLevel != "" {
		log.LogLevel = logLevel
	}
	UpdateLocalConfig(KEY_METADB_LOG, log, 1)
}

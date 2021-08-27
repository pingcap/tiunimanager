package config

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Mod Define configuration module type
type Mod string

const (
	MicroMetaDBMod  Mod = "micro-metadb"
	MicroClusterMod Mod = "micro-cluster"
	MicroApiMod     Mod = "micro-api"
	TiUPInternalMod Mod = "tiup"
	BRInternalMod 	Mod = "br"
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
	case MicroClusterMod:
		convInitMicroClusterConfig(port, dataDir, logLevel)
	case MicroApiMod:
		convInitMicroApiConfig(port, dataDir, logLevel)
	case TiUPInternalMod:
		// update log config
		log := LocalConfig[KEY_TIUPLIB_LOG].Value.(Log)
		if dataDir != "" {
			log.LogFilePath = clientArgs.DataDir + LogDirPrefix + TiUPLogFileName
		}
		if logLevel != "" {
			log.LogLevel = logLevel
		}
		hasSuc, version := UpdateLocalConfig(KEY_TIUPLIB_LOG, log, 1)
		assertUpdateLocalConfig(hasSuc, version, 1)
	case BRInternalMod:
		// update log config
		log := LocalConfig[KEY_BRLIB_LOG].Value.(Log)
		if dataDir != "" {
			log.LogFilePath = clientArgs.DataDir + LogDirPrefix + BRLogFileName
		}
		if logLevel != "" {
			log.LogLevel = logLevel
		}
		hasSuc, version := UpdateLocalConfig(KEY_BRLIB_LOG, log, 1)
		assertUpdateLocalConfig(hasSuc, version, 1)
	default:
		convLogConfig(dataDir, KEY_DEFAULT_LOG, logLevel, DefaultLogFileName)
	}
}

// convert client-args init cert file config
func convInitCertConfig(deployDir string) {
	if deployDir != "" {
		hasSuc, version := UpdateLocalConfig(KEY_CERTIFICATES, Certificates{
			CrtFilePath: clientArgs.DeployDir + CertDirPrefix + CrtFileName,
			KeyFilePath: clientArgs.DeployDir + CertDirPrefix + KeyFileName,
		}, 1)
		assertUpdateLocalConfig(hasSuc, version, 1)
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
		hasSuc, version := UpdateLocalConfig(KEY_REGISTRY_ADDRESS, registryAddresses, 1)
		assertUpdateLocalConfig(hasSuc, version, 1)
	}
}

// convert client-args init micro-api config
func convInitMicroApiConfig(port int, dataDir string, logLevel string) {
	if port > 0 {
		// micro-api server port
		hasSuc, version := UpdateLocalConfig(KEY_API_PORT, port, 1)
		assertUpdateLocalConfig(hasSuc, version, 1)
	}
	// update micro-api log configØ
	convLogConfig(dataDir, KEY_API_LOG, logLevel, MicroApiLogFileName)
}

// convert client-args init micro-cluster config
func convInitMicroClusterConfig(port int, dataDir string, logLevel string) {
	if port > 0 {
		// micro-cluster server port
		hasSuc, version := UpdateLocalConfig(KEY_CLUSTER_PORT, port, 1)
		assertUpdateLocalConfig(hasSuc, version, 1)
	}
	// update micro-cluster log config
	convLogConfig(dataDir, KEY_CLUSTER_LOG, logLevel, MicroClusterLogFileName)
}

// convert client-args init micro-metadb config
func convInitMicroMetaDBConfig(port int, dataDir string, logLevel string) {
	if port > 0 {
		// micro-metadb server port
		hasSuc, version := UpdateLocalConfig(KEY_METADB_PORT, port, 1)
		assertUpdateLocalConfig(hasSuc, version, 1)
	}
	if dataDir != "" {
		// sqlite db file path config
		hasSuc, version := UpdateLocalConfig(KEY_SQLITE_FILE_PATH, clientArgs.DataDir+"/"+SqliteFileName, 1)
		assertUpdateLocalConfig(hasSuc, version, 1)
	}
	// update micro-metadb log config
	convLogConfig(dataDir, KEY_METADB_LOG, logLevel, MicroMetaDBLogFileName)
}

// convert base log config
func convLogConfig(dataDir string, logKey Key, logLevel, logFileName string) {
	// update log config
	log := LocalConfig[logKey].Value.(Log)
	if dataDir != "" {
		log.LogFilePath = clientArgs.DataDir + LogDirPrefix + logFileName
	}
	if logLevel != "" {
		log.LogLevel = logLevel
	}
	hasSuc, version := UpdateLocalConfig(logKey, log, 1)
	assertUpdateLocalConfig(hasSuc, version, 1)
}

func assertUpdateLocalConfig(hasSuc bool, version, expVersion int) {
	assert(hasSuc)
	assert(version == expVersion)
}

func assert(b bool) {
	if !b {
		fmt.Println("unexpected panic with stack trace:", string(debug.Stack()))
		panic("unexpected")
	}
}

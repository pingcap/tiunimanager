package framework
//
//import (
//	"fmt"
//	"github.com/pingcap-inc/tiem/library/framework/common"
//	"github.com/pingcap-inc/tiem/library/framework/config"
//	"runtime/debug"
//	"strings"
//)
//
//// Mod Define configuration module type
//type Mod string
//
//const (
//	MicroMetaDBMod  Mod = "micro-metadb"
//	MicroClusterMod Mod = "micro-cluster"
//	MicroApiMod     Mod = "micro-api"
//	TiUPInternalMod Mod = "tiup"
//)
//
//// InitForClientArgs Init config for client args
//func InitForClientArgs(mod Mod) {
//	// Prefer client args
//	registryAddress := clientArgs.RegistryAddress
//	convInitRegistryAddressConfig(registryAddress)
//
//	deployDir := clientArgs.DeployDir
//	convInitCertConfig(deployDir)
//
//	port := clientArgs.Port
//	dataDir := clientArgs.DataDir
//
//	// log level
//	logLevel := clientArgs.LogLevel
//
//	switch mod {
//	case MicroMetaDBMod:
//		convInitMicroMetaDBConfig(port, dataDir, logLevel)
//	case MicroClusterMod:
//		convInitMicroClusterConfig(port, dataDir, logLevel)
//	case MicroApiMod:
//		convInitMicroApiConfig(port, dataDir, logLevel)
//	case TiUPInternalMod:
//		// update log config
//		log := config.LocalConfig[config.KEY_TIUPLIB_LOG].Value.(Log)
//		if dataDir != "" {
//			log.LogFilePath = clientArgs.DataDir + common.LogDirPrefix + common.TiUPLogFileName
//		}
//		if logLevel != "" {
//			log.LogLevel = logLevel
//		}
//		hasSuc, version := config.UpdateLocalConfig(config.KEY_TIUPLIB_LOG, log, 1)
//		assertUpdateLocalConfig(hasSuc, version, 1)
//	default:
//		convLogConfig(dataDir, config.KEY_DEFAULT_LOG, logLevel, common.DefaultLogFileName)
//	}
//}
//
//// convert client-args init micro-api config
//func convInitMicroApiConfig(port int, dataDir string, logLevel string) {
//	if port > 0 {
//		// micro-api server port
//		hasSuc, version := config.UpdateLocalConfig(config.KEY_API_PORT, port, 1)
//		assertUpdateLocalConfig(hasSuc, version, 1)
//	}
//	// update micro-api log configØ
//	convLogConfig(dataDir, config.KEY_API_LOG, logLevel, common.MicroApiLogFileName)
//}
//
//// convert client-args init micro-cluster config
//func convInitMicroClusterConfig(port int, dataDir string, logLevel string) {
//	if port > 0 {
//		// micro-cluster server port
//		hasSuc, version := config.UpdateLocalConfig(config.KEY_CLUSTER_PORT, port, 1)
//		assertUpdateLocalConfig(hasSuc, version, 1)
//	}
//	// update micro-cluster log config
//	convLogConfig(dataDir, config.KEY_CLUSTER_LOG, logLevel, common.MicroClusterLogFileName)
//}
//
//// convert client-args init micro-metadb config
//func convInitMicroMetaDBConfig(port int, dataDir string, logLevel string) {
//	if port > 0 {
//		// micro-metadb server port
//		hasSuc, version := config.UpdateLocalConfig(config.KEY_METADB_PORT, port, 1)
//		assertUpdateLocalConfig(hasSuc, version, 1)
//	}
//	if dataDir != "" {
//		// sqlite db file path config
//		hasSuc, version := config.UpdateLocalConfig(config.KEY_SQLITE_FILE_PATH, clientArgs.DataDir+"/"+common.SqliteFileName, 1)
//		assertUpdateLocalConfig(hasSuc, version, 1)
//	}
//	// update micro-metadb log config
//	convLogConfig(dataDir, config.KEY_METADB_LOG, logLevel, common.MicroMetaDBLogFileName)
//}
//
//// convert base log config
//func convLogConfig(dataDir string, logKey config.Key, logLevel, logFileName string) {
//	// update log config
//	log := config.LocalConfig[logKey].Value.(Log)
//	if dataDir != "" {
//		log.LogFilePath = clientArgs.DataDir + common.LogDirPrefix + logFileName
//	}
//	if logLevel != "" {
//		log.LogLevel = logLevel
//	}
//	hasSuc, version := config.UpdateLocalConfig(logKey, log, 1)
//	assertUpdateLocalConfig(hasSuc, version, 1)
//}
//
//func assertUpdateLocalConfig(hasSuc bool, version, expVersion int) {
//	assert(hasSuc)
//	assert(version == expVersion)
//}
//
//func assert(b bool) {
//	if !b {
//		fmt.Println("unexpected panic with stack trace:", string(debug.Stack()))
//		panic("unexpected")
//	}
//}

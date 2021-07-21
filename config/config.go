package config

import (
	"log"

	"github.com/BurntSushi/toml"
	"github.com/asim/go-micro/v3"
	"github.com/micro/cli/v2"
)

var configFilePath = ""

func GetMicroCliArgsOption() micro.Option {
	return micro.Flags(
		&cli.StringFlag{
			Name:        "tidb-cloud-platform-conf-file",
			Value:       "",
			Usage:       "specify the configure file path of tidb cloud platform",
			Destination: &configFilePath,
		},
	)
}

func GetConfigFilePath() string {
	return configFilePath
}

type Config struct {
	SqliteFilePath string
	Certificates   Certificates
	Log            Log
	OpenApiPort    int
	PrometheusPort int
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

func GetLogConfig() Log {
	return cfg.Log
}

type Certificates struct {
	CrtFilePath string
	KeyFilePath string
}

func GetSqliteFilePath() string {
	return cfg.SqliteFilePath
}

func GetOpenApiPort() int {
	return cfg.OpenApiPort
}

func GetPrometheusPort() int {
	return cfg.PrometheusPort
}

var cfg Config

func Init() error {
	_, err := toml.DecodeFile(GetConfigFilePath(), &cfg)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func GetCertificateCrtFilePath() string {
	return cfg.Certificates.CrtFilePath
}

func GetCertificateKeyFilePath() string {
	return cfg.Certificates.KeyFilePath
}

package config

type Config struct {
	MysqlConf string
}

func GetCertificateCrtFilePath() string {
	return "../../config/example/server.crt"
}

func GetCertificateKeyFilePath() string {
	return "../../config/example/server.key"
}

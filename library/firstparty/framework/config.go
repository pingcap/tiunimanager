package framework

import "github.com/pingcap/tiem/library/firstparty/util"

type Config interface {
	GetCertificateCrtFilePath() string
	GetCertificateKeyFilePath() string
	GetRegistryAddress() []string
	// To be completed
	// ...
}

func initConfig() (Config, error) {
	util.AssertWithInfo(false, "Not Implemented Yet")
	return nil, nil
}

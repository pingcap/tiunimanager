package certificate

import (
	"github.com/pingcap-inc/tiem/library/framework/args"
	"github.com/pingcap-inc/tiem/library/framework/common"
)

type CertificateInfo struct {
	CertificateCrtFilePath string
	CertificateKeyFilePath string
}

func NewCertificateFromArgs(args *args.ClientArgs) *CertificateInfo {
	return &CertificateInfo{
		CertificateCrtFilePath: args.DeployDir + common.CertDirPrefix + common.CrtFileName,
		CertificateKeyFilePath: args.DeployDir + common.CertDirPrefix + common.KeyFileName,
	}
}
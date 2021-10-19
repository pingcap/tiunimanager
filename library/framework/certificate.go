package framework

import (
	"github.com/pingcap-inc/tiem/library/common"
)

type CertificateInfo struct {
	CertificateCrtFilePath string
	CertificateKeyFilePath string
}

func NewCertificateFromArgs(args *ClientArgs) *CertificateInfo {
	return &CertificateInfo{
		CertificateCrtFilePath: args.DeployDir + common.CertDirPrefix + common.CrtFileName,
		CertificateKeyFilePath: args.DeployDir + common.CertDirPrefix + common.KeyFileName,
	}
}
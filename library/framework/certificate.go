package framework

import (
	common2 "github.com/pingcap-inc/tiem/library/common"
)

type CertificateInfo struct {
	CertificateCrtFilePath string
	CertificateKeyFilePath string
}

func NewCertificateFromArgs(args *ClientArgs) *CertificateInfo {
	return &CertificateInfo{
		CertificateCrtFilePath: args.DeployDir + common2.CertDirPrefix + common2.CrtFileName,
		CertificateKeyFilePath: args.DeployDir + common2.CertDirPrefix + common2.KeyFileName,
	}
}
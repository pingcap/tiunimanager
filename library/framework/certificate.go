/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package framework

import (
	"github.com/pingcap/tiunimanager/common/constants"
)

type CertificateInfo struct {
	CertificateCrtFilePath string
	CertificateKeyFilePath string
}

func NewCertificateFromArgs(args *ClientArgs) *CertificateInfo {
	return &CertificateInfo{
		CertificateCrtFilePath: args.DeployDir + constants.CertDirPrefix + constants.CertFileName,
		CertificateKeyFilePath: args.DeployDir + constants.CertDirPrefix + constants.KeyFileName,
	}
}

func NewAesKeyFilePathFromArgs(args *ClientArgs) string {
	return args.DeployDir + constants.CertDirPrefix + constants.AesKeyFileName
}

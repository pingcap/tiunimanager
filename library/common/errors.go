// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

// tiem all errno
const (
	TIEM_SUCCESS 	= 0
	TIEM_PARAMETER_INVALID = 1

	TIEM_ACCOUNT_NOT_FOUND = 100
	TIME_ACCOUNT_EXIST = 101

	TIEM_TENANT_NOT_FOUND = 200
	TIEM_TENANT_EXIST =201

	TIEM_QUERY_PERMISSION_FAILED = 300

	TIEM_ADD_TOKEN_FAILED = 400
	TIEM_TOKEN_NOT_FOUND = 401
)

var TiEMErrMsg = map[uint32] string {
	TIEM_PARAMETER_INVALID : "parameter is invalid",
	TIEM_ACCOUNT_NOT_FOUND : "account is not found",
	TIME_ACCOUNT_EXIST : "account is exist",

	TIEM_TENANT_NOT_FOUND : "tenant is not found",
	TIEM_TENANT_EXIST : "tenant is exist",

	TIEM_QUERY_PERMISSION_FAILED :"query permission failed",

	TIEM_ADD_TOKEN_FAILED :"add token failed",

	TIEM_TOKEN_NOT_FOUND : "token not found",

}

var TiEMErrName = map[uint32]string{
	TIEM_SUCCESS: "successufl",
}

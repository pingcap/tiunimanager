/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package service

const (
	FILE_SUCCESS           = 0
	FILE_CANT_PARSE_FORM   = 100
	FILE_INVALID 		   = 101
	FILE_TOO_BIG     	   = 102
	FILE_INVALID_TYPE	   = 103
	FILE_CANT_WRITE		   = 104
)

var TiEMErrMsg = map[uint32]string{
	FILE_SUCCESS:           "success",
	FILE_CANT_PARSE_FORM:	"can not parse multipart form",
}

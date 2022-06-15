/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
 ******************************************************************************/

package controller

type FileResultMark struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CommonFileResult struct {
	FileResultMark
	Data interface{} `json:"data"`
}

func Success(data interface{}) *CommonFileResult {
	return &CommonFileResult{FileResultMark: FileResultMark{0, "OK"}, Data: data}
}

func Fail(code int, message string) *CommonFileResult {
	return &CommonFileResult{FileResultMark{code, message}, struct{}{}}
}

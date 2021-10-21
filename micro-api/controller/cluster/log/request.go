
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

package log

type SearchTiDBLogReq struct {
	Module    string `form:"module" example:"tidb"`
	Level     string `form:"level" example:"warn"`
	Ip        string `form:"ip" example:"127.0.0.1"`
	Message   string `form:"message" example:"tidb log"`
	StartTime string `form:"startTime" example:"2021-09-01 12:00:00"`
	EndTime   string `form:"endTime" example:"2021-12-01 12:00:00"`
	Page      int    `form:"page" example:"1"`
	PageSize  int    `form:"pageSize" example:"10"`
}

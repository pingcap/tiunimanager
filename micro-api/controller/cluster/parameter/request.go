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

package parameter

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

type ParamQueryReq struct {
	controller.PageRequest
}

type UpdateParamsReq struct {
	Params     []UpdateParam `json:"params"`
	NeedReboot bool          `json:"needReboot" example:"false"`
}

type UpdateParam struct {
	ParamId       int64          `json:"paramId" example:"1"`
	Name          string         `json:"name" example:"binlog_cache"`
	ComponentType string         `json:"componentType" example:"TiDB"`
	HasReboot     int32          `json:"hasReboot" example:"0" enums:"0,1"`
	Source        int32          `json:"source" example:"0" enums:"0,1,2,3"`
	Type          int32          `json:"type" example:"0" enums:"0,1,2"`
	RealValue     ParamRealValue `json:"realValue"`
}

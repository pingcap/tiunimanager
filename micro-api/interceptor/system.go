/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

/*******************************************************************************
 * @File: system.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/23
*******************************************************************************/

package interceptor

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/common/client"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/micro-api/controller"
	"github.com/pingcap/tiunimanager/proto/clusterservices"
	"net/http"
)

func SystemRunning(c *gin.Context) {
	req := message.GetSystemInfoReq{
		WithVersionDetail: false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		framework.LogWithContext(c).Errorf("marshal request error: %s", err.Error())
		c.Error(err)
		c.Status(errors.TIUNIMANAGER_MARSHAL_ERROR.GetHttpCode())
		c.Abort()
	}
	rpcResp, err := client.ClusterClient.GetSystemInfo(framework.NewMicroCtxFromGinCtx(c), &clusterservices.RpcRequest{Request: string(body)}, controller.DefaultTimeout)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		c.Abort()
	} else if rpcResp.Code != int32(errors.TIUNIMANAGER_SUCCESS) {
		framework.LogWithContext(c).Error(rpcResp.Message)
		code := errors.EM_ERROR_CODE(rpcResp.Code)
		msg := rpcResp.Message
		c.JSON(code.GetHttpCode(), controller.Fail(int(code), msg))
		c.Abort()
	} else {
		var result message.GetSystemInfoResp
		err = json.Unmarshal([]byte(rpcResp.Response), &result)
		if err != nil {
			framework.LogWithContext(c).Errorf("unmarshal get system info rpc response error: %s", err.Error())
			c.Error(err)
			c.Status(errors.TIUNIMANAGER_UNMARSHAL_ERROR.GetHttpCode())
			c.Abort()
		}
		if result.Info.State != string(constants.SystemRunning) {
			stateErr := errors.NewErrorf(errors.TIUNIMANAGER_SYSTEM_STATE_CONFLICT, "current state is %s", result.Info.State)
			framework.LogWithContext(c).Error(stateErr.Error())
			c.JSON(stateErr.GetCode().GetHttpCode(), controller.Fail(int(stateErr.GetCode()), stateErr.Error()))
			c.Abort()
		}
		c.Next()
	}
}

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
 *                                                                            *
 ******************************************************************************/

package interceptor

import (
	"encoding/json"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/file-server/controller"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/proto/clusterservices"
	utils "github.com/pingcap/tiunimanager/util/stringutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/common/client"
)

const VisitorIdentityKey = "VisitorIdentity"

type VisitorIdentity struct {
	AccountId   string
	AccountName string
	TenantId    string
}

func VerifyIdentity(c *gin.Context) {
	bearerTokenStr := c.GetHeader("Authorization")

	tokenString, err := utils.GetTokenFromBearer(bearerTokenStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
	}

	req := message.AccessibleReq{
		TokenString: structs.SensitiveText(tokenString),
		CheckPassword: true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		framework.LogWithContext(c).Errorf("marshal request error: %s", err.Error())
		c.Error(err)
		c.Status(errors.TIUNIMANAGER_MARSHAL_ERROR.GetHttpCode())
		c.Abort()
	}

	rpcResp, err := client.ClusterClient.VerifyIdentity(framework.NewMicroCtxFromGinCtx(c), &clusterservices.RpcRequest{Request: string(body)}, controller.DefaultTimeout)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		c.Abort()
	} else if rpcResp.Code != int32(errors.TIUNIMANAGER_SUCCESS) {
		framework.LogWithContext(c).Error(rpcResp.Message)
		c.Error(err)
		c.Status(errors.EM_ERROR_CODE(rpcResp.Code).GetHttpCode())
		c.Abort()
	} else {
		var result message.AccessibleResp
		err = json.Unmarshal([]byte(rpcResp.Response), &result)
		if err != nil {
			framework.LogWithContext(c).Errorf("unmarshal get system config rpc response error: %s", err.Error())
			c.Error(err)
			c.Status(errors.TIUNIMANAGER_UNMARSHAL_ERROR.GetHttpCode())
			c.Abort()
		}
		c.Set(framework.TiUniManager_X_USER_ID_KEY, result.UserID)
		c.Set(framework.TiUniManager_X_TENANT_ID_KEY, result.TenantID)
		c.Next()
	}
}

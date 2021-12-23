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

package interceptor

import (
	"encoding/json"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/file-server/controller"
	"github.com/pingcap-inc/tiem/message"
	"net/http"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	utils "github.com/pingcap-inc/tiem/library/util/stringutil"
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

	req := message.AccessibleReq {
		TokenString: tokenString,
	}

	body, err := json.Marshal(req)
	rpcResp, err := client.ClusterClient.VerifyIdentity(framework.NewMicroCtxFromGinCtx(c), &clusterpb.RpcRequest{Request: string(body)}, controller.DefaultTimeout)

	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		c.Abort()
	} else if rpcResp.Code != int32(errors.TIEM_SUCCESS) {
		framework.LogWithContext(c).Error(rpcResp.Message)
		c.Status(errors.EM_ERROR_CODE(rpcResp.Code).GetHttpCode())
		c.Abort()
	} else {
		var result message.AccessibleResp
		err = json.Unmarshal([]byte(rpcResp.Response), &result)

		c.Set(framework.TiEM_X_USER_ID_KEY, result.AccountID)
		c.Set(framework.TiEM_X_USER_NAME_KEY, result.AccountName)
		c.Set(framework.TiEM_X_TENANT_ID_KEY, result.TenantID)
		c.Next()
	}
}

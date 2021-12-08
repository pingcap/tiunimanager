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

	path := c.Request.URL
	req := clusterpb.VerifyIdentityRequest{TokenString: tokenString, Path: path.String()}

	result, err := client.ClusterClient.VerifyIdentity(framework.NewMicroCtxFromGinCtx(c), &req)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		c.Abort()
	} else if result.Status.Code != 0 {
		c.Status(int(result.Status.Code))
		c.Abort()
	} else {
		c.Set(VisitorIdentityKey, &VisitorIdentity{
			AccountId:   result.AccountId,
			AccountName: result.AccountName,
			TenantId:    result.TenantId,
		})
		c.Set(framework.TiEM_X_USER_ID_KEY, result.AccountId)
		c.Set(framework.TiEM_X_USER_NAME_KEY, result.AccountName)
		c.Set(framework.TiEM_X_TENANT_ID_KEY, result.TenantId)
		c.Next()
	}
}

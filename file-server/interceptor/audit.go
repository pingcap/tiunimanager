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
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	log "github.com/sirupsen/logrus"
)

func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		visitor := &VisitorIdentity{
			"unknown",
			"unknown",
			"unknown",
		}

		v, _ := c.Get(VisitorIdentityKey)
		if v != nil {
			visitor, _ = v.(*VisitorIdentity)
		}

		path := c.Request.URL.Path

		entry := framework.Current.GetRootLogger().ForkFile(common.LogFileAudit).WithFields(log.Fields{
			"operatorId":       visitor.AccountId,
			"operatorName":     visitor.AccountName,
			"operatorTenantId": visitor.TenantId,

			"clientIP":  c.ClientIP(),
			"method":    c.Request.Method,
			"path":      path,
			"referer":   c.Request.Referer(),
			"userAgent": c.Request.UserAgent(),
		})
		entry.Info("some do something")
	}
}

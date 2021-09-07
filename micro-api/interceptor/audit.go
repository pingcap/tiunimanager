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

		entry := framework.LogForkFile(common.LogFileAudit).WithFields(log.Fields{
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

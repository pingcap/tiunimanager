package interceptor

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
)

// Tiem-X-Trace-ID
func GinTraceIDHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(framework.TiEM_X_TRACE_ID_NAME)
		if len(id) <= 0 {
			id = uuidutil.GenerateID()
		}
		c.Set(framework.TiEM_X_TRACE_ID_NAME, id)
		c.Header(framework.TiEM_X_TRACE_ID_NAME, id)
		c.Next()
	}
}

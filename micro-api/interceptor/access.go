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
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/library/framework"
	log "github.com/sirupsen/logrus"
)

func AccessLog() gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		remoteIP, _ := c.RemoteIP()
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		entry := framework.LogForkFile(constants.LogFileAccess).WithFields(
			log.Fields{
				"hostname":   hostname,
				"clientIP":   c.ClientIP(),
				"remoteIP":   remoteIP,
				"status":     c.Writer.Status(),
				"latency":    latency, // time to process
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"dataLength": c.Writer.Size(),
				"referer":    c.Request.Referer(),
				"userAgent":  c.Request.UserAgent(),
			})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s - %s \"%s %s\" %d %d \"%s\" \"%s\" (%dms)",
				c.ClientIP(), hostname, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), c.Writer.Size(),
				c.Request.Referer(), c.Request.UserAgent(), latency)
			if c.Writer.Status() >= http.StatusInternalServerError {
				entry.Error(msg)
			} else if c.Writer.Status() >= http.StatusBadRequest {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}

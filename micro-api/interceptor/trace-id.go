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
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/util/uuidutil"
)

// GinTraceIDHandler EM-X-Trace-ID
func GinTraceIDHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(framework.TiUniManager_X_TRACE_ID_KEY)
		if len(id) <= 0 {
			id = uuidutil.GenerateID()
		}
		c.Set(framework.TiUniManager_X_TRACE_ID_KEY, id)
		c.Header(framework.TiUniManager_X_TRACE_ID_KEY, id)
		c.Next()
	}
}

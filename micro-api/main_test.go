
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/route"
	"testing"
)

var g *gin.Engine

func TestMain(m *testing.M) {
	f := framework.InitBaseFrameworkForUt(framework.ApiService)

	gin.SetMode(gin.ReleaseMode)
	g = gin.New()

	route.Route(g)
	//
	//port := f.GetServiceMeta().ServicePort
	//
	//addr := fmt.Sprintf(":%d", port)
	//
	//go func() {
	//	if err := g.Run(addr); err != nil {
	//		f.GetRootLogger().Fatal(err)
	//	}
	//}()

	m.Run()
	f.Shutdown()

}

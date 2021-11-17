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

package specs

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"net/http"
)

// ClusterKnowledge show cluster knowledge
// @Summary show cluster knowledge
// @Description show cluster knowledge
// @Tags knowledge
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} controller.CommonResult{data=[]knowledge.ClusterTypeSpec}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /knowledges/ [get]
func ClusterKnowledge(c *gin.Context) {
	var allSpec = new([]knowledge.ClusterTypeSpec)
	b, err := json.Marshal(knowledge.SpecKnowledge.Specs)

	if err !=nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, ""))
	} else {
		json.Unmarshal(b, allSpec)
		if allSpec != nil {
			for _, eachCluster := range *allSpec {
				for _, eachVersion := range eachCluster.VersionSpecs {
					for i := 0; i < len(eachVersion.ComponentSpecs); i++ {
						if eachVersion.ComponentSpecs[i].ComponentConstraint.Parasite == true {
							eachVersion.ComponentSpecs = append(eachVersion.ComponentSpecs[:i], eachVersion.ComponentSpecs[i+1:]...)
							i--
						}
					}
				}
			}
			c.JSON(http.StatusOK, controller.Success(knowledge.SpecKnowledge.Specs))
		}
	}

}

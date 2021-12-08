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

/*******************************************************************************
 * @File: request
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/3
*******************************************************************************/

package upgrade

type QueryUpgradePathReq struct {
	ClusterID string `json:"ClusterId"`
}

type Path struct {
	Type     string   `json:"type"`
	Versions []string `json:"versions"`
}

type QueryUpgradePathRsp struct {
	Paths []*Path
}

type QueryUpgradeVersionDiffInfoReq struct {
	ClusterID string `json:"ClusterId"`
	Version   string `json:"version"`
}

type QueryUpgradeVersionDiffInfoRsp struct {
	ConfigDiffInfo []struct {
		Name         string `json:"name"`
		InstanceType string `json:"instanceType"`
		CurrentVal   string `json:"currentVal"`
		SuggestVal   string `json:"suggestVal"`
		Range        string `json:"range"`
		Description  string `json:"description"`
	}
}

type ClusterUpgradeReq struct {
	ClusterID     string `json:"ClusterId"`
	TargetVersion string `json:"targetVersion"`
	Parameters    []struct {
		Name         string `json:"name"`
		InstanceType string `json:"instanceType"`
		Value        string `json:"value"`
	}
}

type ClusterUpgradeRsp struct {
	AsyncTaskWorkFlowInfo
}

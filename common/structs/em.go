/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 ******************************************************************************/

/*******************************************************************************
 * @File: em.go
 * @Description: em topology structs
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/7 14:56
*******************************************************************************/

package structs

type EMMetaTopo struct {
	ClusterMeta EMMetaInfo   `json:"cluster_meta"`
	Instances   []EMInstance `json:"instances"`
}

type EMMetaInfo struct {
	ClusterType    string `json:"cluster_type"`
	ClusterName    string `json:"cluster_name"`
	ClusterVersion string `json:"cluster_version"`
	DeployUser     string `json:"deploy_user"`
	SshType        string `json:"ssh_type"`
}

type EMInstance struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Host      string `json:"host"`
	Ports     string `json:"ports"`
	OsArch    string `json:"os_arch"`
	Status    string `json:"status"`
	Since     string `json:"since"`
	DataDir   string `json:"data_dir"`
	DeployDir string `json:"deploy_dir"`
}

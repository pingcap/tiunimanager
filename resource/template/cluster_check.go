/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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

package template

var EMClusterCheck = `
global:
  user: {{ .GlobalUser }}
  group: {{ .GlobalGroup }}
  ssh_port: {{ .GlobalSSHPort }}
  arch: {{ .GlobalArch }}
{{ if (len .TemplateItemsForCompute) }}
tidb_servers:
{{ range .TemplateItemsForCompute }}
  - host: {{ .HostIP }}
    deploy_dir: {{ .DeployDir }}/tidb_deploy
    port: {{ .Port1 }}
    status_port: {{ .Port2 }}
{{ end }}
{{ end }}
{{ if (len .TemplateItemsForStorage) }}
tikv_servers:
{{ range .TemplateItemsForStorage }}
  - host: {{ .HostIP }}
    data_dir: {{ .DataDir }}/tikv_data
    deploy_dir: {{ .DeployDir }}/tikv_deploy
    port: {{ .Port1 }}
    status_port: {{ .Port2 }}
{{ end }}
{{ end }}
{{ if (len .TemplateItemsForSchedule) }}
pd_servers:
{{ range .TemplateItemsForSchedule }}
  - host: {{ .HostIP }}
    data_dir: {{ .DataDir }}/pd_data
    deploy_dir: {{ .DeployDir }}/pd_deploy
    client_port: {{ .Port1 }}
    peer_port: {{ .Port2 }}
{{ end }}
{{ end }}
`

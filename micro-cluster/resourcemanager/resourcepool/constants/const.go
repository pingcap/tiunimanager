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
 ******************************************************************************/

package constants

const SecondPartyReady = true

const (
	FlowImportHosts            string = "ImportHosts"            // A normal flow to import hosts
	FlowImportHostsWithoutInit string = "ImportHostsWithoutInit" // import hosts without initialization
	FlowTakeOverHosts          string = "TakeOverHosts"          // A flow to take over hosts
	FlowDeleteHosts            string = "DeleteHosts"            // A normal flow to delete hosts
	FlowDeleteHostsByForce     string = "DeleteHostsByForce"     // delete hosts by force - without uninstall filebeat .etc.
)

const (
	ContextResourcePoolKey   string = "resourcePool"
	ContextHostInfoArrayKey  string = "hostInfoArray"
	ContextHostIDArrayKey    string = "hostIDArray"
	ContextWorkFlowNodeIDKey string = "resourceWorkFlowNodeID"
)

const (
	HostSSHPort       = 22
	HostFileBeatPort  = 0
	FileBeatDataDir   = "/tiem-data"
	FileBeatDeployDir = "/tiem-deploy"
)

const (
	DefaultTiupTimeOut      = 360
	DefaultCopySshIDTimeOut = 10
)

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

package adapt

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
)

type TiUPTiDBMetadataManager struct {
	componentParsers map[string]domain.ComponentParser
}

func NewTiUPTiDBMetadataManager() *TiUPTiDBMetadataManager {
	mgr := new(TiUPTiDBMetadataManager)
	mgr.componentParsers = map[string]domain.ComponentParser{}

	parser := TiDBComponentParser{}
	mgr.componentParsers[parser.GetComponent().ComponentType] = parser

	parser2 := TiKVComponentParser{}
	mgr.componentParsers[parser2.GetComponent().ComponentType] = parser2

	parser3 := PDComponentParser{}
	mgr.componentParsers[parser3.GetComponent().ComponentType] = parser3

	parser4 := TiFlashComponentParser{}
	mgr.componentParsers[parser4.GetComponent().ComponentType] = parser4

	return mgr
}

func (t TiUPTiDBMetadataManager) FetchFromRemoteCluster(ctx context.Context, req *clusterpb.ClusterTakeoverReqDTO) (spec.Metadata, error) {
	Conf := ssh.ClientConfig{User: req.TiupUserName,
		Auth: []ssh.AuthMethod{ssh.Password(req.TiupUserPassword)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	Client, err := ssh.Dial("tcp", net.JoinHostPort(req.TiupIp, req.Port), &Conf)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("FetchFromRemoteCluster, error: %s", err.Error())
		return nil, framework.WrapError(common.TIEM_TAKEOVER_NOT_REACHABLE, "ssh dial error", err)
	}
	defer Client.Close()

	sftpClient, err := sftp.NewClient(Client)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("FetchFromRemoteCluster, error: %s", err.Error())
		return nil, framework.WrapError(common.TIEM_TAKEOVER_INCORRECT_PATH, "new sftp client error", err)
	}
	defer sftpClient.Close()

	remoteFileName := fmt.Sprintf("%sstorage/cluster/clusters/%s/meta.yaml", req.TiupPath, req.ClusterNames[0])
	remoteFile, err := sftpClient.Open(remoteFileName)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("FetchFromRemoteCluster, error: %s", err.Error())
		return nil, framework.WrapError(common.TIEM_TAKEOVER_INCORRECT_PATH, "open sftp client error", err)
	}
	defer remoteFile.Close()

	dataByte, err := ioutil.ReadAll(remoteFile)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("FetchFromRemoteCluster, error: %s", err.Error())
		return nil, framework.WrapError(common.TIEM_TAKEOVER_INCORRECT_PATH, "read remote file error", err)
	}

	metadata := &spec.ClusterMeta{}
	err = yaml.Unmarshal(dataByte, metadata)

	return metadata, err
}

func (t TiUPTiDBMetadataManager) FetchFromLocal(ctx context.Context, tiupPath string, clusterName string) (spec.Metadata, error) {

	fileName := fmt.Sprintf("%sstorage/cluster/clusters/%s/meta.yaml", tiupPath, clusterName)
	file, err := os.Open(fileName)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("FetchFromLocal, error: %s", err.Error())
		return nil, framework.WrapError(common.TIEM_TAKEOVER_INCORRECT_PATH, "open file error", err)
	}
	defer file.Close()

	dataByte, err := ioutil.ReadAll(file)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("FetchFromLocal, error: %s", err.Error())
		return nil, framework.WrapError(common.TIEM_TAKEOVER_INCORRECT_PATH, "read file error", err)
	}

	metadata := &spec.ClusterMeta{}
	err = yaml.Unmarshal(dataByte, metadata)

	return metadata, err
}



func (t TiUPTiDBMetadataManager) RebuildMetadataFromComponents(ctx context.Context, cluster *domain.Cluster, components []*domain.ComponentGroup) (spec.Metadata, error) {
	panic("implement me")
}

func (t TiUPTiDBMetadataManager) ParseComponentsFromMetaData(ctx context.Context, metadata spec.Metadata) ([]*domain.ComponentGroup, error) {
	version := metadata.GetBaseMeta().Version

	clusterSpec := metadata.GetTopology().(*spec.Specification)
	componentsList := knowledge.GetComponentsForCluster("TiDB", version)

	componentGroups := make([]*domain.ComponentGroup, 0)
	for _, component := range componentsList {
		componentGroups = append(componentGroups, t.componentParsers[component.ComponentType].ParseComponent(clusterSpec))
	}

	return componentGroups, nil
}

func (t TiUPTiDBMetadataManager) ParseClusterInfoFromMetaData(ctx context.Context, meta spec.BaseMeta) (clusterType, user string, group string, version string) {
	return "TiDB", meta.User, meta.Group, meta.Version
}

type TiDBComponentParser struct{}

func (t TiDBComponentParser) GetComponent() *knowledge.ClusterComponent {
	return knowledge.ClusterComponentFromCode("TiDB")
}

func (t TiDBComponentParser) ParseComponent(spec *spec.Specification) *domain.ComponentGroup {
	group := &domain.ComponentGroup{
		ComponentType: t.GetComponent(),
		Nodes:         make([]*domain.ComponentInstance, 0),
	}

	for _, server := range spec.TiDBServers {
		componentInstance := &domain.ComponentInstance{
			ComponentType: t.GetComponent(),
			Host: server.Host,
			PortList: []int{server.Port, server.StatusPort},
		}
		group.Nodes = append(group.Nodes, componentInstance)
	}

	return group
}

type TiKVComponentParser struct{}

func (t TiKVComponentParser) GetComponent() *knowledge.ClusterComponent {
	return knowledge.ClusterComponentFromCode("TiKV")
}

func (t TiKVComponentParser) ParseComponent(spec *spec.Specification) *domain.ComponentGroup {
	group := &domain.ComponentGroup{
		ComponentType: t.GetComponent(),
		Nodes:         make([]*domain.ComponentInstance, 0),
	}

	for _, server := range spec.TiKVServers {
		componentInstance := &domain.ComponentInstance{
			ComponentType: t.GetComponent(),
			Host: server.Host,
			PortList: []int{server.Port, server.StatusPort},
		}
		group.Nodes = append(group.Nodes, componentInstance)
	}

	return group
}

type PDComponentParser struct{}

func (t PDComponentParser) GetComponent() *knowledge.ClusterComponent {
	return knowledge.ClusterComponentFromCode("PD")
}

func (t PDComponentParser) ParseComponent(spec *spec.Specification) *domain.ComponentGroup {
	group := &domain.ComponentGroup{
		ComponentType: t.GetComponent(),
		Nodes:         make([]*domain.ComponentInstance, 0),
	}

	for _, server := range spec.PDServers {
		componentInstance := &domain.ComponentInstance{
			ComponentType: t.GetComponent(),
			Host: server.Host,
			PortList: []int{server.ClientPort, server.PeerPort},
		}
		group.Nodes = append(group.Nodes, componentInstance)
	}

	return group
}

type TiFlashComponentParser struct{}

func (t TiFlashComponentParser) GetComponent() *knowledge.ClusterComponent {
	return knowledge.ClusterComponentFromCode("TiFlash")
}

func (t TiFlashComponentParser) ParseComponent(spec *spec.Specification) *domain.ComponentGroup {
	group := &domain.ComponentGroup{
		ComponentType: t.GetComponent(),
		Nodes:         make([]*domain.ComponentInstance, 0),
	}

	for _, server := range spec.TiFlashServers {
		componentInstance := &domain.ComponentInstance{
			ComponentType: t.GetComponent(),
			Host: server.Host,
			PortList: []int{server.TCPPort, server.HTTPPort, server.StatusPort, server.FlashProxyPort, server.FlashServicePort, server.FlashProxyStatusPort},
		}
		group.Nodes = append(group.Nodes, componentInstance)
	}

	return group
}

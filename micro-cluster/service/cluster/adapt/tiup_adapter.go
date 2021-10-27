package adapt

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
)

type TiUPMetadataManager struct {}

func (t TiUPMetadataManager) FetchFromRemoteCluster(ctx context.Context, req *clusterpb.ClusterTakeoverReqDTO) (spec.Metadata, error) {
	Conf := ssh.ClientConfig{User: req.TiupUserName,
		Auth: []ssh.AuthMethod{ssh.Password(req.TiupUserPassword)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	Client, err := ssh.Dial("tcp", net.JoinHostPort(req.TiupIp, req.Port), &Conf)
	if err != nil {
		return nil, err
	}
	defer Client.Close()

	sftpClient, err := sftp.NewClient(Client)
	if err != nil {
		return nil, err
	}
	defer sftpClient.Close()

	remoteFileName := fmt.Sprintf("%sstorage/cluster/clusters/%s/meta.yaml", req.TiupPath, req.ClusterNames[0])
	remoteFile, err := sftpClient.Open(remoteFileName)
	if err != nil {
		return nil, err
	}
	defer remoteFile.Close()
	if err != nil {
		return nil, err
	}
	dataByte, err := ioutil.ReadAll(remoteFile)
	if err != nil {
		if err != nil {
			return nil, err
		}
	}

	metadata := &spec.ClusterMeta{}
	yaml.Unmarshal(dataByte, metadata)

	return metadata, nil
}

func (t TiUPMetadataManager) RebuildMetadataFromComponents(cluster *domain.Cluster, components []domain.ComponentGroup) (spec.Metadata, error) {
	panic("implement me")
}

func (t TiUPMetadataManager) ParseComponentsFromMetaData(metadata spec.Metadata) ([]domain.ComponentGroup, error) {
	panic("implement me")
}

func (t TiUPMetadataManager) ParseClusterInfoFromMetaData(meta spec.BaseMeta) (user string, group string, version string) {
	return meta.User, meta.Group, meta.Version
}



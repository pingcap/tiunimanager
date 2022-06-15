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
 *                                                                            *
 ******************************************************************************/

package registry

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/pingcap/tiunimanager/library/framework"

	etcd "go.etcd.io/etcd/server/v3/embed"
)

const (
	namePrefix   string = "etcd"
	dirPrefix    string = "data_"
	localAddress string = "0.0.0.0"
	httpProtocol        = "https://"
)

type EmbedEtcdConfig struct {
	Name           string
	Index          int
	Dir            string
	ClientUrl      string
	PeerUrl        string
	EtcdClientUrls []string
	EtcdPeerUrls   []string
}

var log *logrus.Entry
var clientArgs *framework.ClientArgs

func InitEmbedEtcd(b *framework.BaseFramework) error {
	// init log and client args
	log = b.GetRootLogger().ForkFile(b.GetServiceMeta().ServiceName.ServerName())
	clientArgs = b.GetClientArgs()

	// parse client inject param config
	embedEtcdConfig, err := parseEtcdConfig()
	if err != nil {
		return err
	}
	// start embed etcd server
	return startEmbedEtcd(embedEtcdConfig)
}

// start embed etcd server
func startEmbedEtcd(embedEtcdConfig *EmbedEtcdConfig) error {
	cfg := etcd.NewConfig()
	// enable v2 http api
	cfg.EnableV2 = true
	cfg.Dir = embedEtcdConfig.Dir
	cfg.Name = embedEtcdConfig.Name

	cfg.ClientTLSInfo = framework.EtcdCert.ServerTLSInfo
	cfg.PeerTLSInfo = framework.EtcdCert.PeerTLSInfo
	cfg.StrictReconfigCheck = false
	log.Debugf("start embed etcd name: %s, dir: %s", cfg.Dir, cfg.Name)
	// advertise peer urls, e.g.: 192.168.1.101:2380,192.168.1.102:2380,192.168.1.102:2380
	cfg.APUrls = parsePeers([]string{embedEtcdConfig.PeerUrl})

	// listen peer urls, e.g.: 0.0.0.0:2380
	cfg.LPUrls = parsePeers([]string{localAddress + ":" + strings.Split(embedEtcdConfig.PeerUrl, ":")[1]})
	// advertise client urls, e.g.: 192.168.1.101:2379,192.168.1.102:2379,192.168.1.102:2379
	cfg.ACUrls = parseClients(embedEtcdConfig.EtcdClientUrls)
	// listen client urls, e.g.: 0.0.0.0:2379
	cfg.LCUrls = parseClients([]string{localAddress + ":" + strings.Split(embedEtcdConfig.ClientUrl, ":")[1]})

	cfg.InitialCluster = parseInitialCluster(embedEtcdConfig.EtcdPeerUrls)
	log.Debugf("initial LPUrls: %v, ACUrls: %v, LCUrls: %v, InitialCluster: %v:",
		cfg.LPUrls, cfg.ACUrls, cfg.LCUrls, cfg.InitialCluster)
	e, err := etcd.StartEtcd(cfg)
	if err != nil {
		return err
	}
	defer e.Close()

	select {
	case <-e.Server.ReadyNotify():
		log.Info("Server is ready!")
	case <-time.After(60 * time.Second):
		e.Server.Stop() // trigger a shutdown
		log.Warn("Server took too long to start!")
	}
	log.Fatal(<-e.Err())
	return err
}

// parse client inject param config
func parseEtcdConfig() (*EmbedEtcdConfig, error) {
	// Get injected client parameters
	endpoint := clientArgs.Host + ":" + strconv.Itoa(clientArgs.RegistryClientPort)
	log.Debugf("client injected registry client port: %d, peer port: %d, registry address: %s, endpoint: %s",
		clientArgs.RegistryClientPort, clientArgs.RegistryPeerPort, clientArgs.RegistryAddress, endpoint)

	etcdAddresses := strings.Split(clientArgs.RegistryAddress, ",")
	embedEtcdConfig := &EmbedEtcdConfig{
		EtcdClientUrls: make([]string, len(etcdAddresses)),
		EtcdPeerUrls:   make([]string, len(etcdAddresses)),
	}
	for i, addr := range etcdAddresses {
		embedEtcdConfig.EtcdClientUrls[i] = addr

		ipAndPort := strings.Split(addr, ":")
		if len(ipAndPort) != 2 {
			return nil, errors.New("registry address is invalid, address: " + addr)
		}
		clientPort, err := strconv.Atoi(ipAndPort[1])
		if err != nil {
			return nil, err
		}

		if strings.Contains(addr, endpoint) {
			// current host address
			embedEtcdConfig.Index = i
			embedEtcdConfig.ClientUrl = addr
			embedEtcdConfig.EtcdPeerUrls[i] = ipAndPort[0] + ":" + strconv.Itoa(clientArgs.RegistryPeerPort)
			embedEtcdConfig.PeerUrl = embedEtcdConfig.EtcdPeerUrls[i]
		} else {
			peerPort := clientPort + 1
			embedEtcdConfig.EtcdPeerUrls[i] = ipAndPort[0] + ":" + strconv.Itoa(peerPort)
			embedEtcdConfig.PeerUrl = embedEtcdConfig.EtcdPeerUrls[i]
		}
	}
	// set name
	embedEtcdConfig.Name = namePrefix + strconv.Itoa(embedEtcdConfig.Index)
	// set dir
	embedEtcdConfig.Dir = clientArgs.DataDir + "/" + namePrefix + "/" + dirPrefix + embedEtcdConfig.Name
	return embedEtcdConfig, nil
}

// parse peer urls by host address
func parsePeers(eps []string) []url.URL {
	urls := make([]url.URL, len(eps))
	for i, ep := range eps {
		u, err := url.Parse(httpProtocol + ep)
		if err != nil {
			return []url.URL{}
		}
		urls[i] = *u
	}
	return urls
}

// parse client urls by host address
func parseClients(eps []string) []url.URL {
	urls := make([]url.URL, len(eps))
	for i, ep := range eps {
		u, err := url.Parse(httpProtocol + ep)
		if err != nil {
			return []url.URL{}
		}
		urls[i] = *u
	}
	return urls
}

// parse initial cluster urls by host address
func parseInitialCluster(eps []string) string {
	urls := ""
	for i, ep := range eps {
		urls += namePrefix + strconv.Itoa(i) + "=" + httpProtocol + ep
		if i+1 < len(eps) {
			urls += ","
		}
	}
	return urls
}
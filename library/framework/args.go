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

package framework

import (
	"github.com/micro/cli/v2"
)

// ClientArgs Client startup parameter structure
//  Host Host ip address. e.g.: 127.0.0.1
//  EnableHttps enable https for openapi, default false
//  SkipHostInit Skip initialize host when importing, default false
//  IgnoreHostWarns Ignore host warnings when importing, default false
//  Port Micro service management port.
//  MetricsPort Monitoring port exposed by the service, default by 4121.
//  RestPort Restful api port. micro-api specific args.
//  RegistryClientPort Registry center client externally exposed port.
//  RegistryPeerPort Registry internal communication port.
//	RegistryAddress Registry center and config center address, default registry is etcd.
//  TracerAddress Opentracing jaeger server address.
//  DeployDir The binary and configuration files deploy dir
//  DataDir Persistent data storage directory
//  LogLevel Minimum log level, default by info.
type ClientArgs struct {
	Host                 string
	EnableHttps          bool
	SkipHostInit         bool
	IgnoreHostWarns      bool
	Port                 int
	MetricsPort          int
	RegistryClientPort   int
	RegistryPeerPort     int
	RegistryAddress      string
	TracerAddress        string
	DeployDir            string
	DataDir              string
	LogLevel             string
	ElasticsearchAddress string
	EMClusterName        string
	EMVersion            string
	DeployUser           string
}

func AllFlags(receiver *ClientArgs) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "host",
			Value:       "127.0.0.1",
			Usage:       "Specify the host ip address.",
			Destination: &receiver.Host,
		},
		&cli.BoolFlag{
			Name:        "enable-https",
			Value:       false,
			Usage:       "Enable https for open-api.",
			Destination: &receiver.EnableHttps,
		},
		&cli.BoolFlag{
			Name:        "skip-host-init",
			Value:       false,
			Usage:       "Skip initialize host when importing.",
			Destination: &receiver.SkipHostInit,
		},
		&cli.BoolFlag{
			Name:        "ignore-host-warns",
			Value:       false,
			Usage:       "Ignore host warnings when importing.",
			Destination: &receiver.IgnoreHostWarns,
		},
		&cli.IntFlag{
			Name:        "port",
			Usage:       "Specify the micro service management port.",
			Destination: &receiver.Port,
		},
		&cli.IntFlag{
			Name:        "metrics-port",
			Value:       4103,
			Usage:       "Specify the metrics port exposed by the service.",
			Destination: &receiver.MetricsPort,
		},
		&cli.IntFlag{
			Name:        "registry-client-port",
			Value:       4106,
			Usage:       "Specify the default etcd registry client port.",
			Destination: &receiver.RegistryClientPort,
		},
		&cli.IntFlag{
			Name:        "registry-peer-port",
			Value:       4107,
			Usage:       "Specify the default etcd registry internal communication port.",
			Destination: &receiver.RegistryPeerPort,
		},
		&cli.StringFlag{
			Name: "registry-address",
			// For convenience, set the default value after the embedded etcd is completed
			Value:       "127.0.0.1:4106",
			Usage:       "Specify the default etcd registry address.",
			Destination: &receiver.RegistryAddress,
		},
		&cli.StringFlag{
			Name:        "tracer-address",
			Value:       "127.0.0.1:4114",
			Usage:       "Specify the opentracing jaeger server address.",
			Destination: &receiver.TracerAddress,
		},
		&cli.StringFlag{
			Name:        "deploy-dir",
			Value:       "bin",
			Usage:       "Specify the binary and configuration files deploy dir.",
			Destination: &receiver.DeployDir,
		},
		&cli.StringFlag{
			Name:        "data-dir",
			Value:       ".",
			Usage:       "Specify the persistent data storage directory.",
			Destination: &receiver.DataDir,
		},
		&cli.StringFlag{
			Name:        "log-level",
			Value:       "info",
			Usage:       "Specify the minimum log level.",
			Destination: &receiver.LogLevel,
		},
		&cli.StringFlag{
			Name:        "elasticsearch-address",
			Value:       "127.0.0.1:4108",
			Usage:       "Specify the default elasticsearch address.",
			Destination: &receiver.ElasticsearchAddress,
		},
		&cli.StringFlag{
			Name:        "em-cluster-name",
			Value:       "",
			Usage:       "Specify the EM cluster name.",
			Destination: &receiver.EMClusterName,
		},
		&cli.StringFlag{
			Name:        "em-version",
			Value:       "",
			Usage:       "Specify the EM version.",
			Destination: &receiver.EMVersion,
		},
		&cli.StringFlag{
			Name:        "deploy-user",
			Value:       "",
			Usage:       "Specify the EM deploy user.",
			Destination: &receiver.DeployUser,
		},
	}
}

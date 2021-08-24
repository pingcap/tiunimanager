package config

import (
	"github.com/asim/go-micro/v3"
	"github.com/micro/cli/v2"
)

// ClientArgs Client startup parameter structure
//  Host Host ip address. e.g.: 127.0.0.1
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
	Host               string
	Port               int
	MetricsPort        int
	RestPort           int
	RegistryClientPort int
	RegistryPeerPort   int
	RegistryAddress    string
	TracerAddress      string
	DeployDir          string
	DataDir            string
	LogLevel           string
}

var clientArgs = new(ClientArgs)

func GetClientArgs() *ClientArgs {
	return clientArgs
}

var (
	hostFlag = &cli.StringFlag{
		Name:        "host",
		Value:       "127.0.0.1",
		Usage:       "Specify the host ip address.",
		Destination: &clientArgs.Host,
	}
	portFlag = &cli.IntFlag{
		Name:        "port",
		Usage:       "Specify the micro service management port.",
		Destination: &clientArgs.Port,
	}
	restPortFlag = &cli.IntFlag{
		Name:        "rest-port",
		Value:       4116,
		Usage:       "Specify the rest port of the instance.",
		Destination: &clientArgs.RestPort,
	}
	metricsPortFlag = &cli.IntFlag{
		Name:        "metrics-port",
		Value:       4121,
		Usage:       "Specify the metrics port exposed by the service.",
		Destination: &clientArgs.MetricsPort,
	}
	registryClientPortFlag = &cli.IntFlag{
		Name:        "registry-client-port",
		Value:       4101,
		Usage:       "Specify the default etcd registry client port.",
		Destination: &clientArgs.RegistryClientPort,
	}
	registryPeerPortFlag = &cli.IntFlag{
		Name:        "registry-peer-port",
		Value:       4102,
		Usage:       "Specify the default etcd registry internal communication port.",
		Destination: &clientArgs.RegistryPeerPort,
	}
	registryAddressFlag = &cli.StringFlag{
		Name: "registry-address",
		// For convenience, set the default value after the embedded etcd is completed
		//Value:       "127.0.0.1:4101",
		Usage:       "Specify the default etcd registry address.",
		Destination: &clientArgs.RegistryAddress,
	}
	tracerAddressFlag = &cli.StringFlag{
		Name:        "tracer-address",
		Value:       "",
		Usage:       "Specify the opentracing jaeger server address.",
		Destination: &clientArgs.TracerAddress,
	}
	deployDirFlag = &cli.StringFlag{
		Name:        "deploy-dir",
		Value:       "",
		Usage:       "Specify the binary and configuration files deploy dir.",
		Destination: &clientArgs.DeployDir,
	}
	dataDirFlag = &cli.StringFlag{
		Name:        "data-dir",
		Value:       "",
		Usage:       "Specify the persistent data storage directory.",
		Destination: &clientArgs.DataDir,
	}
	logLevelFlag = &cli.StringFlag{
		Name:        "log-level",
		Value:       "info",
		Usage:       "Specify the minimum log level.",
		Destination: &clientArgs.LogLevel,
	}
)

// GetMicroMetaDBCliArgsOption get micro-metadb client args
func GetMicroMetaDBCliArgsOption() micro.Option {
	return micro.Flags(
		hostFlag,
		portFlag,
		metricsPortFlag,
		registryClientPortFlag,
		registryPeerPortFlag,
		registryAddressFlag,
		tracerAddressFlag,
		deployDirFlag,
		dataDirFlag,
		logLevelFlag,
	)
}

// GetMicroClusterCliArgsOption get micro-cluster client args
func GetMicroClusterCliArgsOption() micro.Option {
	return micro.Flags(
		hostFlag,
		portFlag,
		metricsPortFlag,
		registryAddressFlag,
		tracerAddressFlag,
		deployDirFlag,
		dataDirFlag,
		logLevelFlag,
	)
}

// GetMicroApiCliArgsOption get micro-api client args
func GetMicroApiCliArgsOption() micro.Option {
	return micro.Flags(
		hostFlag,
		portFlag,
		restPortFlag,
		metricsPortFlag,
		registryAddressFlag,
		tracerAddressFlag,
		deployDirFlag,
		dataDirFlag,
		logLevelFlag,
	)
}


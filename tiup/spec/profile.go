// Copyright 2021 PingCAP
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/pingcap/errors"
	tiuplocaldata "github.com/pingcap/tiup/pkg/localdata"
	"github.com/pingcap/tiup/pkg/utils"
)

// sub directory names
const (
	TiUPPackageCacheDir = "packages"
	TiUPClusterDir      = "clusters"
	TiUPAuditDir        = "audit"
	TLSCertKeyDir       = "tls"
	TLSCACert           = "ca.crt"
	TLSCAKey            = "ca.pem"
	TLSClientCert       = "client.crt"
	TLSClientKey        = "client.pem"
	PFXClientCert       = "client.pfx"
)

var profileDir string

// getHomeDir get the home directory of current user (if they have one).
// The result path might be empty.
func getHomeDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", errors.Trace(err)
	}
	return u.HomeDir, nil
}

var initialized = false

// Initialize initializes the global variables of meta package. If the
// environment variable TIUP_COMPONENT_DATA_DIR is set, it is used as root of
// the profile directory, otherwise the `$HOME/.tiup` of current user is used.
// The directory will be created before return if it does not already exist.
func Initialize(base string) error {
	tiupData := os.Getenv(tiuplocaldata.EnvNameComponentDataDir)
	if tiupData == "" {
		homeDir, err := getHomeDir()
		if err != nil {
			return errors.Trace(err)
		}
		profileHome := homeDir + "/.tiup"
		if tiupHome := os.Getenv(tiuplocaldata.EnvNameHome); tiupHome != "" {
			profileHome = strings.TrimRight(tiupHome, "/")
		}
		profileDir = path.Join(profileHome, tiuplocaldata.StorageParentDir, base)
	} else {
		profileDir = tiupData
	}

	clusterBaseDir := filepath.Join(profileDir, TiUPClusterDir)
	specManager = NewSpec(clusterBaseDir, func() Metadata {
		return &EMMeta{
			Topology: new(Specification),
		}
	})
	initialized = true
	// make sure the dir exist
	return utils.CreateDir(profileDir)
}

// ProfileDir returns the full profile directory path of TiUP.
func ProfileDir() string {
	return profileDir
}

// ProfilePath joins a path under the profile dir
func ProfilePath(subpath ...string) string {
	return path.Join(append([]string{profileDir}, subpath...)...)
}

// ClusterPath returns the full path to a subpath (file or directory) of a
// cluster, it is a subdir in the profile dir of the user, with the cluster name
// as its name.
// It is not guaranteed the path already exist.
func ClusterPath(cluster string, subpath ...string) string {
	return GetSpecManager().Path(cluster, subpath...)
}

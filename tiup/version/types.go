// Copyright 2021 PingCAP, Inc.
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

package version

import (
	"fmt"
	"runtime"

	tiupver "github.com/pingcap/tiup/pkg/version"
)

// ComponentVersion is the semver of Component
type ComponentVersion struct {
	major int
	minor int
	patch int
	name  string
	tiup  string
}

// NewComponentVersion creates a ComponentVersion object
func NewComponentVersion() *ComponentVersion {
	return &ComponentVersion{
		major: ComponentVerMajor,
		minor: ComponentVerMinor,
		patch: ComponentVerPatch,
		name:  ComponentVerName,
		tiup:  tiupver.NewTiUPVersion().SemVer(),
	}
}

// Name returns the alternave name of ComponentVersion
func (v *ComponentVersion) Name() string {
	return v.name
}

// SemVer returns ComponentVersion in semver format
func (v *ComponentVersion) SemVer() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

// String converts ComponentVersion to a string
func (v *ComponentVersion) String() string {
	return fmt.Sprintf("%s %s (feat. TiUP %s)\n%s", v.SemVer(), v.name, v.tiup, NewComponentBuildInfo())
}

// ComponentBuild is the info of building environment
type ComponentBuild struct {
	GitHash   string `json:"gitHash"`
	GitRef    string `json:"gitRef"`
	GoVersion string `json:"goVersion"`
}

// NewComponentBuildInfo creates a ComponentBuild object
func NewComponentBuildInfo() *ComponentBuild {
	return &ComponentBuild{
		GitHash:   GitHash,
		GitRef:    GitRef,
		GoVersion: runtime.Version(),
	}
}

// String converts ComponentBuild to a string
func (v *ComponentBuild) String() string {
	return fmt.Sprintf("Go Version: %s\nGit Ref: %s\nGitHash: %s", v.GoVersion, v.GitRef, v.GitHash)
}

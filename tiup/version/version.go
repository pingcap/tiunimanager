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

// This file only contains version related variables and consts, all
// type definitions and functions shall not be implemented here.
// This file is excluded from CI tests.

package version

var (
	// ComponentVerMajor is the major version of Component
	ComponentVerMajor = 1
	// ComponentVerMinor is the minor version of Component
	ComponentVerMinor = 0
	// ComponentVerPatch is the patch version of Component
	ComponentVerPatch = 0
	// ComponentVerName is an alternative name of the version
	ComponentVerName = "em"
	// GitHash is the current git commit hash
	GitHash = "Unknown"
	// GitRef is the current git reference name (branch or tag)
	GitRef = "Unknown"
)

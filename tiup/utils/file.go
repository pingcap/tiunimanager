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

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

func ReadFile(filePath string) ([]byte, error) {
	if !IsFile(filePath) {
		return nil, fmt.Errorf("the file path '%s' is a directory, not a file", filePath)
	}
	b, e := ioutil.ReadFile(filePath)
	if e != nil {
		return nil, e
	}
	return b, nil
}

func IsFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

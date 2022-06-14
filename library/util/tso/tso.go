// Copyright 2019 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tso

import (
	"time"
)

const (
	physicalShiftBits = 18
	logicalBits       = (1 << physicalShiftBits) - 1
)

// ParseTS parses the ts to (physical,logical).
func ParseTS(ts uint64) (time.Time, uint64) {
	logical := ts & logicalBits
	physical := ts >> physicalShiftBits
	physicalTime := time.Unix(int64(physical/1000), int64(physical)%1000*time.Millisecond.Nanoseconds())
	return physicalTime, logical
}

// ComposeTS generate an `uint64` TS by passing the physical and logical parts.
func ComposeTS(physical, logical int64) uint64 {
	return uint64(physical)<<18 | uint64(logical)&0x3FFFF
}

// GenerateTSO generate a TSO by passing `time.Time` and `uint64`
func GenerateTSO(physical time.Time, logical uint64) uint64 {
	return ComposeTS(physical.UnixNano()/int64(time.Millisecond), int64(logical))
}

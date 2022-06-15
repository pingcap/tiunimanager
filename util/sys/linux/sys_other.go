// +build !linux

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

package linux

import (
	"context"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"runtime"
)

// OSVersion returns version info of operating system.
// for non-linux system will only return os and arch info.
func OSVersion() (osVersion string, err error) {
	osVersion = runtime.GOOS + "." + runtime.GOARCH
	return
}

//OSInfo returns version info of operating system
func OSInfo(ctx context.Context) (platform, family, version string, err error) {
	platform, family, version, err = host.PlatformInformationWithContext(ctx)
	return
}

// SetAffinity sets cpu affinity.
func SetAffinity(cpus []int) error {
	return nil
}

//GetLoadavg get current host loadavg
func GetLoadavg(ctx context.Context) (load5, load15 float64, err error) {
	l, er := load.AvgWithContext(ctx)
	if er != nil {
		return 0.0, 0.0, er
	}
	return l.Load5, l.Load15, nil
}

//GetCpuInfo get current host cpus information
//go's reported runtime.NUMCPU()
//number of cpus reported cores for first cpu
//reported model name e.g. `Intel(R) Core(TM) i7-7920HQ CPU @ 3.10GHz`
//speed of first cpu e.g. 3100
func GetCpuInfo(ctx context.Context) (num, sockets int, cores int32, model string, hz float64, er error) {
	cpus, err := cpu.InfoWithContext(ctx)
	if err != nil || len(cpus) <= 0 {
		return runtime.NumCPU(), 0, 0, "", 0.0, err
	}
	return runtime.NumCPU(), len(cpus), cpus[0].Cores, cpus[0].ModelName, cpus[0].Mhz, nil
}

//GetVirtualMemory get current host virtual memory information
func GetVirtualMemory(ctx context.Context) (total, available uint64, err error) {
	m, er := mem.VirtualMemory()
	if er != nil {
		return 0.0, 0.0, err
	}
	return m.Total, m.Available, nil
}

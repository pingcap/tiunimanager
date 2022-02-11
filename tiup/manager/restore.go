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

package manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/pingcap-inc/tiem/tiup/spec"

	"github.com/joomcode/errorx"
	operator "github.com/pingcap-inc/tiem/tiup/operation"
	"github.com/pingcap-inc/tiem/tiup/task"
	perrs "github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/cluster/clusterutil"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/set"
)

// RestoreOptions for exec shell commanm.
type RestoreOptions struct {
	Source string
	Limit  int // rate limit in Kbit/s
}

// Backup copies files from or to host in the tidb cluster.
func (m *Manager) Restore(name string, opt RestoreOptions, gOpt operator.Options) error {
	if err := clusterutil.ValidateClusterNameOrError(name); err != nil {
		return err
	}

	metadata, err := m.meta(name)
	if err != nil {
		return err
	}

	topo := metadata.GetTopology()
	base := metadata.GetBaseMeta()

	filterRoles := set.NewStringSet(gOpt.Roles...)
	filterNodes := set.NewStringSet(gOpt.Nodes...)

	var shellTasks []task.Task

	uniqueHosts := map[string]set.StringSet{} // host-sshPort -> {target-path}
	topo.IterInstance(func(inst spec.Instance) {
		key := fmt.Sprintf("%s-%d", inst.GetHost(), inst.GetSSHPort())
		if inst.ComponentName() == spec.ComponentTiEMClusterServer {
			dstPath := inst.DataDir() + "/" + spec.DBName // e.g.: /tiem-data/cluster-server-4101/em.db
			uniqueHosts[key] = set.NewStringSet(dstPath)
		}
		if _, found := uniqueHosts[key]; !found {
			if len(gOpt.Roles) > 0 && !filterRoles.Exist(inst.Role()) {
				return
			}
			if len(gOpt.Nodes) > 0 && !filterNodes.Exist(inst.GetHost()) {
				return
			}
		}
	})

	for hostKey, i := range uniqueHosts {
		host := strings.Split(hostKey, "-")[0]
		for _, p := range i.Slice() {
			t := task.NewBuilder()
			t.CopyFile(opt.Source, p, host, false, opt.Limit)
			shellTasks = append(shellTasks, t.Build())
		}
	}

	b, err := m.sshTaskBuilder(name, topo, base.User, gOpt)
	if err != nil {
		return err
	}

	t := b.
		Parallel(false, shellTasks...).
		Build()

	execCtx := ctxt.New(context.Background(), gOpt.Concurrency)
	if err := t.Execute(execCtx); err != nil {
		if errorx.Cast(err) != nil {
			// FIXME: Map possible task errors and give suggestions.
			return err
		}
		return perrs.Trace(err)
	}

	return nil
}

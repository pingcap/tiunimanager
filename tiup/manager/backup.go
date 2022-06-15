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

package manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/pingcap/tiunimanager/tiup/spec"

	"github.com/joomcode/errorx"
	perrs "github.com/pingcap/errors"
	operator "github.com/pingcap/tiunimanager/tiup/operation"
	"github.com/pingcap/tiunimanager/tiup/task"
	"github.com/pingcap/tiup/pkg/cluster/clusterutil"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/logger/log"
	"github.com/pingcap/tiup/pkg/set"
)

// BackupOptions for exec shell commanm.
type BackupOptions struct {
	Target string
	Limit  int // rate limit in Kbit/s
}

// Backup copies files from or to host in the tidb cluster.
func (m *Manager) Backup(name string, opt BackupOptions, gOpt operator.Options) error {
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

	var srcPath string
	uniqueHosts := map[string]set.StringSet{} // host-sshPort -> {target-path}
	topo.IterInstance(func(inst spec.Instance) {
		if inst.ComponentName() == spec.ComponentTiUniManagerClusterServer {
			srcPath = inst.DataDir() + "/" + spec.DBName // e.g.: /tiunimanager-data/cluster-server-4101/em.db
		}
		key := fmt.Sprintf("%s-%d", inst.GetHost(), inst.GetSSHPort())
		if _, found := uniqueHosts[key]; !found {
			if len(gOpt.Roles) > 0 && !filterRoles.Exist(inst.Role()) {
				return
			}
			if len(gOpt.Nodes) > 0 && !filterNodes.Exist(inst.GetHost()) {
				return
			}

			// render remote path
			instPath := opt.Target
			paths, err := renderInstanceSpec(instPath, inst)
			if err != nil {
				log.Debugf("error rendering remote path with spec: %s", err)
				return // skip
			}
			pathSet := set.NewStringSet(paths...)
			if _, ok := uniqueHosts[key]; ok {
				uniqueHosts[key].Join(pathSet)
				return
			}
			uniqueHosts[key] = pathSet
		}
	})

	for hostKey, i := range uniqueHosts {
		host := strings.Split(hostKey, "-")[0]
		for _, p := range i.Slice() {
			t := task.NewBuilder()
			t.CopyFile(srcPath, p+"/"+spec.DBName, host, false, opt.Limit)
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

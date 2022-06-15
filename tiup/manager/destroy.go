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
	"crypto/tls"
	"errors"

	"github.com/fatih/color"
	"github.com/joomcode/errorx"
	perrs "github.com/pingcap/errors"
	operator "github.com/pingcap/tiunimanager/tiup/operation"
	"github.com/pingcap/tiup/pkg/cluster/clusterutil"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/logger/log"
	"github.com/pingcap/tiup/pkg/meta"
	"github.com/pingcap/tiup/pkg/tui"
)

// DestroyCluster destroy the cluster.
func (m *Manager) DestroyCluster(name string, gOpt operator.Options, destroyOpt operator.Options, skipConfirm bool) error {
	if err := clusterutil.ValidateClusterNameOrError(name); err != nil {
		return err
	}

	metadata, err := m.meta(name)
	if err != nil && !errors.Is(perrs.Cause(err), meta.ErrValidate) {
		return err
	}

	topo := metadata.GetTopology()
	base := metadata.GetBaseMeta()

	tlsCfg := &tls.Config{} // not implemented for tiunimanager

	if !skipConfirm {
		if err := tui.PromptForConfirmOrAbortError(
			"This operation will destroy %s %s cluster %s and its data.\nDo you want to continue? [y/N]:",
			m.sysName,
			color.HiYellowString(base.Version),
			color.HiYellowString(name)); err != nil {
			return err
		}
		log.Infof("Destroying cluster...")
	}

	b, err := m.sshTaskBuilder(name, topo, base.User, gOpt)
	if err != nil {
		return err
	}
	t := b.
		Func("StopCluster", func(ctx context.Context) error {
			return operator.Stop(ctx, topo, operator.Options{Force: destroyOpt.Force}, tlsCfg)
		}).
		Func("DestroyCluster", func(ctx context.Context) error {
			return operator.Destroy(ctx, topo, destroyOpt)
		}).
		Build()

	if err := t.Execute(ctxt.New(context.Background(), gOpt.Concurrency)); err != nil {
		if errorx.Cast(err) != nil {
			// FIXME: Map possible task errors and give suggestions.
			return err
		}
		return perrs.Trace(err)
	}

	if err := m.specManager.Remove(name); err != nil {
		return perrs.Trace(err)
	}

	log.Infof("Destroyed cluster `%s` successfully", name)
	return nil
}

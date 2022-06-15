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

	"github.com/fatih/color"
	"github.com/joomcode/errorx"
	perrs "github.com/pingcap/errors"
	operator "github.com/pingcap/tiunimanager/tiup/operation"
	"github.com/pingcap/tiunimanager/tiup/spec"
	"github.com/pingcap/tiunimanager/tiup/task"
	"github.com/pingcap/tiup/pkg/cluster/clusterutil"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/cluster/executor"
	"github.com/pingcap/tiup/pkg/logger/log"
	"github.com/pingcap/tiup/pkg/set"
	"github.com/pingcap/tiup/pkg/tui"
	"github.com/pingcap/tiup/pkg/utils"
	"gopkg.in/yaml.v3"
)

// ScaleOut scale out the cluster.
func (m *Manager) ScaleOut(
	name string,
	topoFile string,
	afterDeploy func(b *task.Builder, newPart spec.Topology),
	final func(b *task.Builder, name string, meta spec.Metadata),
	opt DeployOptions,
	skipConfirm bool,
	gOpt operator.Options,
) error {
	if err := clusterutil.ValidateClusterNameOrError(name); err != nil {
		return err
	}

	// check for the input topology to let user confirm if there're any
	// global configs set
	if err := checkForGlobalConfigs(topoFile); err != nil {
		return err
	}

	metadata, err := m.meta(name)
	// allow specific validation errors so that user can recover a broken
	// cluster if it is somehow in a bad state.
	if err != nil {
		return err
	}

	topo := metadata.GetTopology()
	base := metadata.GetBaseMeta()
	// Inherit existing global configuration. We must assign the inherited values before unmarshalling
	// because some default value rely on the global options and monitored options.
	newPart := topo.NewPart()

	// The no tispark master error is ignored, as if the tispark master is removed from the topology
	// file for some reason (manual edit, for example), it is still possible to scale-out it to make
	// the whole topology back to normal state.
	if err := spec.ParseTopologyYaml(topoFile, newPart); err != nil {
		return err
	}

	var (
		sshConnProps  *tui.SSHConnectionProps = &tui.SSHConnectionProps{}
		sshProxyProps *tui.SSHConnectionProps = &tui.SSHConnectionProps{}
	)
	if gOpt.SSHType != executor.SSHTypeNone {
		var err error
		if sshConnProps, err = tui.ReadIdentityFileOrPassword(opt.IdentityFile, opt.UsePassword); err != nil {
			return err
		}
		if len(gOpt.SSHProxyHost) != 0 {
			if sshProxyProps, err = tui.ReadIdentityFileOrPassword(gOpt.SSHProxyIdentity, gOpt.SSHProxyUsePassword); err != nil {
				return err
			}
		}
	}

	if err := m.fillHostArch(sshConnProps, sshProxyProps, newPart, &gOpt, opt.User); err != nil {
		return err
	}

	// Abort scale out operation if the merged topology is invalid
	mergedTopo := topo.MergeTopo(newPart)
	if err := mergedTopo.Validate(); err != nil {
		return err
	}
	spec.ExpandRelativeDir(mergedTopo)

	patchedComponents := set.NewStringSet()
	newPart.IterInstance(func(instance spec.Instance) {
		if utils.IsExist(m.specManager.Path(name, spec.PatchDirName, instance.ComponentName()+".tar.gz")) {
			patchedComponents.Insert(instance.ComponentName())
			instance.SetPatched(true)
		}
	})

	if !skipConfirm {
		// patchedComponents are components that have been patched and overwrited
		if err := m.confirmTopology(name, base.Version, newPart, patchedComponents); err != nil {
			return err
		}
	}

	// Build the scale out tasks
	t, err := buildScaleOutTask(
		m, name, metadata, mergedTopo, opt, sshConnProps, sshProxyProps, newPart,
		patchedComponents, gOpt, afterDeploy, final)
	if err != nil {
		return err
	}

	ctx := ctxt.New(context.Background(), gOpt.Concurrency)
	ctx = context.WithValue(ctx, ctxt.CtxBaseTopo, topo)
	if err := t.Execute(ctx); err != nil {
		if errorx.Cast(err) != nil {
			// FIXME: Map possible task errors and give suggestions.
			return err
		}
		return perrs.Trace(err)
	}

	log.Infof("Scaled cluster `%s` out successfully", name)

	return nil
}

// checkForGlobalConfigs checks the input scale out topology to make sure users are aware
// of the global config fields in it will be ignored.
func checkForGlobalConfigs(topoFile string) error {
	yamlFile, err := spec.ReadYamlFile(topoFile)
	if err != nil {
		return err
	}

	var newPart map[string]interface{}
	if err := yaml.Unmarshal(yamlFile, &newPart); err != nil {
		return err
	}

	for k := range newPart {
		switch k {
		case "global",
			"monitored",
			"server_configs":
			log.Warnf(color.YellowString(
				`You have one or more of ["global", "monitored", "server_configs"] fields configured in
the scale out topology, but they will be ignored during the scaling out process.
If you want to use configs different from the existing cluster, cancel now and
set them in the specification fileds for each host.`))
			if err := tui.PromptForConfirmOrAbortError("Do you want to continue? [y/N]: "); err != nil {
				return err
			}
			return nil // user confirmed, skip further checks
		}
	}
	return nil
}

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
	"crypto/tls"
	"fmt"
	"path/filepath"
	"strings"

	operator "github.com/pingcap-inc/tiem/tiup/operation"
	"github.com/pingcap-inc/tiem/tiup/spec"
	"github.com/pingcap-inc/tiem/tiup/task"
	"github.com/pingcap-inc/tiem/tiup/templates/config"
	cspec "github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/pingcap/tiup/pkg/meta"
	"github.com/pingcap/tiup/pkg/set"
	"github.com/pingcap/tiup/pkg/tui"
)

// buildReloadPromTasks reloads Prometheus configuration
func buildReloadPromTasks(topo spec.Topology, nodes ...string) []task.Task {
	monitor := spec.FindComponent(topo, spec.ComponentPrometheus)
	if monitor == nil {
		return nil
	}
	instances := monitor.Instances()
	if len(instances) == 0 {
		return nil
	}
	var tasks []task.Task
	deletedNodes := set.NewStringSet(nodes...)
	for _, inst := range monitor.Instances() {
		if deletedNodes.Exist(inst.ID()) {
			continue
		}
		t := task.NewBuilder().
			SystemCtl(inst.GetHost(), inst.ServiceName(), "reload", true).
			Build()
		tasks = append(tasks, t)
	}
	return tasks
}

func buildScaleOutTask(
	m *Manager,
	name string,
	metadata spec.Metadata,
	mergedTopo spec.Topology,
	opt DeployOptions,
	s, p *tui.SSHConnectionProps,
	newPart spec.Topology,
	patchedComponents set.StringSet,
	gOpt operator.Options,
	afterDeploy func(b *task.Builder, newPart spec.Topology),
	final func(b *task.Builder, name string, meta spec.Metadata),
) (task.Task, error) {
	var (
		envInitTasks       []task.Task // tasks which are used to initialize environment
		downloadCompTasks  []task.Task // tasks which are used to download components
		deployCompTasks    []task.Task // tasks which are used to copy components to remote host
		refreshConfigTasks []task.Task // tasks which are used to refresh configuration
	)

	topo := metadata.GetTopology()
	base := metadata.GetBaseMeta()
	specManager := m.specManager

	tlsCfg := &tls.Config{} // not implemented for tiem

	// Initialize the environments
	initializedHosts := set.NewStringSet()
	metadata.GetTopology().IterInstance(func(instance spec.Instance) {
		initializedHosts.Insert(instance.GetHost())
	})
	// uninitializedHosts are hosts which haven't been initialized yet
	uninitializedHosts := make(map[string]hostInfo) // host -> ssh-port, os, arch
	newPart.IterInstance(func(instance spec.Instance) {
		host := instance.GetHost()
		if initializedHosts.Exist(host) {
			return
		}
		if _, found := uninitializedHosts[host]; found {
			return
		}

		uninitializedHosts[host] = hostInfo{
			ssh:  instance.GetSSHPort(),
			os:   instance.OS(),
			arch: instance.Arch(),
		}

		var dirs []string
		globalOptions := metadata.GetTopology().BaseTopo().GlobalOptions
		for _, dir := range []string{globalOptions.DeployDir, globalOptions.DataDir, globalOptions.LogDir} {
			for _, dirname := range strings.Split(dir, ",") {
				if dirname == "" {
					continue
				}
				dirs = append(dirs, spec.Abs(globalOptions.User, dirname))
			}
		}
		t := task.NewBuilder().
			RootSSH(
				instance.GetHost(),
				instance.GetSSHPort(),
				opt.User,
				s.Password,
				s.IdentityFile,
				s.IdentityFilePassphrase,
				gOpt.SSHTimeout,
				gOpt.OptTimeout,
				gOpt.SSHProxyHost,
				gOpt.SSHProxyPort,
				gOpt.SSHProxyUser,
				p.Password,
				p.IdentityFile,
				p.IdentityFilePassphrase,
				gOpt.SSHProxyTimeout,
				gOpt.SSHType,
				globalOptions.SSHType,
			).
			EnvInit(instance.GetHost(), base.User, base.Group, opt.SkipCreateUser || globalOptions.User == opt.User).
			Mkdir(globalOptions.User, instance.GetHost(), dirs...).
			Build()
		envInitTasks = append(envInitTasks, t)
	})

	// Download missing component
	downloadCompTasks = convertStepDisplaysToTasks(buildDownloadCompTasks(base.Version, newPart, m.bindVersion))

	sshType := topo.BaseTopo().GlobalOptions.SSHType

	var iterErr error
	// Deploy the new topology and refresh the configuration
	newPart.IterInstance(func(inst spec.Instance) {
		version := m.bindVersion(inst.ComponentName(), base.Version)
		deployDir := spec.Abs(base.User, inst.DeployDir())
		// data dir would be empty for components which don't need it
		dataDirs := spec.MultiDirAbs(base.User, inst.DataDir())
		// log dir will always be with values, but might not used by the component
		logDir := spec.Abs(base.User, inst.LogDir())

		deployDirs := []string{
			deployDir,
			filepath.Join(deployDir, "bin"),
			filepath.Join(deployDir, "conf"),
			filepath.Join(deployDir, "scripts"),
		}
		// Deploy component
		tb := task.NewBuilder().
			UserSSH(
				inst.GetHost(),
				inst.GetSSHPort(),
				base.User,
				gOpt.SSHTimeout,
				gOpt.OptTimeout,
				gOpt.SSHProxyHost,
				gOpt.SSHProxyPort,
				gOpt.SSHProxyUser,
				p.Password,
				p.IdentityFile,
				p.IdentityFilePassphrase,
				gOpt.SSHProxyTimeout,
				gOpt.SSHType,
				sshType,
			).
			Mkdir(base.User, inst.GetHost(), deployDirs...).
			Mkdir(base.User, inst.GetHost(), dataDirs...).
			Mkdir(base.User, inst.GetHost(), logDir)

		srcPath := ""
		if patchedComponents.Exist(inst.ComponentName()) {
			srcPath = specManager.Path(name, spec.PatchDirName, inst.ComponentName()+".tar.gz")
		}

		if deployerInstance, ok := inst.(DeployerInstance); ok {
			deployerInstance.Deploy(tb, srcPath, deployDir, version, name, version)
		} else {
			// copy dependency component if needed
			switch inst.ComponentName() {
			default:
				tb.CopyComponent(
					inst.ComponentName(),
					inst.OS(),
					inst.Arch(),
					version,
					srcPath,
					inst.GetHost(),
					deployDir,
				)
			}
		}

		t := tb.ScaleConfig(name,
			base.Version,
			m.specManager,
			topo,
			inst,
			base.User,
			meta.DirPaths{
				Deploy: deployDir,
				Data:   dataDirs,
				Log:    logDir,
			},
		).Build()
		deployCompTasks = append(deployCompTasks, t)
	})
	if iterErr != nil {
		return nil, iterErr
	}

	noAgentHosts := set.NewStringSet()

	mergedTopo.IterInstance(func(inst spec.Instance) {
		deployDir := spec.Abs(base.User, inst.DeployDir())
		// data dir would be empty for components which don't need it
		dataDirs := spec.MultiDirAbs(base.User, inst.DataDir())
		// log dir will always be with values, but might not used by the component
		logDir := spec.Abs(base.User, inst.LogDir())

		// Download and copy the latest component to remote if the cluster is imported from Ansible
		tb := task.NewBuilder()

		// add the instance to ignore list if it marks itself as ignore_exporter
		if inst.IgnoreMonitorAgent() {
			noAgentHosts.Insert(inst.GetHost())
		}

		// Refresh all configuration
		t := tb.InitConfig(name,
			base.Version,
			m.specManager,
			inst,
			base.User,
			true, // always ignore config check result in scale out
			meta.DirPaths{
				Deploy: deployDir,
				Data:   dataDirs,
				Log:    logDir,
				Cache:  specManager.Path(name, spec.TempConfigPath),
			},
		).Build()
		refreshConfigTasks = append(refreshConfigTasks, t)
	})

	// Deploy monitor relevant components to remote
	dlTasks, dpTasks, err := buildMonitoredDeployTask(
		m,
		name,
		uninitializedHosts,
		noAgentHosts,
		topo.BaseTopo().GlobalOptions,
		topo.BaseTopo().MonitoredOptions,
		base.Version,
		gOpt,
		topo.(*spec.Specification).ElasticSearchAddress(),
		topo.(*spec.Specification).TiEMLogPaths(),
		p,
	)
	if err != nil {
		return nil, err
	}
	downloadCompTasks = append(downloadCompTasks, convertStepDisplaysToTasks(dlTasks)...)
	deployCompTasks = append(deployCompTasks, convertStepDisplaysToTasks(dpTasks)...)

	builder, err := m.sshTaskBuilder(name, topo, base.User, gOpt)
	if err != nil {
		return nil, err
	}
	builder.
		Parallel(false, downloadCompTasks...).
		Parallel(false, envInitTasks...).
		Parallel(false, deployCompTasks...)

	if afterDeploy != nil {
		afterDeploy(builder, newPart)
	}

	builder.
		Func("Save meta", func(_ context.Context) error {
			metadata.SetTopology(mergedTopo)
			return m.specManager.SaveMeta(name, metadata)
		}).
		Parallel(false, refreshConfigTasks...).
		Parallel(false, buildReloadPromTasks(metadata.GetTopology())...).
		Func("StartCluster", func(ctx context.Context) error {
			return operator.Start(ctx, newPart, operator.Options{OptTimeout: gOpt.OptTimeout, Operation: operator.ScaleOutOperation}, tlsCfg)
		})

	if final != nil {
		final(builder, name, metadata)
	}

	return builder.Build(), nil
}

type hostInfo struct {
	ssh  int    // ssh port of host
	os   string // operating system
	arch string // cpu architecture
	// vendor string
}

// Deprecated
func convertStepDisplaysToTasks(t []*task.StepDisplay) []task.Task {
	tasks := make([]task.Task, 0, len(t))
	for _, sd := range t {
		tasks = append(tasks, sd)
	}
	return tasks
}

func buildMonitoredDeployTask(
	m *Manager,
	name string,
	uniqueHosts map[string]hostInfo, // host -> ssh-port, os, arch
	noAgentHosts set.StringSet, // hosts that do not deploy monitor agents
	globalOptions *spec.GlobalOptions,
	monitoredOptions *cspec.MonitoredOptions,
	version string,
	gOpt operator.Options,
	esHosts []string,
	tiemHosts map[string]*config.LogPathInfo,
	p *tui.SSHConnectionProps,
) (downloadCompTasks []*task.StepDisplay, deployCompTasks []*task.StepDisplay, err error) {
	if monitoredOptions == nil {
		return
	}

	if len(esHosts) < 1 {
		err = fmt.Errorf("no elasticsearch node available in the topology")
		return
	}

	uniqueCompOSArch := set.NewStringSet()
	// monitoring agents
	for _, comp := range []string{
		spec.ComponentNodeExporter,
		spec.ComponentBlackboxExporter,
		spec.ComponentFilebeat,
	} {
		version := m.bindVersion(comp, version)

		for host, info := range uniqueHosts {
			// skip deploying monitoring agents if the instance is marked so
			if noAgentHosts.Exist(host) {
				continue
			}

			// populate unique comp-os-arch set
			key := fmt.Sprintf("%s-%s-%s", comp, info.os, info.arch)
			if found := uniqueCompOSArch.Exist(key); !found {
				uniqueCompOSArch.Insert(key)
				downloadCompTasks = append(downloadCompTasks, task.NewBuilder().
					Download(comp, info.os, info.arch, version).
					BuildAsStep(fmt.Sprintf("  - Download %s:%s (%s/%s)", comp, version, info.os, info.arch)))
			}

			deployDir := spec.Abs(globalOptions.User, monitoredOptions.DeployDir)
			// data dir would be empty for components which don't need it
			dataDir := monitoredOptions.DataDir
			// the default data_dir is relative to deploy_dir
			if dataDir != "" && !strings.HasPrefix(dataDir, "/") {
				dataDir = filepath.Join(deployDir, dataDir)
			}
			// log dir will always be with values, but might not used by the component
			logDir := spec.Abs(globalOptions.User, monitoredOptions.LogDir)

			deployDirs := []string{
				deployDir,
				dataDir,
				logDir,
				filepath.Join(deployDir, "bin"),
				filepath.Join(deployDir, "conf"),
				filepath.Join(deployDir, "scripts"),
			}

			// Deploy component
			tb := task.NewBuilder().
				UserSSH(
					host,
					info.ssh,
					globalOptions.User,
					gOpt.SSHTimeout,
					gOpt.OptTimeout,
					gOpt.SSHProxyHost,
					gOpt.SSHProxyPort,
					gOpt.SSHProxyUser,
					p.Password,
					p.IdentityFile,
					p.IdentityFilePassphrase,
					gOpt.SSHProxyTimeout,
					gOpt.SSHType,
					globalOptions.SSHType,
				).
				Mkdir(globalOptions.User, host, deployDirs...).
				CopyComponent(
					comp,
					info.os,
					info.arch,
					version,
					"",
					host,
					deployDir,
				).
				MonitoredConfig(
					name,
					comp,
					host,
					globalOptions.ResourceControl,
					monitoredOptions,
					globalOptions.User,
					meta.DirPaths{
						Deploy: deployDir,
						Data:   []string{dataDir},
						Log:    logDir,
						Cache:  m.specManager.Path(name, spec.TempConfigPath),
					},
					esHosts[0],
					tiemHosts,
				)

			deployCompTasks = append(deployCompTasks, tb.BuildAsStep(fmt.Sprintf("  - Copy %s -> %s", comp, host)))
		}
	}
	return
}

func buildRefreshMonitoredConfigTasks(
	specManager *spec.SpecManager,
	name string,
	uniqueHosts map[string]hostInfo, // host -> ssh-port, os, arch
	noAgentHosts set.StringSet,
	globalOptions spec.GlobalOptions,
	monitoredOptions *cspec.MonitoredOptions,
	sshTimeout, exeTimeout uint64,
	gOpt operator.Options,
	esHosts []string,
	tiemHosts map[string]*config.LogPathInfo,
	p *tui.SSHConnectionProps,
) []*task.StepDisplay {
	if monitoredOptions == nil {
		return nil
	}

	tasks := []*task.StepDisplay{}
	// monitoring agents
	for _, comp := range []string{
		spec.ComponentNodeExporter,
		spec.ComponentBlackboxExporter,
		spec.ComponentFilebeat,
	} {
		for host, info := range uniqueHosts {
			if noAgentHosts.Exist(host) {
				continue
			}

			deployDir := spec.Abs(globalOptions.User, monitoredOptions.DeployDir)
			// data dir would be empty for components which don't need it
			dataDir := monitoredOptions.DataDir
			// the default data_dir is relative to deploy_dir
			if dataDir != "" && !strings.HasPrefix(dataDir, "/") {
				dataDir = filepath.Join(deployDir, dataDir)
			}
			// log dir will always be with values, but might not used by the component
			logDir := spec.Abs(globalOptions.User, monitoredOptions.LogDir)
			// Generate configs
			t := task.NewBuilder().
				UserSSH(
					host,
					info.ssh,
					globalOptions.User,
					sshTimeout,
					exeTimeout,
					gOpt.SSHProxyHost,
					gOpt.SSHProxyPort,
					gOpt.SSHProxyUser,
					p.Password,
					p.IdentityFile,
					p.IdentityFilePassphrase,
					gOpt.SSHProxyTimeout,
					gOpt.SSHType,
					globalOptions.SSHType,
				).
				MonitoredConfig(
					name,
					comp,
					host,
					globalOptions.ResourceControl,
					monitoredOptions,
					globalOptions.User,
					meta.DirPaths{
						Deploy: deployDir,
						Data:   []string{dataDir},
						Log:    logDir,
						Cache:  specManager.Path(name, spec.TempConfigPath),
					},
					esHosts[0],
					tiemHosts,
				).
				BuildAsStep(fmt.Sprintf("  - Refresh config %s -> %s", comp, host))
			tasks = append(tasks, t)
		}
	}
	return tasks
}

func buildRegenConfigTasks(m *Manager, name string, topo spec.Topology, base *spec.BaseMeta, nodes []string, ignoreCheck bool) []*task.StepDisplay {
	var tasks []*task.StepDisplay
	deletedNodes := set.NewStringSet(nodes...)

	topo.IterInstance(func(instance spec.Instance) {
		if deletedNodes.Exist(instance.ID()) {
			return
		}
		compName := instance.ComponentName()
		deployDir := spec.Abs(base.User, instance.DeployDir())
		// data dir would be empty for components which don't need it
		dataDirs := spec.MultiDirAbs(base.User, instance.DataDir())
		// log dir will always be with values, but might not used by the component
		logDir := spec.Abs(base.User, instance.LogDir())

		// Download and copy the latest component to remote if the cluster is imported from Ansible
		tb := task.NewBuilder()
		t := tb.
			InitConfig(
				name,
				base.Version,
				m.specManager,
				instance,
				base.User,
				ignoreCheck,
				meta.DirPaths{
					Deploy: deployDir,
					Data:   dataDirs,
					Log:    logDir,
					Cache:  m.specManager.Path(name, spec.TempConfigPath),
				},
			).
			BuildAsStep(fmt.Sprintf("  - Regenerate config %s -> %s", compName, instance.ID()))
		tasks = append(tasks, t)
	})

	return tasks
}

// buildDownloadCompTasks build download component tasks
func buildDownloadCompTasks(clusterVersion string, topo spec.Topology, bindVersion spec.BindVersion) []*task.StepDisplay {
	var tasks []*task.StepDisplay
	uniqueTaskList := set.NewStringSet()
	topo.IterInstance(func(inst spec.Instance) {
		key := fmt.Sprintf("%s-%s-%s", inst.ComponentName(), inst.OS(), inst.Arch())
		if found := uniqueTaskList.Exist(key); !found {
			uniqueTaskList.Insert(key)

			version := bindVersion(inst.ComponentName(), clusterVersion)

			t := task.NewBuilder().
				Download(inst.ComponentName(), inst.OS(), inst.Arch(), version).
				BuildAsStep(fmt.Sprintf("  - Download %s:%s (%s/%s)",
					inst.ComponentName(), version, inst.OS(), inst.Arch()))
			tasks = append(tasks, t)
		}
	})
	return tasks
}

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

package operator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/errors"
	perrs "github.com/pingcap/errors"
	"github.com/pingcap/tiunimanager/tiup/spec"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/cluster/executor"
	"github.com/pingcap/tiup/pkg/cluster/module"
	"github.com/pingcap/tiup/pkg/logger/log"
	"github.com/pingcap/tiup/pkg/set"
)

// Destroy the cluster.
func Destroy(
	ctx context.Context,
	cluster spec.Topology,
	options Options,
) error {
	coms := cluster.ComponentsByStopOrder()

	instCount := map[string]int{}
	cluster.IterInstance(func(inst spec.Instance) {
		instCount[inst.GetHost()]++
	})

	for _, com := range coms {
		insts := com.Instances()
		err := DestroyComponent(ctx, insts, cluster, options)
		if err != nil && !options.Force {
			return errors.Annotatef(err, "failed to destroy %s", com.Name())
		}
		for _, inst := range insts {
			instCount[inst.GetHost()]--
			if instCount[inst.GetHost()] == 0 {
				if cluster.GetMonitoredOptions() != nil {
					if err := DestroyMonitored(ctx, inst, cluster.GetMonitoredOptions(), options.OptTimeout); err != nil && !options.Force {
						return err
					}
				}
			}
		}
	}

	gOpts := cluster.BaseTopo().GlobalOptions

	// Delete all global deploy directory
	for host := range instCount {
		if err := DeleteGlobalDirs(ctx, host, gOpts); err != nil {
			return nil
		}
	}

	// after all things done, try to remove SSH public key
	for host := range instCount {
		if err := DeletePublicKey(ctx, host); err != nil {
			return nil
		}
	}

	return nil
}

// StopAndDestroyInstance stop and destroy the instance,
// if this instance is the host's last one, and the host has monitor deployed,
// we need to destroy the monitor, too
func StopAndDestroyInstance(ctx context.Context, cluster spec.Topology, instance spec.Instance, options Options, destroyNode bool) error {
	ignoreErr := options.Force
	compName := instance.ComponentName()
	noAgentHosts := set.NewStringSet()

	cluster.IterInstance(func(inst spec.Instance) {
		if inst.IgnoreMonitorAgent() {
			noAgentHosts.Insert(inst.GetHost())
		}
	})

	// just try to stop and destroy
	if err := StopComponent(ctx, []spec.Instance{instance}, noAgentHosts, options.OptTimeout); err != nil {
		if !ignoreErr {
			return errors.Annotatef(err, "failed to stop %s", compName)
		}
		log.Warnf("failed to stop %s: %v", compName, err)
	}
	if err := DestroyComponent(ctx, []spec.Instance{instance}, cluster, options); err != nil {
		if !ignoreErr {
			return errors.Annotatef(err, "failed to destroy %s", compName)
		}
		log.Warnf("failed to destroy %s: %v", compName, err)
	}

	if destroyNode {
		// monitoredOptions for dm cluster is nil
		monitoredOptions := cluster.GetMonitoredOptions()

		if monitoredOptions != nil && !instance.IgnoreMonitorAgent() {
			if err := StopMonitored(ctx, []string{instance.GetHost()}, noAgentHosts, monitoredOptions, options.OptTimeout); err != nil {
				if !ignoreErr {
					return errors.Annotatef(err, "failed to stop monitor")
				}
				log.Warnf("failed to stop %s: %v", "monitor", err)
			}
			if err := DestroyMonitored(ctx, instance, monitoredOptions, options.OptTimeout); err != nil {
				if !ignoreErr {
					return errors.Annotatef(err, "failed to destroy monitor")
				}
				log.Warnf("failed to destroy %s: %v", "monitor", err)
			}
		}

		if err := DeletePublicKey(ctx, instance.GetHost()); err != nil {
			if !ignoreErr {
				return errors.Annotatef(err, "failed to delete public key")
			}
			log.Warnf("failed to delete public key")
		}
	}
	return nil
}

// DeleteGlobalDirs deletes all global directories if they are empty
func DeleteGlobalDirs(ctx context.Context, host string, options *spec.GlobalOptions) error {
	if options == nil {
		return nil
	}

	e := ctxt.GetInner(ctx).Get(host)
	log.Infof("Clean global directories %s", host)
	for _, dir := range []string{options.LogDir, options.DeployDir, options.DataDir} {
		if dir == "" {
			continue
		}
		dir = spec.Abs(options.User, dir)

		log.Infof("\tClean directory %s on instance %s", dir, host)

		c := module.ShellModuleConfig{
			Command:  fmt.Sprintf("rmdir %s > /dev/null 2>&1 || true", dir),
			Chdir:    "",
			UseShell: false,
		}
		shell := module.NewShellModule(c)
		stdout, stderr, err := shell.Execute(ctx, e)

		if len(stdout) > 0 {
			fmt.Println(string(stdout))
		}
		if len(stderr) > 0 {
			log.Errorf(string(stderr))
		}

		if err != nil {
			return errors.Annotatef(err, "failed to clean directory %s on: %s", dir, host)
		}
	}

	log.Infof("Clean global directories %s success", host)
	return nil
}

// DeletePublicKey deletes the SSH public key from host
func DeletePublicKey(ctx context.Context, host string) error {
	e := ctxt.GetInner(ctx).Get(host)
	log.Infof("Delete public key %s", host)
	_, pubKeyPath := ctxt.GetInner(ctx).GetSSHKeySet()
	publicKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return perrs.Trace(err)
	}

	pubKey := string(bytes.TrimSpace(publicKey))
	pubKey = strings.ReplaceAll(pubKey, "/", "\\/")
	pubKeysFile := executor.FindSSHAuthorizedKeysFile(ctx, e)

	// delete the public key with Linux `sed` toolkit
	c := module.ShellModuleConfig{
		Command:  fmt.Sprintf("sed -i '/%s/d' %s", pubKey, pubKeysFile),
		UseShell: false,
	}
	shell := module.NewShellModule(c)
	stdout, stderr, err := shell.Execute(ctx, e)

	if len(stdout) > 0 {
		fmt.Println(string(stdout))
	}
	if len(stderr) > 0 {
		log.Errorf(string(stderr))
	}

	if err != nil {
		return errors.Annotatef(err, "failed to delete pulblic key on: %s", host)
	}

	log.Infof("Delete public key %s success", host)
	return nil
}

// DestroyMonitored destroy the monitored service.
func DestroyMonitored(ctx context.Context, inst spec.Instance, options *spec.MonitoredOptions, timeout uint64) error {
	e := ctxt.GetInner(ctx).Get(inst.GetHost())
	log.Infof("Destroying monitored %s", inst.GetHost())

	log.Infof("\tDestroying instance %s", inst.GetHost())

	// Stop by systemd.
	delPaths := make([]string, 0)

	delPaths = append(delPaths, options.DataDir)
	delPaths = append(delPaths, options.LogDir)

	// In TiDB-Ansible, deploy dir are shared by all components on the same
	// host, so not deleting it.
	if !inst.IsImported() {
		delPaths = append(delPaths, options.DeployDir)
	} else {
		log.Warnf("Monitored deploy dir %s not deleted for TiDB-Ansible imported instance %s.",
			options.DeployDir, inst.InstanceName())
	}

	delPaths = append(delPaths, fmt.Sprintf("/etc/systemd/system/%s-%d.service", spec.ComponentNodeExporter, options.NodeExporterPort))

	c := module.ShellModuleConfig{
		Command:  fmt.Sprintf("rm -rf %s;", strings.Join(delPaths, " ")),
		Sudo:     true, // the .service files are in a directory owned by root
		Chdir:    "",
		UseShell: false,
	}
	shell := module.NewShellModule(c)
	stdout, stderr, err := shell.Execute(ctx, e)

	if len(stdout) > 0 {
		fmt.Println(string(stdout))
	}
	if len(stderr) > 0 {
		log.Errorf(string(stderr))
	}

	if err != nil {
		return errors.Annotatef(err, "failed to destroy monitored: %s", inst.GetHost())
	}

	if err := spec.PortStopped(ctx, e, options.NodeExporterPort, timeout); err != nil {
		str := fmt.Sprintf("%s failed to destroy node exportoer: %s", inst.GetHost(), err)
		log.Errorf(str)
		return errors.Annotatef(err, str)
	}

	log.Infof("Destroy monitored on %s success", inst.GetHost())

	return nil
}

// CleanupComponent cleanup the instances
func CleanupComponent(ctx context.Context, delFileMaps map[string]set.StringSet) error {
	for host, delFiles := range delFileMaps {
		e := ctxt.GetInner(ctx).Get(host)
		log.Infof("Cleanup instance %s", host)
		log.Debugf("Deleting paths on %s: %s", host, strings.Join(delFiles.Slice(), " "))
		c := module.ShellModuleConfig{
			Command:  fmt.Sprintf("rm -rf %s;", strings.Join(delFiles.Slice(), " ")),
			Sudo:     true, // the .service files are in a directory owned by root
			Chdir:    "",
			UseShell: true,
		}
		shell := module.NewShellModule(c)
		stdout, stderr, err := shell.Execute(ctx, e)

		if len(stdout) > 0 {
			fmt.Println(string(stdout))
		}
		if len(stderr) > 0 {
			log.Errorf(string(stderr))
		}

		if err != nil {
			return errors.Annotatef(err, "failed to cleanup: %s", host)
		}

		log.Infof("Cleanup %s success", host)
	}

	return nil
}

// DestroyComponent destroy the instances.
func DestroyComponent(ctx context.Context, instances []spec.Instance, cls spec.Topology, options Options) error {
	if len(instances) == 0 {
		return nil
	}

	name := instances[0].ComponentName()
	log.Infof("Destroying component %s", name)

	retainDataRoles := set.NewStringSet(options.RetainDataRoles...)
	retainDataNodes := set.NewStringSet(options.RetainDataNodes...)

	for _, ins := range instances {
		// Some data of instances will be retained
		dataRetained := retainDataRoles.Exist(ins.ComponentName()) ||
			retainDataNodes.Exist(ins.ID()) || retainDataNodes.Exist(ins.GetHost())

		e := ctxt.GetInner(ctx).Get(ins.GetHost())
		log.Infof("Destroying instance %s", ins.GetHost())

		var dataDirs []string
		if len(ins.DataDir()) > 0 {
			dataDirs = strings.Split(ins.DataDir(), ",")
		}

		deployDir := ins.DeployDir()
		delPaths := set.NewStringSet()

		// Retain the deploy directory if the users want to retain the data directory
		// and the data directory is a sub-directory of deploy directory
		keepDeployDir := false

		for _, dataDir := range dataDirs {
			// Don't delete the parent directory if any sub-directory retained
			keepDeployDir = (dataRetained && strings.HasPrefix(dataDir, deployDir)) || keepDeployDir
			if !dataRetained && cls.CountDir(ins.GetHost(), dataDir) == 1 {
				// only delete path if it is not used by any other instance in the cluster
				delPaths.Insert(dataDir)
			}
		}

		logDir := ins.LogDir()

		// In TiDB-Ansible, deploy dir are shared by all components on the same
		// host, so not deleting it.
		if ins.IsImported() {
			// not deleting files for imported clusters
			if !strings.HasPrefix(logDir, ins.DeployDir()) && cls.CountDir(ins.GetHost(), logDir) == 1 {
				delPaths.Insert(logDir)
			}
			log.Warnf("Deploy dir %s not deleted for TiDB-Ansible imported instance %s.",
				ins.DeployDir(), ins.InstanceName())
		} else {
			if keepDeployDir {
				delPaths.Insert(filepath.Join(deployDir, "conf"))
				delPaths.Insert(filepath.Join(deployDir, "bin"))
				delPaths.Insert(filepath.Join(deployDir, "scripts"))
				// only delete path if it is not used by any other instance in the cluster
				if strings.HasPrefix(logDir, deployDir) && cls.CountDir(ins.GetHost(), logDir) == 1 {
					delPaths.Insert(logDir)
				}
			} else {
				// only delete path if it is not used by any other instance in the cluster
				if cls.CountDir(ins.GetHost(), logDir) == 1 {
					delPaths.Insert(logDir)
				}
				if cls.CountDir(ins.GetHost(), ins.DeployDir()) == 1 {
					delPaths.Insert(ins.DeployDir())
				}
			}
		}

		// check for deploy dir again, to avoid unused files being left on disk
		dpCnt := 0
		for _, dir := range delPaths.Slice() {
			if strings.HasPrefix(dir, deployDir+"/") { // only check subdir of deploy dir
				dpCnt++
			}
		}
		if !ins.IsImported() && cls.CountDir(ins.GetHost(), deployDir)-dpCnt == 1 {
			delPaths.Insert(deployDir)
		}

		if svc := ins.ServiceName(); svc != "" {
			delPaths.Insert(fmt.Sprintf("/etc/systemd/system/%s", svc))
		}
		log.Debugf("Deleting paths on %s: %s", ins.GetHost(), strings.Join(delPaths.Slice(), " "))
		c := module.ShellModuleConfig{
			Command:  fmt.Sprintf("rm -rf %s;", strings.Join(delPaths.Slice(), " ")),
			Sudo:     true, // the .service files are in a directory owned by root
			Chdir:    "",
			UseShell: false,
		}
		shell := module.NewShellModule(c)
		stdout, stderr, err := shell.Execute(ctx, e)

		if len(stdout) > 0 {
			fmt.Println(string(stdout))
		}
		if len(stderr) > 0 {
			log.Errorf(string(stderr))
		}

		if err != nil {
			return errors.Annotatef(err, "failed to destroy: %s", ins.GetHost())
		}

		log.Infof("Destroy %s success", ins.GetHost())
		log.Infof("- Destroy %s paths: %v", ins.ComponentName(), delPaths.Slice())
	}

	return nil
}

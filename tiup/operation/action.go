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

package operator

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	tiuputils "github.com/pingcap/tiup/pkg/utils"

	"github.com/pingcap-inc/tiem/tiup/spec"
	"github.com/pingcap-inc/tiem/tiup/utils"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/checkpoint"
	"github.com/pingcap/tiup/pkg/cluster/ctxt"
	"github.com/pingcap/tiup/pkg/cluster/module"
	"github.com/pingcap/tiup/pkg/logger/log"
	"github.com/pingcap/tiup/pkg/set"
	"golang.org/x/sync/errgroup"
)

var (
	actionPrevMsgs = map[string]string{
		"start":   "Starting",
		"stop":    "Stopping",
		"enable":  "Enabling",
		"disable": "Disabling",
	}
	actionPostMsgs = map[string]string{}
)

func init() {
	for action := range actionPrevMsgs {
		actionPostMsgs[action] = strings.Title(action)
	}
}

// Enable will enable/disable the cluster
func Enable(
	ctx context.Context,
	cluster spec.Topology,
	options Options,
	isEnable bool,
) error {
	roleFilter := set.NewStringSet(options.Roles...)
	nodeFilter := set.NewStringSet(options.Nodes...)
	components := cluster.ComponentsByStartOrder()
	components = FilterComponent(components, roleFilter)
	monitoredOptions := cluster.GetMonitoredOptions()
	noAgentHosts := set.NewStringSet()

	instCount := map[string]int{}
	cluster.IterInstance(func(inst spec.Instance) {
		if inst.IgnoreMonitorAgent() {
			noAgentHosts.Insert(inst.GetHost())
		} else {
			instCount[inst.GetHost()]++
		}
	})

	for _, comp := range components {
		insts := FilterInstance(comp.Instances(), nodeFilter)
		err := EnableComponent(ctx, insts, noAgentHosts, options, isEnable)
		if err != nil {
			return errors.Annotatef(err, "failed to enable/disable %s", comp.Name())
		}

		for _, inst := range insts {
			if !inst.IgnoreMonitorAgent() {
				instCount[inst.GetHost()]--
			}
		}
	}

	if monitoredOptions == nil {
		return nil
	}

	hosts := make([]string, 0)
	for host, count := range instCount {
		// don't disable the monitor component if the instance's host contain other components
		if count != 0 {
			continue
		}
		hosts = append(hosts, host)
	}

	return EnableMonitored(ctx, hosts, noAgentHosts, monitoredOptions, options.OptTimeout, isEnable)
}

// Start the cluster.
func Start(
	ctx context.Context,
	cluster spec.Topology,
	options Options,
	tlsCfg *tls.Config,
) error {
	uniqueHosts := set.NewStringSet()
	roleFilter := set.NewStringSet(options.Roles...)
	nodeFilter := set.NewStringSet(options.Nodes...)
	components := cluster.ComponentsByStartOrder()
	components = FilterComponent(components, roleFilter)
	monitoredOptions := cluster.GetMonitoredOptions()
	noAgentHosts := set.NewStringSet()

	cluster.IterInstance(func(inst spec.Instance) {
		if inst.IgnoreMonitorAgent() {
			noAgentHosts.Insert(inst.GetHost())
		}
	})

	for _, comp := range components {
		insts := FilterInstance(comp.Instances(), nodeFilter)
		err := StartComponent(ctx, insts, noAgentHosts, options, tlsCfg)
		if err != nil {
			return errors.Annotatef(err, "failed to start %s", comp.Name())
		}
		for _, inst := range insts {
			if !inst.IgnoreMonitorAgent() {
				uniqueHosts.Insert(inst.GetHost())
			}
			// init elasticsearch index settings
			if comp.Name() == spec.ComponentElasticSearchServer {
				if err = initElasticSearch(tlsCfg, inst); err != nil {
					return err
				}
			}
			// init kibana index patterns
			if comp.Name() == spec.ComponentKibana {
				if err = initKibana(ctx, tlsCfg, inst); err != nil {
					return err
				}
			}
			// init grafana index patterns
			if comp.Name() == spec.ComponentGrafana {
				if len(cluster.BaseTopo().WebServers) <= 0 {
					return fmt.Errorf("Component web_servers can not empty ")
				}
				if err = initGrafana(ctx, tlsCfg, inst, cluster.BaseTopo().WebServers[0]); err != nil {
					return err
				}
			}
		}
	}

	if monitoredOptions == nil {
		return nil
	}

	hosts := make([]string, 0, len(uniqueHosts))
	for host := range uniqueHosts {
		hosts = append(hosts, host)
	}
	return StartMonitored(ctx, hosts, noAgentHosts, monitoredOptions, options.OptTimeout)
}

func initElasticSearch(tlsCfg *tls.Config, inst spec.Instance) error {
	// loop get es status
	startTime := time.Now().Unix()
	for {
		client := tiuputils.NewHTTPClient(2*time.Second, tlsCfg)
		_, err := client.Get(context.TODO(), fmt.Sprintf("http://%s:%d/_cat", inst.GetHost(), inst.GetPort()))
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
		log.Debugf("check elasticsearch status error: %s", err.Error())
		if time.Now().Unix()-startTime > 180 {
			return fmt.Errorf("Start compoent %s timeout, more than 180s, please check logs: %s/kibana.log ", spec.ComponentElasticSearchServer, inst.LogDir())
		}
	}
	path := "/_template/custom_em"
	url := fmt.Sprintf("http://%s:%d%s", inst.GetHost(), inst.GetPort(), path)
	log.Debugf("init elasticsearch index template url: %s", url)

	content := map[string]interface{}{}
	params := map[string]int64{}
	// set setting max_result_window by indices settings
	params["max_result_window"] = 100000000
	content["index_patterns"] = []string{"em-*"}
	content["settings"] = params
	log.Debugf("init elasticsearch index template params: %s", content)
	code, resp := utils.PUT(url, content, map[string]string{})
	log.Debugf("init elasticsearch index template response code: %d, content: %s", code, resp)
	return nil
}

func initKibana(ctx context.Context, tlsCfg *tls.Config, inst spec.Instance) error {
	startTime := time.Now().Unix()
	// loop get kibana status
	for {
		client := tiuputils.NewHTTPClient(2*time.Second, tlsCfg)
		_, err := client.Get(context.TODO(), fmt.Sprintf("http://%s:%d/status", inst.GetHost(), inst.GetPort()))
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
		log.Debugf("check kibana status error: %s", err.Error())
		if time.Now().Unix()-startTime > 180 {
			return fmt.Errorf("Start compoent %s timeout, more than 180s, please check logs: %s/kibana.log ", spec.ComponentKibana, inst.LogDir())
		}
	}

	// Copy to remote server
	exec, found := ctxt.GetInner(ctx).GetExecutor(inst.GetHost())
	if !found {
		return stderrors.New("no executor")
	}

	path := "/api/saved_objects/_import?overwrite=true"
	url := fmt.Sprintf("http://%s:%d%s", inst.GetHost(), inst.GetPort(), path)
	log.Debugf("init kibana index patterns url: %s", url)

	fileName := "index_patterns.ndjson"
	dstPath := "/tmp/" + fileName
	if err := exec.Transfer(ctx, inst.DeployDir()+"/bin/"+fileName, dstPath, true, 0); err != nil {
		return fmt.Errorf("init kibana transfer file error: %s", err.Error())
	}

	uploads := make([]utils.UploadFile, 0)
	uploads = append(uploads, utils.UploadFile{
		Name:     "file",
		Filepath: dstPath,
	})
	headers := map[string]string{"kbn-xsrf": "reporting"}
	code, resp := utils.PostFile(url, map[string]interface{}{}, uploads, headers)
	log.Debugf("init kibana index patterns response code: %d, content: %s", code, resp)
	return nil
}

func initGrafana(ctx context.Context, tlsCfg *tls.Config, inst spec.Instance, webServer *spec.WebServerSpec) error {
	startTime := time.Now().Unix()
	baseUrl := fmt.Sprintf("http://%s:%d/grafana", webServer.Host, webServer.Port)
	log.Debugf("request grafana base url: %s", baseUrl)
	// loop get grafana status
	for {
		client := tiuputils.NewHTTPClient(2*time.Second, tlsCfg)
		_, err := client.Get(context.TODO(), baseUrl)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
		log.Debugf("check grafana status error: %s", err.Error())
		if time.Now().Unix()-startTime > 180 {
			return fmt.Errorf("Start compoent %s timeout, more than 180s, please check logs: %s/%s.log ", spec.ComponentGrafana, inst.LogDir(), spec.ComponentGrafana)
		}
	}

	folderName := "rules"

	// Determine if initialisation is required
	code, resp := utils.Get(baseUrl+"/api/folders/"+folderName, map[string]string{}, map[string]string{})
	log.Debugf("check init grafana response code: %d, content: %s", code, resp)
	if code == http.StatusOK {
		// The response status code 200, indicating that it has been initialized, is skipped directly
		return nil
	}

	// Copy to remote server
	exec, found := ctxt.GetInner(ctx).GetExecutor(inst.GetHost())
	if !found {
		return stderrors.New("no executor")
	}

	// alert configuration, include contact points & notification policies
	path := "/api/alertmanager/grafana/config/api/v1/alerts"
	url := baseUrl + path
	log.Debugf("init grafana alert configuration url: %s", url)
	fileName := "configuration.json"
	dstPath := "/tmp/" + fileName
	if err := exec.Transfer(ctx, inst.DeployDir()+"/bin/rules/"+fileName, dstPath, true, 0); err != nil {
		return fmt.Errorf("init grafana alert configuration transfer file error: %s", err.Error())
	}
	content, err := utils.ReadFile(dstPath)
	if err != nil {
		return errors.Annotatef(err, "failed to read file %s", dstPath)
	}
	data := map[string]interface{}{}
	if err = json.Unmarshal(content, &data); err != nil {
		return errors.Annotatef(err, "failed to json unmarshal")
	}
	code, resp = utils.PostJSON(url, data, map[string]string{})
	log.Debugf("init grafana alert configuration response code: %d, content: %s", code, resp)

	// create folder rules.
	data = map[string]interface{}{}
	data["title"] = folderName
	data["uid"] = folderName
	code, resp = utils.PostJSON(baseUrl+"/api/folders", data, map[string]string{})
	log.Debugf("init grafana create folder response code: %d, content: %s", code, resp)

	// alert rules
	path = "/api/ruler/grafana/api/v1/rules/rules"
	url = baseUrl + path
	log.Debugf("init grafana alert rules url: %s", url)
	fileName = "alerts.json"
	dstPath = "/tmp/" + fileName
	if err := exec.Transfer(ctx, inst.DeployDir()+"/bin/rules/"+fileName, dstPath, true, 0); err != nil {
		return fmt.Errorf("init grafana alert rules transfer file error: %s", err.Error())
	}
	content, err = utils.ReadFile(dstPath)
	if err != nil {
		return errors.Annotatef(err, "failed to read file %s", dstPath)
	}
	data = map[string]interface{}{}
	if err = json.Unmarshal(content, &data); err != nil {
		return errors.Annotatef(err, "failed to json unmarshal")
	}
	for _, v := range data {
		rules := make([]interface{}, 0)
		if err := utils.ConvertObj(v, &rules); err != nil {
			return errors.Annotatef(err, "failed to convert object")
		}
		for _, rule := range rules {
			data = map[string]interface{}{}
			if err := utils.ConvertObj(rule, &data); err != nil {
				return errors.Annotatef(err, "failed to convert object")
			}
			code, resp = utils.PostJSON(url, data, map[string]string{})
			log.Debugf("init grafana alert rules response code: %d, content: %s", code, resp)
		}
	}
	return nil
}

// Stop the cluster.
func Stop(
	ctx context.Context,
	cluster spec.Topology,
	options Options,
	tlsCfg *tls.Config,
) error {
	roleFilter := set.NewStringSet(options.Roles...)
	nodeFilter := set.NewStringSet(options.Nodes...)
	components := cluster.ComponentsByStopOrder()
	components = FilterComponent(components, roleFilter)
	monitoredOptions := cluster.GetMonitoredOptions()
	noAgentHosts := set.NewStringSet()

	instCount := map[string]int{}
	cluster.IterInstance(func(inst spec.Instance) {
		if inst.IgnoreMonitorAgent() {
			noAgentHosts.Insert(inst.GetHost())
		} else {
			instCount[inst.GetHost()]++
		}
	})

	for _, comp := range components {
		insts := FilterInstance(comp.Instances(), nodeFilter)
		err := StopComponent(ctx, insts, noAgentHosts, options.OptTimeout)
		if err != nil && !options.Force {
			return errors.Annotatef(err, "failed to stop %s", comp.Name())
		}
		for _, inst := range insts {
			if !inst.IgnoreMonitorAgent() {
				instCount[inst.GetHost()]--
			}
		}
	}

	if monitoredOptions == nil {
		return nil
	}

	hosts := make([]string, 0)
	for host, count := range instCount {
		if count != 0 {
			continue
		}
		hosts = append(hosts, host)
	}

	if err := StopMonitored(ctx, hosts, noAgentHosts, monitoredOptions, options.OptTimeout); err != nil && !options.Force {
		return err
	}
	return nil
}

// NeedCheckTombstone return true if we need to check and destroy some node.
func NeedCheckTombstone(topo *spec.Specification) bool {
	return false // not implemented for tiem
}

// Restart the cluster.
func Restart(
	ctx context.Context,
	cluster spec.Topology,
	options Options,
	tlsCfg *tls.Config,
) error {
	err := Stop(ctx, cluster, options, tlsCfg)
	if err != nil {
		return errors.Annotatef(err, "failed to stop")
	}

	err = Start(ctx, cluster, options, tlsCfg)
	if err != nil {
		return errors.Annotatef(err, "failed to start")
	}

	return nil
}

// StartMonitored start BlackboxExporter and NodeExporter
func StartMonitored(ctx context.Context, hosts []string, noAgentHosts set.StringSet, options *spec.MonitoredOptions, timeout uint64) error {
	return systemctlMonitor(ctx, hosts, noAgentHosts, options, "start", timeout)
}

// StopMonitored stop BlackboxExporter and NodeExporter
func StopMonitored(ctx context.Context, hosts []string, noAgentHosts set.StringSet, options *spec.MonitoredOptions, timeout uint64) error {
	return systemctlMonitor(ctx, hosts, noAgentHosts, options, "stop", timeout)
}

// EnableMonitored enable/disable monitor service in a cluster
func EnableMonitored(ctx context.Context, hosts []string, noAgentHosts set.StringSet, options *spec.MonitoredOptions, timeout uint64, isEnable bool) error {
	action := "disable"
	if isEnable {
		action = "enable"
	}

	return systemctlMonitor(ctx, hosts, noAgentHosts, options, action, timeout)
}

func systemctlMonitor(ctx context.Context, hosts []string, noAgentHosts set.StringSet, options *spec.MonitoredOptions, action string, timeout uint64) error {
	ports := monitorPortMap(options)
	for _, comp := range []string{
		spec.ComponentNodeExporter,
	} {
		log.Infof("%s component %s", actionPrevMsgs[action], comp)

		errg, _ := errgroup.WithContext(ctx)
		for _, host := range hosts {
			host := host
			if noAgentHosts.Exist(host) {
				log.Debugf("Ignored %s component %s for %s", action, comp, host)
				continue
			}
			nctx := checkpoint.NewContext(ctx)
			errg.Go(func() error {
				log.Infof("\t%s instance %s", actionPrevMsgs[action], host)
				e := ctxt.GetInner(nctx).Get(host)
				service := fmt.Sprintf("%s-%d.service", comp, ports[comp])

				if err := systemctl(nctx, e, service, action, timeout); err != nil {
					return toFailedActionError(err, action, host, service, "")
				}

				var err error
				switch action {
				case "start":
					err = spec.PortStarted(nctx, e, ports[comp], timeout)
				case "stop":
					err = spec.PortStopped(nctx, e, ports[comp], timeout)
				}

				if err != nil {
					return toFailedActionError(err, action, host, service, "")
				}
				log.Infof("\t%s %s success", actionPostMsgs[action], host)
				return nil
			})
		}
		if err := errg.Wait(); err != nil {
			return err
		}
	}

	return nil
}

func restartInstance(ctx context.Context, ins spec.Instance, timeout uint64) error {
	e := ctxt.GetInner(ctx).Get(ins.GetHost())
	log.Infof("\tRestarting instance %s", ins.ID())

	if err := systemctl(ctx, e, ins.ServiceName(), "restart", timeout); err != nil {
		return toFailedActionError(err, "restart", ins.GetHost(), ins.ServiceName(), ins.LogDir())
	}

	// Check ready.
	if err := ins.Ready(ctx, e, timeout); err != nil {
		return toFailedActionError(err, "restart", ins.GetHost(), ins.ServiceName(), ins.LogDir())
	}

	log.Infof("\tRestart instance %s success", ins.ID())

	return nil
}

// RestartComponent restarts the component.
func RestartComponent(ctx context.Context, instances []spec.Instance, timeout uint64) error {
	if len(instances) == 0 {
		return nil
	}

	name := instances[0].ComponentName()
	log.Infof("Restarting component %s", name)

	for _, ins := range instances {
		err := restartInstance(ctx, ins, timeout)
		if err != nil {
			return err
		}
	}

	return nil
}

func enableInstance(ctx context.Context, ins spec.Instance, timeout uint64, isEnable bool) error {
	e := ctxt.GetInner(ctx).Get(ins.GetHost())

	action := "disable"
	if isEnable {
		action = "enable"
	}
	log.Infof("\t%s instance %s", actionPrevMsgs[action], ins.ID())

	// Enable/Disable by systemd.
	if err := systemctl(ctx, e, ins.ServiceName(), action, timeout); err != nil {
		return toFailedActionError(err, action, ins.GetHost(), ins.ServiceName(), ins.LogDir())
	}

	log.Infof("\t%s instance %s success", actionPostMsgs[action], ins.ID())

	return nil
}

func startInstance(ctx context.Context, ins spec.Instance, timeout uint64) error {
	e := ctxt.GetInner(ctx).Get(ins.GetHost())
	log.Infof("\tStarting instance %s", ins.ID())

	if err := systemctl(ctx, e, ins.ServiceName(), "start", timeout); err != nil {
		return toFailedActionError(err, "start", ins.GetHost(), ins.ServiceName(), ins.LogDir())
	}

	// Check ready.
	if err := ins.Ready(ctx, e, timeout); err != nil {
		return toFailedActionError(err, "start", ins.GetHost(), ins.ServiceName(), ins.LogDir())
	}

	log.Infof("\tStart instance %s success", ins.ID())

	return nil
}

func systemctl(ctx context.Context, executor ctxt.Executor, service string, action string, timeout uint64) error {
	c := module.SystemdModuleConfig{
		Unit:         service,
		ReloadDaemon: true,
		Action:       action,
		Timeout:      time.Second * time.Duration(timeout),
	}
	systemd := module.NewSystemdModule(c)
	stdout, stderr, err := systemd.Execute(ctx, executor)

	if len(stdout) > 0 {
		fmt.Println(string(stdout))
	}
	if len(stderr) > 0 && !bytes.Contains(stderr, []byte("Created symlink ")) && !bytes.Contains(stderr, []byte("Removed symlink ")) {
		log.Errorf(string(stderr))
	}
	if len(stderr) > 0 && action == "stop" {
		// ignore "unit not loaded" error, as this means the unit is not
		// exist, and that's exactly what we want
		// NOTE: there will be a potential bug if the unit name is set
		// wrong and the real unit still remains started.
		if bytes.Contains(stderr, []byte(" not loaded.")) {
			log.Warnf(string(stderr))
			return nil // reset the error to avoid exiting
		}
		log.Errorf(string(stderr))
	}
	return err
}

// EnableComponent enable/disable the instances
func EnableComponent(ctx context.Context, instances []spec.Instance, noAgentHosts set.StringSet, options Options, isEnable bool) error {
	if len(instances) == 0 {
		return nil
	}

	name := instances[0].ComponentName()
	if isEnable {
		log.Infof("Enabling component %s", name)
	} else {
		log.Infof("Disabling component %s", name)
	}

	errg, _ := errgroup.WithContext(ctx)

	for _, ins := range instances {
		ins := ins

		// skip certain instances
		switch name {
		case spec.ComponentNodeExporter:
			if noAgentHosts.Exist(ins.GetHost()) {
				log.Debugf("Ignored enabling/disabling %s for %s:%d", name, ins.GetHost(), ins.GetPort())
				continue
			}
		}

		// the checkpoint part of context can't be shared between goroutines
		// since it's used to trace the stack, so we must create a new layer
		// of checkpoint context every time put it into a new goroutine.
		nctx := checkpoint.NewContext(ctx)
		errg.Go(func() error {
			err := enableInstance(nctx, ins, options.OptTimeout, isEnable)
			if err != nil {
				return err
			}
			return nil
		})
	}

	return errg.Wait()
}

// StartComponent start the instances.
func StartComponent(ctx context.Context, instances []spec.Instance, noAgentHosts set.StringSet, options Options, tlsCfg *tls.Config) error {
	if len(instances) == 0 {
		return nil
	}

	name := instances[0].ComponentName()
	log.Infof("Starting component %s", name)

	errg, _ := errgroup.WithContext(ctx)

	for _, ins := range instances {
		ins := ins
		switch name {
		case spec.ComponentNodeExporter:
			if noAgentHosts.Exist(ins.GetHost()) {
				log.Debugf("Ignored starting %s for %s:%d", name, ins.GetHost(), ins.GetPort())
				continue
			}
		}

		// the checkpoint part of context can't be shared between goroutines
		// since it's used to trace the stack, so we must create a new layer
		// of checkpoint context every time put it into a new goroutine.
		nctx := checkpoint.NewContext(ctx)
		errg.Go(func() error {
			if err := ins.PrepareStart(nctx, tlsCfg); err != nil {
				return err
			}
			return startInstance(nctx, ins, options.OptTimeout)
		})
	}

	return errg.Wait()
}

func serialStartInstances(ctx context.Context, instances []spec.Instance, options Options, tlsCfg *tls.Config) error {
	for _, ins := range instances {
		if err := ins.PrepareStart(ctx, tlsCfg); err != nil {
			return err
		}
		if err := startInstance(ctx, ins, options.OptTimeout); err != nil {
			return err
		}
	}
	return nil
}

func stopInstance(ctx context.Context, ins spec.Instance, timeout uint64) error {
	e := ctxt.GetInner(ctx).Get(ins.GetHost())
	log.Infof("\tStopping instance %s", ins.GetHost())

	if err := systemctl(ctx, e, ins.ServiceName(), "stop", timeout); err != nil {
		return toFailedActionError(err, "stop", ins.GetHost(), ins.ServiceName(), ins.LogDir())
	}

	log.Infof("\tStop %s %s success", ins.ComponentName(), ins.ID())

	return nil
}

// StopComponent stop the instances.
func StopComponent(ctx context.Context, instances []spec.Instance, noAgentHosts set.StringSet, timeout uint64) error {
	if len(instances) == 0 {
		return nil
	}

	name := instances[0].ComponentName()
	log.Infof("Stopping component %s", name)

	errg, _ := errgroup.WithContext(ctx)

	for _, ins := range instances {
		ins := ins
		switch name {
		case spec.ComponentNodeExporter:
			if noAgentHosts.Exist(ins.GetHost()) {
				log.Debugf("Ignored stopping %s for %s:%d", name, ins.GetHost(), ins.GetPort())
				continue
			}
		}

		// the checkpoint part of context can't be shared between goroutines
		// since it's used to trace the stack, so we must create a new layer
		// of checkpoint context every time put it into a new goroutine.
		nctx := checkpoint.NewContext(ctx)
		errg.Go(func() error {
			err := stopInstance(nctx, ins, timeout)
			if err != nil {
				return err
			}
			return nil
		})
	}

	return errg.Wait()
}

// PrintClusterStatus print cluster status into the io.Writer.
func PrintClusterStatus(ctx context.Context, cluster *spec.Specification) (health bool) {
	health = true

	for _, com := range cluster.ComponentsByStartOrder() {
		if len(com.Instances()) == 0 {
			continue
		}

		log.Infof("Checking service state of %s", com.Name())
		errg, _ := errgroup.WithContext(ctx)
		for _, ins := range com.Instances() {
			ins := ins

			// the checkpoint part of context can't be shared between goroutines
			// since it's used to trace the stack, so we must create a new layer
			// of checkpoint context every time put it into a new goroutine.
			nctx := checkpoint.NewContext(ctx)
			errg.Go(func() error {
				e := ctxt.GetInner(nctx).Get(ins.GetHost())
				active, err := GetServiceStatus(nctx, e, ins.ServiceName())
				if err != nil {
					health = false
					log.Errorf("\t%s\t%v", ins.GetHost(), err)
				} else {
					log.Infof("\t%s\t%s", ins.GetHost(), active)
				}
				return nil
			})
		}
		_ = errg.Wait()
	}

	return
}

// toFailedActionError formats the errror msg for failed action
func toFailedActionError(err error, action string, host, service, logDir string) error {
	return errors.Annotatef(err,
		"failed to %s: %s %s, please check the instance's log(%s) for more detail.",
		action, host, service, logDir,
	)
}

func monitorPortMap(options *spec.MonitoredOptions) map[string]int {
	return map[string]int{
		spec.ComponentNodeExporter: options.NodeExporterPort,
	}
}

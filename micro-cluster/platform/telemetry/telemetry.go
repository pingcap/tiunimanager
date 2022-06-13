/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
 * limitations under the License                                              *
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: telemetry.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/21
*******************************************************************************/

package telemetry

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/cluster/management"
	util "github.com/pingcap/tiunimanager/util/http"
	arch "github.com/pingcap/tiunimanager/util/sys/linux"
	"github.com/pingcap/tiunimanager/util/uuidutil"
	"github.com/pingcap/tiunimanager/util/versioninfo"
	"github.com/prometheus/client_golang/api"
	client "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/robfig/cron"
)

//TODO To move the const to the common
const (
	GetSystemConfigError errors.EM_ERROR_CODE = 70200
	//GetSystemConfigError:  {GetSystemConfigError, "get config from database error", 500},

	TelemetryPushDataToServiceError                    errors.EM_ERROR_CODE = 70500
	TelemetryGenerateEnterpriseNodeInfoDataError       errors.EM_ERROR_CODE = 70501
	TelemetryGenerateEMManagementHostsInfoDataError    errors.EM_ERROR_CODE = 70502
	TelemetryGenerateEMManagementClustersInfoDataError errors.EM_ERROR_CODE = 70503

	//TelemetryPushDataToServiceError:  {TelemetryPushDataToServiceError, "push telemetry data to service error", 500},
	//TelemetryGenerateEnterpriseNodeInfoDataError:  {TelemetryGenerateEnterpriseNodeInfoDataError, "telemetry generate enterprise manager node information data error", 500},
	//TelemetryGenerateEMManagementHostsInfoDataError:  {TelemetryGenerateEMManagementHostsInfoDataError, "telemetry generate enterprise manager management hosts information data error", 500},
	//TelemetryGenerateEMManagementClustersInfoDataError:  {TelemetryGenerateEMManagementClustersInfoDataError, "telemetry generate enterprise manager management clusters information data error", 500},
)

type Manager struct {
	JobCron *cron.Cron
	JobSpec string
}

//autoReportTelemetryDataHandler cron job handler
type autoReportTelemetryDataHandler struct {
	TelemetryManager *Manager
}

// Run cron job executor
func (a autoReportTelemetryDataHandler) Run() {
	framework.LogWithContext(framework.NewMicroCtxFromGinCtx(&gin.Context{})).Info(
		"the system reports telemetry data periodically and checks if this feature is enabled")
	a.TelemetryManager.Report(framework.NewMicroCtxFromGinCtx(&gin.Context{}))
}

func NewTelemetryManager() *Manager {
	mgr := &Manager{
		JobCron: cron.New(),
		JobSpec: "0 * * * * *", // interval hour
	}
	err := mgr.JobCron.AddJob(mgr.JobSpec, &autoReportTelemetryDataHandler{TelemetryManager: mgr})
	framework.Assert(err != nil) //The JobSpec string will not fail, and the program should exit if it does
	go func() {
		time.Sleep(5 * time.Second) //wait db client ready
		mgr.JobCron.Start()
		defer mgr.JobCron.Stop()
		select {}
	}()
	return mgr
}

// Report Interface for reporting data that can be called by regularly scheduled tasks
func (t *Manager) Report(ctx context.Context) error {
	var err error
	var enable bool
	var response *http.Response
	var data *structs.EnterpriseManagerInfo

	//The called function has already output the error log and converted the error,you can return directly
	err, enable = t.isEnabledTelemetry(ctx)
	if err != nil {
		return err
	}

	//If telemetry is closed, there is no need to report information to the telemeter service, just return
	if !enable {
		framework.LogWithContext(ctx).Info("system telemetry disable,no need to report information to the telemetry service")
		return nil
	}

	//Generate data to be reported
	//The called function has already output the error log and converted the error,you can return directly
	data, err = t.generateData(ctx)
	if err != nil {
		return err
	}

	framework.Assert(data != nil)
	response, err = util.PostJSON(constants.TelemetryAPIEndpoint, map[string]interface{}{"EnterpriseManager": data}, nil)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("post telemetry data to %s error %v", constants.TelemetryAPIEndpoint, err)
		return errors.NewErrorf(TelemetryPushDataToServiceError, "post telemetry data to %s error %v", constants.TelemetryAPIEndpoint, err)
	}

	// We don't even want to know any response body. Just close it.ignore error messages.
	_ = response.Body.Close()

	if response.StatusCode != http.StatusOK {
		framework.LogWithContext(ctx).Errorf("post telemetry data to %s error, http code: %d", constants.TelemetryAPIEndpoint, response.StatusCode)
		return errors.NewErrorf(TelemetryPushDataToServiceError, "post telemetry data to %s error,http code: %d", constants.TelemetryAPIEndpoint, response.StatusCode)
	}
	framework.LogWithContext(ctx).Infof("report telemetry data to %s successful", constants.TelemetryAPIEndpoint)
	return nil
}

//generateData Generate all data reported by the system
func (t *Manager) generateData(ctx context.Context) (*structs.EnterpriseManagerInfo, error) {
	var err error
	data := &structs.EnterpriseManagerInfo{}

	//generate enterprise manager info
	data.CreateTime = time.Now()    //TODO
	data.ID = uuidutil.GenerateID() //TODO
	data.Version.BuildTime = versioninfo.TiUniManagerBuildTS
	data.Version.GitBranch = versioninfo.TiUniManagerGitHash
	data.Version.GitHash = versioninfo.TiUniManagerGitHash
	data.Version.Version = versioninfo.TiUniManagerReleaseVersion

	//generate enterprise manager instance information
	data.EmNodes, err = t.generateEMNodeInfo(ctx)
	if err != nil {
		return nil, err
	}

	//generate enterprise manager api access statistics.
	data.EmAPICount, err = t.generateEMAccessStatistics(ctx)
	if err != nil {
		return nil, err
	}

	//generate enterprise manager management hosts.
	data.Nodes, err = t.generateEMManagementHostsInfo(ctx)
	if err != nil {
		return nil, err
	}

	//generate enterprise manager management products & instances.
	data.Clusters, err = t.generateEMManagementClusters(ctx)
	if err != nil {
		return nil, err
	}

	framework.LogWithContext(ctx).Infof("generate telemetry data success")
	return data, nil
}

//generateEMNodeInfo generate enterprise manager instance information
//The current enterprise manager is a single-machine system, and only information about a single machine is available
func (t *Manager) generateEMNodeInfo(ctx context.Context) (nodes []structs.NodeInfo, err error) {
	var node structs.NodeInfo

	//os information
	node.Os.Platform, node.Os.Family, node.Os.Version, err = arch.OSInfo(ctx)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get enterprise manager host os information error: %v", err)
		return nil, errors.NewErrorf(TelemetryGenerateEnterpriseNodeInfoDataError, "get enterprise manager host os information error: %v", err)
	}

	//cpu information
	node.Cpus.Num, node.Cpus.Sockets, node.Cpus.Cores, node.Cpus.Model, node.Cpus.HZ, err = arch.GetCpuInfo(ctx)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get enterprise manager host cpu information error: %v", err)
		return nil, errors.NewErrorf(TelemetryGenerateEnterpriseNodeInfoDataError, "get enterprise manager host cpu information error: %v", err)
	}

	//load information
	_, node.Loadavg15, err = arch.GetLoadavg(ctx)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get enterprise manager host load information error: %v", err)
		return nil, errors.NewErrorf(TelemetryGenerateEnterpriseNodeInfoDataError, "get enterprise manager host load information error: %v", err)
	}

	//memory information
	node.Memory.Total, node.Memory.Available, err = arch.GetVirtualMemory(ctx)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get enterprise manager host memory information error: %v", err)
		return nil, errors.NewErrorf(TelemetryGenerateEnterpriseNodeInfoDataError, "get enterprise manager host memory information error: %v", err)
	}

	//disk information(TBD)

	nodes = append(nodes, node)

	framework.LogWithContext(ctx).Infof("get enterprise manager host information sucess")
	return nodes, nil
}

//generateEMAccessStatistics generate enterprise manager management hosts,data are obtained from prometheus
func (t *Manager) generateEMAccessStatistics(ctx context.Context) (usages []structs.EnterpriseManagerFeatureUsage, er error) {
	var err error
	var value map[string]uint32
	var usage structs.EnterpriseManagerFeatureUsage

	configRW := models.GetConfigReaderWriter()
	config, err := configRW.GetConfig(ctx, constants.ConfigPrometheusAddress)
	if err != nil {
		framework.LogWithContext(ctx).Warningf("query %s from system config error %v", constants.ConfigPrometheusAddress, err)
		return usages, errors.NewErrorf(GetSystemConfigError, "query %s from system config error %v", constants.ConfigPrometheusAddress, err)
	}
	for _, item := range constants.EMMetrics {
		promQL := fmt.Sprintf("%s%s{} by (%s)", constants.OpenAPIMetricsPrefix, item, constants.ComponentIDOpenAPIServer)
		value, err = t.execPQL(ctx, time.Now(), time.Now(), promQL, string(constants.ComponentIDOpenAPIServer), config.ConfigValue, 100*time.Microsecond)
		if err == nil {
			//Only the name and number of accesses to the API are currently reported
			usage.APIName = string(item)
			for _, v := range value {
				usage.Count += v
			}
			usages = append(usages, usage)
		} else {
			framework.LogWithContext(ctx).Warningf("generate enterprise manager access statistics, "+
				"query %s from prometheus %s error %v", promQL, config.ConfigValue, err)
		}
	}
	framework.LogWithContext(ctx).Infof("generate enterprise manager access statistics complete")
	return usages, nil
}

//generateEMManagementHostsInfo generate enterprise manager management hosts,Dependent on resource manager
func (t *Manager) generateEMManagementHostsInfo(ctx context.Context) (nodes []structs.NodeInfo, er error) {

	rw := models.GetResourceReaderWriter()
	hosts, _, err := rw.Query(ctx, &structs.Location{}, &structs.HostFilter{}, 0, math.MaxInt32)
	if err != nil {
		framework.LogWithContext(ctx).Warningf("query hosts from database error %v", err)
		return nil, errors.NewErrorf(TelemetryGenerateEMManagementHostsInfoDataError, "query hosts from database error %v", err)
	}
	//Currently, report the host ID, Memory, CPU Core, Arch,
	//and then decide whether to add more metrics according to the operation situation.
	for _, item := range hosts {
		node := structs.NodeInfo{
			ID:     item.ID,
			Memory: structs.MemoryInfo{Total: uint64(item.Memory)},
			Cpus:   structs.CPUInfo{Cores: item.CpuCores},
			Os:     structs.OSInfo{Platform: item.Arch},
		}
		nodes = append(nodes, node)
	}

	framework.LogWithContext(ctx).Infof("generate enterprise manager management hosts information successful")
	return nodes, nil
}

//generateEMManagementClusters generate enterprise manager management products & instances,Dependent on cluster manager
func (t *Manager) generateEMManagementClusters(ctx context.Context) (clusters []structs.ClusterInfoForTelemetry, err error) {

	filter := management.Filters{}
	pr := structs.PageRequest{PageSize: math.MaxInt32}
	result, _, err := models.GetClusterReaderWriter().QueryMetas(ctx, filter, pr)
	if err != nil {
		framework.LogWithContext(ctx).Warningf("query clusters from database error %v", err)
		return clusters, errors.NewErrorf(TelemetryGenerateEMManagementClustersInfoDataError, "query clusters from database error %v", err)
	}

	//Populate the reported data structure with data from the database
	for _, item := range result {
		info := structs.ClusterInfoForTelemetry{
			ID:              item.Cluster.ID,
			Exclusive:       item.Cluster.Exclusive,
			CpuArchitecture: string(item.Cluster.CpuArchitecture),
			Status:          item.Cluster.Status,
			Type:            item.Cluster.Type,
			CreateTime:      item.Cluster.CreatedAt,
			DeleteTime:      item.Cluster.DeletedAt.Time,
			Version:         structs.Version{Version: item.Cluster.Version},
		}

		for _, value := range item.Instances {
			tmp := structs.InstanceInfoForTelemetry{
				ID:         value.ID,
				Type:       value.Type,
				Version:    item.Cluster.Version,
				Status:     value.Status,
				CreateTime: value.CreatedAt,
				DeleteTime: value.DeletedAt.Time,
				// TODO Spec:""
				// TODO Regions
			}
			info.Instances = append(info.Instances, tmp)
		}
		clusters = append(clusters, info)
	}
	framework.LogWithContext(ctx).Warningf("generate enterprise manager management clusters information successful")
	return clusters, nil
}

//isEnabledTelemetry check whether telemetry enabled.
func (t *Manager) isEnabledTelemetry(ctx context.Context) (error, bool) {
	configRW := models.GetConfigReaderWriter()
	config, err := configRW.GetConfig(ctx, constants.ConfigTelemetrySwitch)
	if err != nil {
		framework.LogWithContext(ctx).Warningf("query %s from system config error %v", constants.ConfigTelemetrySwitch, err)
		return errors.NewErrorf(GetSystemConfigError, "query %s from system config error %v", constants.ConfigTelemetrySwitch, err), false
	}
	return nil, config.ConfigValue == constants.TelemetrySwitchEnable
}

//execPQL Query data from prometheus
//Only applicable to query the summary value of a single indicator, for example: sum(http_requests_total) by (labelName)
func (t *Manager) execPQL(ctx context.Context, startTime, endTime time.Time, promQL, labelName, addr string, timeout time.Duration) (rt map[string]uint32, er error) {
	var value model.Value
	var warnings client.Warnings
	promClient, err := api.NewClient(api.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	promQLAPI := client.NewAPI(promClient)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	r := client.Range{
		Start: startTime,
		End:   endTime,
		Step:  time.Second,
	}

	// Add retry to avoid network error.
	for i := 0; i < 5; i++ {
		value, warnings, err = promQLAPI.QueryRange(ctx, promQL, r)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		return nil, err
	}

	if len(warnings) > 0 {
		framework.LogWithContext(ctx).Warningf("execPQL %s have warnings: %v", promQL, warnings)
	}

	result := make(map[string]uint32)

	switch value.Type() {
	case model.ValVector:
		matrix := value.(model.Vector)
		for _, item := range matrix {
			label := string(item.Metric[model.LabelName(labelName)])
			result[label] = uint32(item.Value)
		}
	}
	return result, err
}

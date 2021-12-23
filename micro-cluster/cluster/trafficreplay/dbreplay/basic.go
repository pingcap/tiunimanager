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
 * limitations under the License.                                             *
 ******************************************************************************/

/**
 * @Author: guobob
 * @Description:
 * @File:  basic.go
 * @Version: 1.0.0
 * @Date: 2021/12/7 15:26
 */

package dbreplay

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/common/cmd"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/common/utils"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	resourcemanager "github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/errors"
	"github.com/yalp/jsonpath"
	"strconv"
	"strings"
)

const (
	ProductHostInfo int = iota
	SimulationHostInfo
	AllHostInfo
)

type HostKey struct {
	HostID string
	Host   string
	Port   int
	//host:port
	Key    string
}



type TrafficReplayBasic struct {
	ID                  string
	// ClusterIDs[0] save product cluster ID
	// ClusterIDs[1] save simulation cluster ID
	ClusterIDs          []string
	// ClusterAggregations[0] save product cluster aggregation
	// ClusterAggregations[1] save simulation cluster aggregation
	ClusterAggregations []*domain.ClusterAggregation

	//HostKyePairs[0] save product tidb HostKey
	//HostKyePairs[1] save simulation tidb HostKey
	HostKyePairs [][]*HostKey
	// key [hostIP:port]
	// value [*structs.HostInfo]
	Hosts    map[string]*structs.HostInfo
	//key [ hostIP:port ]
	//value [tidb server deploy-path ]
	RootPath map[string]string
	Status   string
}



//GetClusterAggregations : Get Cluster Aggregations by cluster ID
func (b *TrafficReplayBasic) GetClusterAggregations(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin get cluster info %s,%s", b.ClusterIDs[0], b.ClusterIDs[1])
	defer framework.LogWithContext(ctx).Infof("end get cluster info")
	var err error

	//get production cluster aggregation
	b.ClusterAggregations[0], err = domain.ClusterRepo.Load(ctx, b.ClusterIDs[0])
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster aggregation by cluster ID %s failed, %v",
			b.ClusterIDs[0], err)
		return err
	}
	if b.ClusterAggregations[0] == nil {
		errStr := fmt.Sprintf("load cluster aggregation by cluster ID %s  failed, aggregation is nil ",
			b.ClusterIDs[0])
		framework.LogWithContext(ctx).Errorf(errStr)
		return fmt.Errorf(errStr)
	}

	//get simulation cluster aggregation
	b.ClusterAggregations[1], err = domain.ClusterRepo.Load(ctx, b.ClusterIDs[1])
	if err != nil {
		framework.LogWithContext(ctx).Errorf("load cluster aggregation by cluster ID %s failed, %v",
			b.ClusterIDs[1], err)
		return err
	}
	if b.ClusterAggregations[1] == nil {
		errStr := fmt.Sprintf("load cluster aggregation  by cluster ID %s failed, aggregation is nil ",
			b.ClusterIDs[1])
		framework.LogWithContext(ctx).Errorf(errStr)
		return fmt.Errorf(errStr)
	}
	return nil
}

//GetHostByID : Get HostInfo by host ID
func GetHostByID(ctx context.Context, hostID string) (*structs.HostInfo, error) {

	filter := &structs.HostFilter{
		HostID: hostID,
	}
	page :=&structs.PageRequest{
		Page: 1,
		PageSize: 10,
	}
	rm := resourcemanager.GetResourcePool()
	//get host info  by host-id
	hostInfo,err := rm.QueryHosts(ctx, filter, page)
	if err != nil {
		return nil, err
	}
	if len(hostInfo) ==0{
		framework.LogWithContext(ctx).Errorf("get hostInfo fail ,hostID[%v]",hostID)
		return nil,fmt.Errorf("get hostInfo fail ,hostID[%v]",hostID)
	}
	return &hostInfo[0], nil
}

//GetProductTiDBServerHosts : get Product TiDBServer Hosts
func (b *TrafficReplayBasic) GetProductTiDBServerHosts(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin get product tidb server hosts ,%v", b.ClusterIDs[0])
	defer framework.LogWithContext(ctx).Infof("end get product tidb server hosts ")
	var err error
	var hostInfo *structs.HostInfo

	//get tidb server host information for production environment
	for _, v := range b.ClusterAggregations[0].CurrentComponentInstances {
		if v.ComponentType.ComponentType != string(constants.EMProductIDTiDB) {
			continue
		}

		hostKey:=&HostKey{
			Host:v.Host,
			HostID:v.HostId,
			Port:v.PortList[0],
			Key :v.Host+strconv.Itoa(v.PortList[0]),
		}
		b.HostKyePairs[0]=append(b.HostKyePairs[0],hostKey)

		hostInfo, err = GetHostByID(ctx, v.HostId)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get host info by hostID %v fail,%v", v.HostId,err)
			return err
		}
		b.Hosts[hostKey.Key] =  hostInfo

	}
	return nil
}

//GetSimulationTiDBServerHosts : get Simulation TiDBServer Hosts
func (b *TrafficReplayBasic) GetSimulationTiDBServerHosts(ctx context.Context) error {
	var err error
	var hostInfo *structs.HostInfo
	framework.LogWithContext(ctx).Infof("begin get Simulation tidb server hosts ,%v", b.ClusterIDs[1])
	defer framework.LogWithContext(ctx).Infof("end get Simulation tidb server hosts ")
	//get tidb server host information for simulation environment
	for _, v := range b.ClusterAggregations[1].CurrentComponentInstances {
		if v.ComponentType.ComponentType != string(constants.EMProductIDTiDB) {
			continue
		}

		hostKey:=&HostKey{
			Host:v.Host,
			HostID:v.HostId,
			Port:v.PortList[0],
			Key :v.Host+strconv.Itoa(v.PortList[0]),
		}
		b.HostKyePairs[1]=append(b.HostKyePairs[1],hostKey)

		hostInfo, err = GetHostByID(ctx, v.HostId)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("get host info by hostID %v fail,%v", v.HostId,err)
			return err
		}
		b.Hosts[hostKey.Key]=hostInfo
	}
	return nil
}

func (b *TrafficReplayBasic)GetKeyFromJson(ctx context.Context,js string, key []string, depth []int) ([]interface{}, error) {
	framework.LogWithContext(ctx).Infof("begin parse json ,key[%v],depth[%v]",key,depth)
	defer framework.LogWithContext(ctx).Infof("end parse json ,key[%v],depth[%v]",key,depth)

	if len(key) != len(depth) || len(key) == 0 {
		return nil, errors.New(fmt.Sprintf("param len  invalid,%v-%v", len(key), len(depth)))
	}

	res := make([]interface{}, len(key))
	var data interface{}
	for k := range key {
		f := fmt.Sprintf("$" + strings.Repeat(".", depth[k]) + key[k])
		filter, err := jsonpath.Prepare(f)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(utils.Slice(js), &data); err != nil {
			return nil, err
		}
		out, err := filter(data)
		if err != nil {
			return nil, err
		}
		res[k] = out
	}
	return res, nil
}

func (b *TrafficReplayBasic)GeneratePairData(ctx context.Context,input []interface{}, keyPos,
	valPos int) (map[string]string, error) {

	framework.LogWithContext(ctx).Infof("begin get data from interface %v ,key %v,value %v",
		input,keyPos,valPos )
	defer framework.LogWithContext(ctx).Infof("end get data from interface %v ,key %v,value %v",
		input,keyPos,valPos )

	res := make(map[string]string)

	ks, ok := input[keyPos].([]interface{})
	if !ok {
		return nil, fmt.Errorf(" param invalid ,%v", input)
	}
	vs, ok := input[valPos].([]interface{})
	if !ok {
		return nil, fmt.Errorf(" param invalid ,%v", input)
	}

	if len(ks) != len(vs) {
		return nil, fmt.Errorf("key value number is not equal，%v-%v", len(ks), len(vs))
	}

	for k := range ks {
		key, ok := ks[k].(string)
		if !ok {
			return nil, fmt.Errorf("key type is not string ,%T", key)
		}
		val, ok := vs[k].(string)
		if !ok {
			return nil, fmt.Errorf("key type is not string ,%T", val)
		}
		res[key] = val
	}

	return res, nil

}

func (b *TrafficReplayBasic)FillRootPath(ctx context.Context,m map[string]string)  {
	for k,v := range m {
		b.RootPath[k]=v
	}
	return
}


func(b *TrafficReplayBasic)AnalysisDisplayOutput(ctx context.Context,
	js string,key[]string,depth []int,keyPos,valPos int) error{
	res , err := b.GetKeyFromJson(ctx ,js , key, depth)
	if err !=nil{
		return err
	}
	m ,err := b.GeneratePairData(ctx,res,keyPos,valPos)
	if err !=nil{
		return err
	}

	b.FillRootPath(ctx,m)
	return nil
}


func (b *TrafficReplayBasic)GetTiDBServerDeployPath(ctx context.Context,pos int ) error{
	clusterAggregation := b.ClusterAggregations[pos]
	flags := []string{"-R " , " tidb ", " --format " , "json"}
	resp,err :=secondparty.Manager.ClusterDisplay(ctx,secondparty.ClusterComponentTypeStr,
		clusterAggregation.Cluster.ClusterName, 10, flags)
	/*
	resp ,err := secondparty.SecondParty.MicroSrvTiupDisplay(ctx , secondparty.ClusterComponentTypeStr,
		clusterAggregation.Cluster.ClusterName, 10, flags)
	 */

	if err !=nil{
		return err
	}
	err = b.AnalysisDisplayOutput(ctx,resp.DisplayRespString,[]string{"id","deploy_dir"},[]int{2,2},0,1)
	if err !=nil{
		return err
	}
	return nil
}

//GetRootPath : get root path from cluster Aggregations
//TODO The default root-path  is tidb server  deploy path now
func (b *TrafficReplayBasic) GetRootPath(ctx context.Context) error  {
	framework.LogWithContext(ctx).Infof("begin set Root Path")
	defer framework.LogWithContext(ctx).Infof("end set Root Path")
	err := b.GetTiDBServerDeployPath(ctx,0)
	if err !=nil {
		return err
	}
	err = b.GetTiDBServerDeployPath(ctx,1)
	if err !=nil {
		return err
	}
	return nil
}

//GetTidbServerHosts : Get host information from production and simulation environment clusters
func (b *TrafficReplayBasic) GetTidbServerHosts(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin get tidb server host")
	defer framework.LogWithContext(ctx).Infof("begin get tidb server host")
	var err error
	err = b.GetProductTiDBServerHosts(ctx)
	if err != nil {
		return err
	}
	err = b.GetSimulationTiDBServerHosts(ctx)
	if err != nil {
		return err
	}
	return nil
}

//CheckParam : check parameters valid
func (b *TrafficReplayBasic) CheckParam(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin check param ")
	defer framework.LogWithContext(ctx).Infof("end check param ")

	var err error

	//The traffic replay run requires two clusters, a production environment
	//cluster and a simulation environment cluster, with the agreement that
	//the first element of the slice is the production cluster and the second
	//element of the slice is the simulation cluster
	if len(b.ClusterIDs) != 2 {
		return fmt.Errorf("param clusterids len  is invald %v", len(b.ClusterIDs))
	}

	//get cluster aggregations by clusterID
	err = b.GetClusterAggregations(ctx)
	if err != nil {
		return err
	}

	//get tidb server host by hostID
	err = b.GetTidbServerHosts(ctx)
	if err != nil {
		return err
	}

	productHost := b.HostKyePairs[0]
	simulationHost := b.HostKyePairs[1]
	//Check that the number of tidb servers in the production and simulation environments are the same
	if len(productHost) != len(simulationHost) || len(productHost) == 0 {
		return fmt.Errorf("the number of tidb servers in the production environment "+
			"and the number of tidb servers in the "+
			"emulation environment are not same. %v-%v", len(productHost), len(simulationHost))
	}

	return nil
}



//GetHostPair : get hostInfo
// resType 0: get product hostInfo
// resType 1: get simulation hostInfo
// resType 2: get product hostInfo && simulation hostInfo
func (b *TrafficReplayBasic) GetHostPair(ctx context.Context,pos int ,resType int )([]*structs.HostInfo  ,error ){
	var res []*structs.HostInfo
	if resType ==ProductHostInfo || resType ==AllHostInfo {
		productHostKey := b.HostKyePairs[0]
		productHost,ok := b.Hosts[productHostKey[pos].Key]
		if !ok {
			return nil,fmt.Errorf("get host info fail ,tidb-id %s",productHostKey[pos].Key)
		}
		res = append(res,productHost)
	}
	if resType ==SimulationHostInfo || resType ==AllHostInfo {
		simulationHostKey := b.HostKyePairs[1]
		simulationHost,ok := b.Hosts[simulationHostKey[pos].Key]
		if !ok {
			return nil, fmt.Errorf("get host info fail ,tidb-id %s", simulationHostKey[pos].Key)
		}
		res = append(res,simulationHost)
	}
	return res,nil
}



//CheckNetNormal : check that the production tidb server and
//the emulated tidb server network are connected
func (b *TrafficReplayBasic) CheckNetNormal(ctx context.Context) error {
	framework.LogWithContext(ctx).Infof("begin check net normal")
	defer framework.LogWithContext(ctx).Infof("end check net normal")
	productHostKey := b.HostKyePairs[0]

	for k := range productHostKey {
		res,err := b.GetHostPair(ctx,k,AllHostInfo)
		if err !=nil{
			return err
		}
		config := cmd.NewSessionConfig(res[0].UserName, res[0].Passwd,
			res[0].IP, 22, 0)

		r, err := cmd.NewRemoteSession(config)
		if err != nil {
			return err
		}

		//Check if the production tidb server and the emulated tidb server
		//are connected by using the ping command
		command := fmt.Sprintf("ping -c 2 %v", res[1].IP)
		err = r.Exec(command, cmd.WaitCmdEnd)
		errClose := r.Close()
		if errClose != nil {
			framework.LogWithContext(ctx).Warn("close ssh connect fail , " + errClose.Error())
		}
		if err != nil {
			return err
		}
	}

	return nil
}


//SetStatus : set traffic replay task status
func (b *TrafficReplayBasic) SetStatus(ctx context.Context, status string) {

	framework.LogWithContext(ctx).Infof("begin set status , before change status %v", b.Status)
	defer framework.LogWithContext(ctx).Infof("end set status ,%v-%v", b.Status, status)
	b.Status = status
}

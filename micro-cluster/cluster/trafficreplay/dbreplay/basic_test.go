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
 * @File:  basic_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/10 15:32
 */

package dbreplay

import (
	"context"
	"github.com/agiledragon/gomonkey"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/cmd"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

/*
func TestGetHostByID_With_GetHostByID_fail(t *testing.T) {
	ctx := context.Background()
	hostID := "host-id-test"

	rm:=new(resourcemanager.ResourcePool)
	patch := gomonkey.ApplyFunc( resourcemanager.GetResourcePool(),func () *resourcemanager.ResourcePool  {
		return rm
	})
	defer patch.Reset()

	err1:= errors.New("get Host Info fail ")
	patches := gomonkey.ApplyMethod(reflect.TypeOf(rm), "QueryHosts",
		func(_ *resourcemanager.ResourcePool,ctx context.Context,
			filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, err error){
			return nil,err1
		})
	defer patches.Reset()

	filter := &structs.HostFilter{
		HostID: hostID,
	}
	page :=&structs.PageRequest{
		Page: 1,
		PageSize: 10,
	}
	host ,err := rm.QueryHosts(ctx,filter,page)

	ast := assert.New(t)
	ast.Nil(host)
	ast.Equal(err,err1)

}


func TestGetHostByIDWith_GetHostByID_succ(t *testing.T) {

	ctx := context.Background()
	hostID := "host-id-test"

	rm:=new(resourcemanager.ResourcePool)
	patch := gomonkey.ApplyFunc( resourcemanager.GetResourcePool(),func () *resourcemanager.ResourcePool  {
		return rm
	})
	defer patch.Reset()


	patches := gomonkey.ApplyMethod(reflect.TypeOf(rm), "QueryHosts",
		func(_ *resourcemanager.ResourcePool,ctx context.Context,
			filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, err error){
			return []structs.HostInfo{
				{},
			},nil
		})
	defer patches.Reset()

	filter := &structs.HostFilter{
		HostID: hostID,
	}
	page :=&structs.PageRequest{
		Page: 1,
		PageSize: 10,
	}
	host ,err := rm.QueryHosts(ctx,filter,page)

	ast := assert.New(t)
	ast.NotNil(host)
	ast.Nil(err)
}
*/

func TestTrafficReplayBasic_CheckNetNormal_with_NewRemoteSession_fail(t *testing.T){
	ctx := context.Background()
	b := new(TrafficReplayBasic)
	b.Hosts=make(map[string]*structs.HostInfo)

	h1 := &HostKey{
		Host: "172.16.4.155",
		HostID: "test-hostid",
		Port:4001,
		Key: "172.16.4.155:4001",
	}

	h2 := &HostKey{
		Host: "172.16.4.116",
		HostID: "test-hostid",
		Port:4001,
		Key: "172.16.4.116:4001",
	}
	productHost := &structs.HostInfo{
		UserName: "root",
		IP:"172.16.4.155",
		Passwd:"******",
	}
	simulationHost := &structs.HostInfo{
			UserName: "root",
			IP:"172.16.4.116",
			Passwd:"******",
	}

	b.Hosts[h1.Key]=productHost
	b.Hosts[h2.Key]=simulationHost

	b.HostKyePairs = make([][]*HostKey,2,2)
	b.HostKyePairs[0]=[]*HostKey{
		h1,
	}
	b.HostKyePairs[1]=[]*HostKey{
		h2,
	}


	//	mockCtl := gomock.NewController(t)
//	remoteSess := cmd.NewMockExecSession(mockCtl)

	err1 := errors.New("connect remote server fail")
	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return nil,err1
	})
	defer patch.Reset()

	err := b.CheckNetNormal(ctx)

	assert.New(t).Equal(err,err1)

}


func TestTrafficReplayBasic_CheckNetNormal_with_Exec_fail(t *testing.T){
	ctx := context.Background()
	b := new(TrafficReplayBasic)
	b.Hosts=make(map[string]*structs.HostInfo)

	h1 := &HostKey{
		Host: "172.16.4.155",
		HostID: "test-hostid",
		Port:4001,
		Key: "172.16.4.155:4001",
	}

	h2 := &HostKey{
		Host: "172.16.4.116",
		HostID: "test-hostid",
		Port:4001,
		Key: "172.16.4.116:4001",
	}
	productHost := &structs.HostInfo{
		UserName: "root",
		IP:"172.16.4.155",
		Passwd:"******",
	}
	simulationHost := &structs.HostInfo{
		UserName: "root",
		IP:"172.16.4.116",
		Passwd:"******",
	}

	b.Hosts[h1.Key]=productHost
	b.Hosts[h2.Key]=simulationHost
	b.HostKyePairs = make([][]*HostKey,2,2)
	b.HostKyePairs[0]=[]*HostKey{
		h1,
	}
	b.HostKyePairs[1]=[]*HostKey{
		h2,
	}


	err1 := errors.New("exec command fail")
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	remoteSess := cmd.NewMockExecSession(mockCtl)
	remoteSess.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(err1)
	remoteSess.EXPECT().Close().Return(nil)

	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return remoteSess,nil
	})
	defer patch.Reset()

	err := b.CheckNetNormal(ctx)

	assert.New(t).Equal(err,err1)

}

func TestTrafficReplayBasic_CheckNetNormal_with_Exec_succ_and_close_fail(t *testing.T){
	ctx := context.Background()
	b := new(TrafficReplayBasic)
	b.Hosts=make(map[string]*structs.HostInfo)

	h1 := &HostKey{
		Host: "172.16.4.155",
		HostID: "test-hostid",
		Port:4001,
		Key: "172.16.4.155:4001",
	}

	h2 := &HostKey{
		Host: "172.16.4.116",
		HostID: "test-hostid",
		Port:4001,
		Key: "172.16.4.116:4001",
	}
	productHost := &structs.HostInfo{
		UserName: "root",
		IP:"172.16.4.155",
		Passwd:"******",
	}
	simulationHost := &structs.HostInfo{
		UserName: "root",
		IP:"172.16.4.116",
		Passwd:"******",
	}

	b.Hosts[h1.Key]=productHost
	b.Hosts[h2.Key]=simulationHost

	b.HostKyePairs = make([][]*HostKey,2,2)
	b.HostKyePairs[0]=[]*HostKey{
		h1,
	}
	b.HostKyePairs[1]=[]*HostKey{
		h2,
	}

	err1 := errors.New("close remote connect fail ")
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	remoteSess := cmd.NewMockExecSession(mockCtl)
	remoteSess.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)

	remoteSess.EXPECT().Close().Return(err1)

	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return remoteSess,nil
	})
	defer patch.Reset()

	err := b.CheckNetNormal(ctx)

	assert.New(t).Nil(err)

}

func Test_CheckParam_with_clusterIDs_invalid(t *testing.T){
	b := new(TrafficReplayBasic)
	b.ClusterIDs=make([]string,1,1)
	b.ClusterIDs[0]="test-cluster-id"
	ctx := context.Background()
	err := b.CheckParam(ctx)
	assert.New(t).NotNil(err)

}

func Test_CheckParam_with_GetClusterAggregations_fail(t *testing.T){
	b := new(TrafficReplayBasic)
	b.ClusterIDs=make([]string,2,2)
	b.ClusterIDs[0]="test-cluster-id1"
	b.ClusterIDs[1]="test-cluster-id2"

	err1:=errors.New("get cluster aggregations fail")
	patches := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetClusterAggregations",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return err1
		})
	defer patches.Reset()

	ctx := context.Background()
	err := b.CheckParam(ctx)
	assert.New(t).Equal(err,err1)

}

func Test_CheckParam_with_GetTidbServerHosts_fail(t *testing.T){
	b := new(TrafficReplayBasic)
	b.ClusterIDs=make([]string,2,2)
	b.ClusterIDs[0]="test-cluster-id1"
	b.ClusterIDs[1]="test-cluster-id2"


	patches := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetClusterAggregations",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return nil
		})
	defer patches.Reset()

	err1:=errors.New("get tidb server host fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetTidbServerHosts",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return err1
		})
	defer patches1.Reset()

	ctx := context.Background()
	err := b.CheckParam(ctx)
	assert.New(t).Equal(err,err1)

}


func Test_CheckParam_with_Check_host_fail(t *testing.T){
	b := new(TrafficReplayBasic)
	b.ClusterIDs=make([]string,2,2)
	b.ClusterIDs[0]="test-cluster-id1"
	b.ClusterIDs[1]="test-cluster-id2"
	b.Hosts=make(map[string]*structs.HostInfo)

	h1 := &HostKey{
		Host: "172.16.4.155",
		HostID: "test-hostid",
		Port:4001,
		Key: "172.16.4.155:4001",
	}


	productHost := &structs.HostInfo{
		UserName: "root",
		IP:"172.16.4.155",
		Passwd:"******",
	}

	b.Hosts[h1.Key]=productHost
	b.HostKyePairs = make([][]*HostKey,2,2)
	b.HostKyePairs[0]=[]*HostKey{
		h1,
	}
	b.HostKyePairs[1]=[]*HostKey{

	}

	patches := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetClusterAggregations",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return nil
		})
	defer patches.Reset()

	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetTidbServerHosts",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return nil
		})
	defer patches1.Reset()

	ctx := context.Background()
	err := b.CheckParam(ctx)
	assert.New(t).NotNil(err)

}

func Test_CheckParam_succ(t *testing.T){
	b := new(TrafficReplayBasic)
	b.ClusterIDs=make([]string,2,2)
	b.ClusterIDs[0]="test-cluster-id1"
	b.ClusterIDs[1]="test-cluster-id2"
	b.Hosts=make(map[string]*structs.HostInfo)

	h1 := &HostKey{
		Host: "172.16.4.155",
		HostID: "test-hostid",
		Port:4001,
		Key: "172.16.4.155:4001",
	}

	h2 := &HostKey{
		Host: "172.16.4.116",
		HostID: "test-hostid",
		Port:4001,
		Key: "172.16.4.116:4001",
	}
	productHost := &structs.HostInfo{
		UserName: "root",
		IP:"172.16.4.155",
		Passwd:"******",
	}
	simulationHost := &structs.HostInfo{
		UserName: "root",
		IP:"172.16.4.116",
		Passwd:"******",
	}

	b.Hosts[h1.Key]=productHost
	b.Hosts[h2.Key]=simulationHost
	b.HostKyePairs = make([][]*HostKey,2,2)
	b.HostKyePairs[0]=[]*HostKey{
		h1,
	}
	b.HostKyePairs[1]=[]*HostKey{
		h2,
	}

	patches := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetClusterAggregations",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return nil
		})
	defer patches.Reset()

	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetTidbServerHosts",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return nil
		})
	defer patches1.Reset()

	ctx := context.Background()
	err := b.CheckParam(ctx)
	assert.New(t).Nil(err)

}

func Test_SetStatus(t *testing.T){
	ctx := context.Background()
	b := new(TrafficReplayBasic)

	b.SetStatus(ctx,"INIT")

	assert.New(t).Equal(b.Status, "INIT")
}

func Test_GetTidbServerHosts_with_GetProductTiDBServerHosts_fail(t *testing.T){
	b := new(TrafficReplayBasic)
	ctx:=context.Background()
	err1 := errors.New("get product tidb server hosts fail")
	patches := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetProductTiDBServerHosts",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return err1
		})
	defer patches.Reset()

	err := b.GetTidbServerHosts(ctx)

	assert.New(t).Equal(err,err1)
}


func Test_GetTidbServerHosts_with_GetSimulationTiDBServerHosts_fail(t *testing.T){
	b := new(TrafficReplayBasic)
	ctx:=context.Background()

	patches := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetProductTiDBServerHosts",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return nil
		})
	defer patches.Reset()

	err1 := errors.New("get simulation tidb server hosts fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetSimulationTiDBServerHosts",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return err1
		})
	defer patches1.Reset()

	err := b.GetTidbServerHosts(ctx)

	assert.New(t).Equal(err,err1)
}


func Test_GetTidbServerHosts_succ(t *testing.T){
	b := new(TrafficReplayBasic)
	ctx:=context.Background()

	patches := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetProductTiDBServerHosts",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return nil
		})
	defer patches.Reset()

	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(b), "GetSimulationTiDBServerHosts",
		func(_ *TrafficReplayBasic,ctx context.Context) error {
			return nil
		})
	defer patches1.Reset()

	err := b.GetTidbServerHosts(ctx)

	assert.New(t).Nil(err)
}

func Test_GetProductTiDBServerHosts_GetHostByID_fail (t *testing.T){
	b := new(TrafficReplayBasic)
	b.HostKyePairs = make([][]*HostKey,2)
	b.HostKyePairs[0]=make([]*HostKey,0)
	b.HostKyePairs[1]=make([]*HostKey,0)
	b.ClusterIDs=make([]string,2,2)
	b.ClusterIDs[0]="test-cluster-id1"
	b.ClusterIDs[1]="test-cluster-id2"
	ctx:=context.Background()
	b.ClusterAggregations = make([]*domain.ClusterAggregation,2,2)
	b.ClusterAggregations[0]=new(domain.ClusterAggregation)
	b.ClusterAggregations[0].CurrentComponentInstances = make([]*domain.ComponentInstance,2,2)
	c1 := new(domain.ComponentInstance)
	c1.ComponentType =new(knowledge.ClusterComponent)
	c1.ComponentType.ComponentType =string(constants.EMProductIDTiDB)
	c1.HostId="test-host-id1"
	c1.PortList=[]int{4001,4002}
	c2:= new(domain.ComponentInstance)
	c2.ComponentType =new(knowledge.ClusterComponent)
	c2.ComponentType.ComponentType =string(constants.EMProductIDTiDB)
	c2.HostId="test-host-id2"
	c2.PortList=[]int{4001,4002}
	b.ClusterAggregations[0].CurrentComponentInstances[0]=c1
	b.ClusterAggregations[0].CurrentComponentInstances[1]=c2
	err1 := errors.New("get host by host-id fail ")
	patch := gomonkey.ApplyFunc(GetHostByID, func(ctx context.Context, hostID string) (*structs.HostInfo, error) {
		return nil, err1
	})
	defer patch.Reset()

	err := b.GetProductTiDBServerHosts(ctx)
	assert.New(t).Equal(err,err1)

}

func Test_GetProductTiDBServerHosts_GetHostByID_succ (t *testing.T){
	b := new(TrafficReplayBasic)
	b.HostKyePairs = make([][]*HostKey,2)
	b.HostKyePairs[0]=make([]*HostKey,0)
	b.HostKyePairs[1]=make([]*HostKey,0)
	b.Hosts =make(map[string]*structs.HostInfo)
	b.ClusterIDs=make([]string,2,2)
	b.ClusterIDs[0]="test-cluster-id1"
	b.ClusterIDs[1]="test-cluster-id2"
	ctx:=context.Background()
	b.ClusterAggregations = make([]*domain.ClusterAggregation,2,2)
	b.ClusterAggregations[0]=new(domain.ClusterAggregation)
	b.ClusterAggregations[0].CurrentComponentInstances = make([]*domain.ComponentInstance,2,2)
	c1 := new(domain.ComponentInstance)
	c1.ComponentType =new(knowledge.ClusterComponent)
	c1.ComponentType.ComponentType =string(constants.EMProductIDTiDB)
	c1.HostId="test-host-id1"
	c1.PortList=[]int{4001,4002}
	c2:= new(domain.ComponentInstance)
	c2.ComponentType =new(knowledge.ClusterComponent)
	c2.ComponentType.ComponentType =string(constants.EMProductIDTiDB)
	c2.HostId="test-host-id2"
	c2.PortList=[]int{4001,4002}
	b.ClusterAggregations[0].CurrentComponentInstances[0]=c1
	b.ClusterAggregations[0].CurrentComponentInstances[1]=c2

	hostInfo:= new(structs.HostInfo)
	patch := gomonkey.ApplyFunc(GetHostByID, func(ctx context.Context, hostID string) (*structs.HostInfo, error) {
		return hostInfo, nil
	})
	defer patch.Reset()

	err := b.GetProductTiDBServerHosts(ctx)
	assert.New(t).Nil(err)

}

func Test_GetSimulationTiDBServerHosts_GetHostByID_fail (t *testing.T){
	b := new(TrafficReplayBasic)
	b.HostKyePairs = make([][]*HostKey,2)
	b.HostKyePairs[0]=make([]*HostKey,0)
	b.HostKyePairs[1]=make([]*HostKey,0)
	b.ClusterIDs=make([]string,2,2)
	b.ClusterIDs[0]="test-cluster-id1"
	b.ClusterIDs[1]="test-cluster-id2"
	ctx:=context.Background()
	b.ClusterAggregations = make([]*domain.ClusterAggregation,2,2)
	b.ClusterAggregations[1]=new(domain.ClusterAggregation)
	b.ClusterAggregations[1].CurrentComponentInstances = make([]*domain.ComponentInstance,2,2)
	c1 := new(domain.ComponentInstance)
	c1.ComponentType =new(knowledge.ClusterComponent)
	c1.ComponentType.ComponentType =string(constants.EMProductIDTiDB)
	c1.HostId="test-host-id1"
	c1.PortList=[]int{4001,4002}
	c2:= new(domain.ComponentInstance)
	c2.ComponentType =new(knowledge.ClusterComponent)
	c2.ComponentType.ComponentType =string(constants.EMProductIDTiDB)
	c2.HostId="test-host-id2"
	c2.PortList=[]int{4001,4002}
	b.ClusterAggregations[1].CurrentComponentInstances[0]=c1
	b.ClusterAggregations[1].CurrentComponentInstances[1]=c2
	err1 := errors.New("get host by host-id fail ")
	patch := gomonkey.ApplyFunc(GetHostByID, func(ctx context.Context, hostID string) (*structs.HostInfo, error) {
		return nil, err1
	})
	defer patch.Reset()

	err := b.GetSimulationTiDBServerHosts(ctx)
	assert.New(t).Equal(err,err1)

}

func Test_GetSimulationTiDBServerHosts_GetHostByID_succ (t *testing.T){
	b := new(TrafficReplayBasic)
	b.HostKyePairs = make([][]*HostKey,2)
	b.HostKyePairs[0]=make([]*HostKey,0)
	b.HostKyePairs[1]=make([]*HostKey,0)
	b.Hosts =make(map[string]*structs.HostInfo)
	b.ClusterIDs=make([]string,2,2)
	b.ClusterIDs[0]="test-cluster-id1"
	b.ClusterIDs[1]="test-cluster-id2"
	ctx:=context.Background()
	b.ClusterAggregations = make([]*domain.ClusterAggregation,2,2)
	b.ClusterAggregations[1]=new(domain.ClusterAggregation)
	b.ClusterAggregations[1].CurrentComponentInstances = make([]*domain.ComponentInstance,2,2)
	c1 := new(domain.ComponentInstance)
	c1.ComponentType =new(knowledge.ClusterComponent)
	c1.ComponentType.ComponentType =string(constants.EMProductIDTiDB)
	c1.HostId="test-host-id1"
	c1.PortList=[]int{4001,4002}
	c2:= new(domain.ComponentInstance)
	c2.ComponentType =new(knowledge.ClusterComponent)
	c2.ComponentType.ComponentType =string(constants.EMProductIDTiDB)
	c2.HostId="test-host-id2"
	c2.PortList=[]int{4001,4002}
	b.ClusterAggregations[1].CurrentComponentInstances[0]=c1
	b.ClusterAggregations[1].CurrentComponentInstances[1]=c2

	hostInfo:= new(structs.HostInfo)
	patch := gomonkey.ApplyFunc(GetHostByID, func(ctx context.Context, hostID string) (*structs.HostInfo, error) {
		return hostInfo, nil
	})
	defer patch.Reset()

	err := b.GetSimulationTiDBServerHosts(ctx)
	assert.New(t).Nil(err)

}

/*
func Test_GetClusterAggregations_first_Load_fail(t *testing.T){
	b := new(TrafficReplayBasic)
	b.ClusterIDs=make([]string,2,2)
	b.ClusterIDs[0]="test-cluster-id1"
	b.ClusterIDs[1]="test-cluster-id2"
	b.ClusterAggregations = make([]*domain.ClusterAggregation,2,2)
	ctx:=context.Background()

	err := b.GetClusterAggregations(ctx)
	fmt.Println(err)

}
 */
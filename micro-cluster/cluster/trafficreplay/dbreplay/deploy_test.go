/**
 * @Author: guobob
 * @Description:
 * @File:  dereplay_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/14 18:03
 */

package dbreplay

import (
	"context"
	"github.com/agiledragon/gomonkey"
	"github.com/golang/mock/gomock"
	"github.com/pingcap-inc/tiem/common/cmd"
	"github.com/pingcap-inc/tiem/common/mutualtrust"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_CreateMutualTrust_SelectSecret_fail (t *testing.T){
	ctx:=context.Background()
	d := new(TrafficReplayDeploy)
	d.Basic=new(TrafficReplayBasic)
	b := d.Basic
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

	err1:=errors.New("get key fail")
	patches := gomonkey.ApplyMethod(reflect.TypeOf(&mutualtrust.Host{}), "SelectSecret",
		func(_ *mutualtrust.Host)  (string, error) {
			return "",err1
		})
	defer patches.Reset()

	err := d.CreateMutualTrust(ctx)

	assert.New(t).Equal(err,err1)
}

func Test_CreateMutualTrust_SecretWrite_fail (t *testing.T){
	ctx:=context.Background()
	d := new(TrafficReplayDeploy)
	d.Basic=new(TrafficReplayBasic)
	b := d.Basic
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

	patches := gomonkey.ApplyMethod(reflect.TypeOf(&mutualtrust.Host{}), "SelectSecret",
		func(_ *mutualtrust.Host)  (string, error) {
			return "aaa-test-aaa",nil
		})
	defer patches.Reset()
	err1:=errors.New("write key fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(&mutualtrust.Host{}), "SecretWrite",
		func(_ *mutualtrust.Host , authKey string) error {
			return err1
		})
	defer patches1.Reset()
	err := d.CreateMutualTrust(ctx)

	assert.New(t).Equal(err,err1)
}

func Test_CreateMutualTrust_succ (t *testing.T){
	ctx:=context.Background()
	d := new(TrafficReplayDeploy)
	d.Basic=new(TrafficReplayBasic)
	b := d.Basic
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

	patches := gomonkey.ApplyMethod(reflect.TypeOf(&mutualtrust.Host{}), "SelectSecret",
		func(_ *mutualtrust.Host)  (string, error) {
			return "aaa-test-aaa",nil
		})
	defer patches.Reset()

	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(&mutualtrust.Host{}), "SecretWrite",
		func(_ *mutualtrust.Host , authKey string) error {
			return nil
		})
	defer patches1.Reset()
	err := d.CreateMutualTrust(ctx)

	assert.New(t).Nil(err)
}

func Test_getBinPath_with_NewRemoteSession_fail(t *testing.T){
	config :=cmd.SessionConfig{
		User: "root",
		Password: "******",
		Host:"192.168.1.1",
		Port: 22,
		Timeout: 0,
	}
	d := new(TrafficReplayDeploy)
	ctx :=context.Background()
	err1 := errors.New("connect remote server fail")
	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return nil,err1
	})
	defer patch.Reset()
	path,err := d.GetBinPath(ctx,config)

	assert.New(t).Equal(path,"")
	assert.New(t).Equal(err1,err)
}

func Test_getBinPath_with_exec_fail(t *testing.T){
	config :=cmd.SessionConfig{
		User: "root",
		Password: "******",
		Host:"192.168.1.1",
		Port: 22,
		Timeout: 0,
	}
	d := new(TrafficReplayDeploy)
	ctx :=context.Background()
	err1 := errors.New("exec command on remote server fail")
	errClose := errors.New("close connect on remote server fail ")
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	remoteSess := cmd.NewMockExecSession(mockCtl)
	remoteSess.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(err1)
	remoteSess.EXPECT().Close().Return(errClose)
	remoteSess.EXPECT().Output().Return(" ")
	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return remoteSess,nil
	})
	defer patch.Reset()
	path,err := d.GetBinPath(ctx,config)

	assert.New(t).Equal(path,"")
	assert.New(t).Equal(err1,err)
}

func Test_getBinPath_succ(t *testing.T){
	config :=cmd.SessionConfig{
		User: "root",
		Password: "******",
		Host:"192.168.1.1",
		Port: 22,
		Timeout: 0,
	}
	d := new(TrafficReplayDeploy)
	ctx :=context.Background()
	path :="/root/.tiup/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:" +
		"/root/bin:/usr/local/go//bin "
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	remoteSess := cmd.NewMockExecSession(mockCtl)
	remoteSess.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
	remoteSess.EXPECT().Close().Return(nil)
	remoteSess.EXPECT().Output().Return(path)
	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return remoteSess,nil
	})
	defer patch.Reset()
	path1,err := d.GetBinPath(ctx,config)

	assert.New(t).Equal(path1,path)
	assert.New(t).Nil(err)
}


func Test_CheckBinExist_with_NewRemoteSession_fail(t *testing.T){
	config :=cmd.SessionConfig{
		User: "root",
		Password: "******",
		Host:"192.168.1.1",
		Port: 22,
		Timeout: 0,
	}
	d := new(TrafficReplayDeploy)
	ctx :=context.Background()
	path:="/usr/sbin"
	binName:="tcpdump"
	err1 := errors.New("connect remote server fail")
	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return nil,err1
	})
	defer patch.Reset()
	res,step,err := d.CheckBinExist(ctx,config,path,binName)

	assert.New(t).Equal(res,false)
	assert.New(t).Equal(step,0)
	assert.New(t).Equal(err,err1)
}

func Test_CheckBinExist_with_exec_fail(t *testing.T){
	config :=cmd.SessionConfig{
		User: "root",
		Password: "******",
		Host:"192.168.1.1",
		Port: 22,
		Timeout: 0,
	}
	d := new(TrafficReplayDeploy)
	ctx :=context.Background()
	path:="/usr/sbin"
	binName:="tcpdump"
	err1 := errors.New("exec command on remote server fail")
	errClose := errors.New("close connect on remote server fail ")
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	remoteSess := cmd.NewMockExecSession(mockCtl)
	remoteSess.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(err1)
	remoteSess.EXPECT().Close().Return(errClose)
	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return remoteSess,nil
	})
	defer patch.Reset()

	res,step,err := d.CheckBinExist(ctx,config,path,binName)

	assert.New(t).Equal(res,false)
	assert.New(t).Equal(step,1)
	assert.New(t).Equal(err,err1)

}

func Test_CheckBinExist_succ(t *testing.T){
	config :=cmd.SessionConfig{
		User: "root",
		Password: "******",
		Host:"192.168.1.1",
		Port: 22,
		Timeout: 0,
	}
	d := new(TrafficReplayDeploy)
	ctx :=context.Background()
	path:="/usr/sbin"
	binName:="tcpdump"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	remoteSess := cmd.NewMockExecSession(mockCtl)
	remoteSess.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
	remoteSess.EXPECT().Close().Return(nil)
	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return remoteSess,nil
	})
	defer patch.Reset()
	res,step,err := d.CheckBinExist(ctx,config,path,binName)

	assert.New(t).Equal(res,true)
	assert.New(t).Equal(step,1)
	assert.New(t).Nil(err)
}


func Test_CheckProductTidbServerBinExist_with_GetBinPath_fail (t *testing.T){
	d := new(TrafficReplayDeploy)
	ctx := context.Background()
	binName := "tcpdump"
	d.Basic=new(TrafficReplayBasic)
	b := d.Basic
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


	err1:=errors.New("get bin path fail")
	patches := gomonkey.ApplyMethod(reflect.TypeOf(d), "GetBinPath",
		func(_ *TrafficReplayDeploy, ctx context.Context, config cmd.SessionConfig) (string, error)  {
			return "",err1
		})
	defer patches.Reset()

	err:= d.CheckProductTidbServerBinExist(ctx,binName)
	assert.New(t).Equal(err,err1)

}

func Test_CheckProductTidbServerBinExist_with_CheckBinExist_step0_fail (t *testing.T){
	d := new(TrafficReplayDeploy)
	ctx := context.Background()
	binName := "tcpdump"
	d.Basic=new(TrafficReplayBasic)
	b := d.Basic
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


	patches := gomonkey.ApplyMethod(reflect.TypeOf(d), "GetBinPath",
		func(_ *TrafficReplayDeploy, ctx context.Context, config cmd.SessionConfig) (string, error)  {
			return "/usr/bin",nil
		})
	defer patches.Reset()

	err1:=errors.New("check bin exist fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(d), "CheckBinExist",
		func(_ *TrafficReplayDeploy, ctx context.Context, config cmd.SessionConfig,
			path, binName string) (bool, int, error) {
			return false,0,err1
		})
	defer patches1.Reset()

	err:= d.CheckProductTidbServerBinExist(ctx,binName)
	assert.New(t).Equal(err,err1)

}

func Test_CheckProductTidbServerBinExist_with_CheckBinExist_step1_fail (t *testing.T){
	d := new(TrafficReplayDeploy)
	ctx := context.Background()
	binName := "tcpdump"
	d.Basic=new(TrafficReplayBasic)
	b := d.Basic
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


	patches := gomonkey.ApplyMethod(reflect.TypeOf(d), "GetBinPath",
		func(_ *TrafficReplayDeploy, ctx context.Context, config cmd.SessionConfig) (string, error)  {
			return "/usr/bin",nil
		})
	defer patches.Reset()

	err1:=errors.New("check bin exist fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(d), "CheckBinExist",
		func(_ *TrafficReplayDeploy, ctx context.Context, config cmd.SessionConfig,
			path, binName string) (bool, int, error) {
			return false,1,err1
		})
	defer patches1.Reset()

	err:= d.CheckProductTidbServerBinExist(ctx,binName)
	assert.New(t).Nil(err)

}

func Test_CheckProductTidbServerBinExist_with_CheckBinExist_step1_succ (t *testing.T){
	d := new(TrafficReplayDeploy)
	ctx := context.Background()
	binName := "tcpdump"
	d.Basic=new(TrafficReplayBasic)
	b := d.Basic
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

	d.TcpdumpExist = make([]bool,len(b.HostKyePairs[0]))

	patches := gomonkey.ApplyMethod(reflect.TypeOf(d), "GetBinPath",
		func(_ *TrafficReplayDeploy, ctx context.Context, config cmd.SessionConfig) (string, error)  {
			return "/usr/bin",nil
		})
	defer patches.Reset()

	err1:=errors.New("check bin exist fail")
	patches1 := gomonkey.ApplyMethod(reflect.TypeOf(d), "CheckBinExist",
		func(_ *TrafficReplayDeploy, ctx context.Context, config cmd.SessionConfig,
			path, binName string) (bool, int, error) {
			return true,1,err1
		})
	defer patches1.Reset()

	err:= d.CheckProductTidbServerBinExist(ctx,binName)
	assert.New(t).Nil(err)
	assert.New(t).True(d.TcpdumpExist[0])

}

func Test_DeployBinOnRemoteServer_with_NewRemoteSession_fail(t *testing.T){
	config :=cmd.SessionConfig{
		User: "root",
		Password: "******",
		Host:"192.168.1.1",
		Port: 22,
		Timeout: 0,
	}
	d := new(TrafficReplayDeploy)
	ctx :=context.Background()
	binName:="tcpdump"
	err1 := errors.New("connect remote server fail")
	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return nil,err1
	})
	defer patch.Reset()
	err := d.DeployBinOnRemoteServer(ctx,config,binName)

	assert.New(t).Equal(err,err1)
}

func Test_DeployBinOnRemoteServer_with_exec_fail(t *testing.T){
	config :=cmd.SessionConfig{
		User: "root",
		Password: "******",
		Host:"192.168.1.1",
		Port: 22,
		Timeout: 0,
	}
	d := new(TrafficReplayDeploy)
	ctx :=context.Background()
	binName:="tcpdump"
	err1 := errors.New("exec command on remote server fail")
	errClose := errors.New("close connect on remote server fail ")
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	remoteSess := cmd.NewMockExecSession(mockCtl)
	remoteSess.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(err1)
	remoteSess.EXPECT().Close().Return(errClose)
	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return remoteSess,nil
	})
	defer patch.Reset()

	err := d.DeployBinOnRemoteServer(ctx,config,binName)

	assert.New(t).Equal(err,err1)

}

func Test_DeployBinOnRemoteServer_succ(t *testing.T){
	config :=cmd.SessionConfig{
		User: "root",
		Password: "******",
		Host:"192.168.1.1",
		Port: 22,
		Timeout: 0,
	}
	d := new(TrafficReplayDeploy)
	ctx :=context.Background()
	binName:="tcpdump"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	remoteSess := cmd.NewMockExecSession(mockCtl)
	remoteSess.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
	remoteSess.EXPECT().Close().Return(nil)
	patch := gomonkey.ApplyFunc( cmd.NewRemoteSession,func (config cmd.SessionConfig) (cmd.ExecSession, error)  {
		return remoteSess,nil
	})
	defer patch.Reset()
	err := d.DeployBinOnRemoteServer(ctx,config,binName)

	assert.New(t).Nil(err)
}
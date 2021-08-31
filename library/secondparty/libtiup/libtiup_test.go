package libtiup

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/pingcap-inc/tiem/library/framework"
)

func init() {
	// make sure log would not cause nil pointer problem
	log = framework.GetLogger()
	TiupMgrInit()
}

const (
	cmdDeployNonJsonStr = "nonJsonStr"
)

func loopUntilDone() {
	for {
		time.Sleep(time.Second)
		glMgrStatusMapSync()
		taskStatusMapValue := glMgrTaskStatusMap[0]
		fmt.Println("taskId 0 status: ", taskStatusMapValue.stat.Status)
		if taskStatusMapValue.stat.Status == TaskStatusError || taskStatusMapValue.stat.Status == TaskStatusFinished {
			break
		}
	}
}

func TestMgrHandleCmdDeployReqWithNonJsonStr(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		mgrHandleCmdDeployReq(cmdDeployNonJsonStr)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMgrHandleCmdDeployReqWithNonJsonStr")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

// todo: make it standalone(depends on tiup binary), comment out for now
/**
func TestMgrHandleCmdDeployReq(t *testing.T) {
	cmdDeployReq := CmdDeployReq{
		TaskID: 0,
		InstanceName: "test-cluster",
		Version: "v5.1.1",
		ConfigStrYaml: "global:\n  user: \"root\"\n  ssh_port: 22\n  deploy_dir: \"/root/tidb-deploy\"\n  data_dir: \"/root/tidb-data\"\n  arch: \"amd64\"\n\nmonitored:\n  node_exporter_port: 9100\n  blackbox_exporter_port: 9115\n\npd_servers:\n  - host: 172.16.5.128\n\ntidb_servers:\n  - host: 172.16.5.128\n    port: 4000\n    status_port: 10080\n    deploy_dir: \"/root/tidb-deploy/tidb-4000\"\n    log_dir: \"/root/tidb-deploy/tidb-4000/log\"\n\ntikv_servers:\n  - host: 172.16.5.128\n    port: 20160\n    status_port: 20180\n    deploy_dir: \"/root/data1/tidb-deploy/tikv-20160\"\n    data_dir: \"/root/tidb-data/tikv-20160\"\n    log_dir: \"/root/tidb-deploy/tikv-20160/log\"\n\nmonitoring_servers:\n  - host: 172.16.5.128\n\ngrafana_servers:\n  - host: 172.16.5.128\n\nalertmanager_servers:\n  - host: 172.16.5.128",
		TiupPath: "/root/.tiup/bin/tiup",
	}
	mgrHandleCmdDeployReq(string(jsonMustMarshal(cmdDeployReq)))
	loopUntilDone()
	// then check the cluster by command (tiup cluster list AND tiup cluster display test-cluster) for now
}
*/

// todo: make it standalone(depends on tiup binary), comment out for now
/**
func TestMgrHandleCmdStartReq(t *testing.T) {
	cmdStartReq := CmdStartReq{
		TaskID: 0,
		InstanceName: "test-cluster",
		TiupPath: "/root/.tiup/bin/tiup",
	}
	mgrHandleCmdStartReq(string(jsonMustMarshal(cmdStartReq)))
	loopUntilDone()
	// then check the cluster by command (tiup cluster list AND tiup cluster display test-cluster) for now
}
*/

// todo: make it standalone(depends on tiup binary), comment out for now
/**
func TestMgrHandleCmdDestroyReq(t *testing.T) {
	cmdDestroyReq := CmdDestroyReq{
		TaskID: 0,
		InstanceName: "test-cluster",
		TiupPath: "/root/.tiup/bin/tiup",
	}
	mgrHandleCmdDestroyReq(string(jsonMustMarshal(cmdDestroyReq)))
	loopUntilDone()
	// then check the cluster by command (tiup cluster list AND tiup cluster display test-cluster) for now
}
*/

// todo: make it standalone(depends on tiup binary), comment out for now
/**
func TestMgrHandleClusterDisplayReq(t *testing.T) {
	cmdClusterDisplayReq := CmdClusterDisplayReq{
		InstanceName: "test-cluster",
		TiupPath: "/root/.tiup/bin/tiup",
	}
	resp := mgrHandleClusterDisplayReq(string(jsonMustMarshal(cmdClusterDisplayReq)))
	fmt.Println("tiup cluster display resp: ", resp)
	// then check the cluster by command (tiup cluster list AND tiup cluster display test-cluster) for now
}
*/

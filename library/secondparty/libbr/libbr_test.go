package libbr

import (
	"fmt"
	"github.com/pingcap/tiem/library/thirdparty/logger"
	"testing"
)

func init() {
	fmt.Println("init before test")
	log = logger.GetLogger("config_key_test_log")
	//BrMgrInit()
	// todo: all binaries has been put under bin folder of project root path in another branch, need to change later
	MicroInit("../../../bin/micro-cluster/brmgr/brmgr", "/tmp/log/br/")

}

func TestBrMgrInit(t *testing.T) {
	BrMgrInit()
	mgrTaskStatusChCap := cap(glMgrTaskStatusCh)
	if mgrTaskStatusChCap != 1024 {
		t.Errorf("glMgrTaskStatusCh cap was incorrect, got: %d, want: %d.", mgrTaskStatusChCap, 1024)
	}
}

func TestBackUpAndInfo(t *testing.T) {
}

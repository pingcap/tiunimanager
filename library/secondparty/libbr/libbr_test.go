package libbr

import "testing"

func TestBrMgrInit(t *testing.T) {
	BrMgrInit()
	mgrTaskStatusChCap := cap(glMgrTaskStatusCh)
	if mgrTaskStatusChCap != 1024 {
		t.Errorf("glMgrTaskStatusCh cap was incorrect, got: %d, want: %d.", mgrTaskStatusChCap, 1024)
	}
}

func TestBackUpAndInfo(t *testing.T) {

}

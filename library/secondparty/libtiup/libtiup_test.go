package libtiup

import (
	"testing"
)

func TestMicroSrvTiupClusterDisplay(t *testing.T) {
	err := MicroSrvTiupClusterDisplay("testClusterName")
	expected := "respond error on purpose: testClusterName"
	if err == nil || err.Error() != expected {
		t.Errorf("error msg was incorrect, got: %s, want: %s.", expected, err.Error())
	}
}
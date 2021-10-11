package secondparty

import (
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"testing"
)

func init() {
	logger = framework.LogForkFile(common.LogFileLibTiup)
}

func Test_assert_false(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	assert(false)
}

func Test_assert_true(t *testing.T) {
	assert(true)
}

func Test_myPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	myPanic("panic info")
}

func Test_newTmpFileWithContent(t *testing.T) {
	content := []byte("temp info")
	_, err := newTmpFileWithContent(content)
	if err != nil {
		t.Errorf(err.Error())
	}
}
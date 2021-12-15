/*******************************************************************************
 * @File: common_test
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/14
*******************************************************************************/

package handler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContain(t *testing.T) {
	ports := []int{4000, 4001, 4002}
	got := Contain(ports, 4000)
	assert.Equal(t, got, true)

	got = Contain(ports, 4003)
	assert.Equal(t, got, false)

	hosts := []string{"127.0.0.1", "127.0.0.2"}
	got = Contain(hosts, "127.0.0.1")
	assert.Equal(t, got, true)

	got = Contain(hosts, "127.0.0.3")
	assert.Equal(t, got, false)
}

func TestScaleOutPreCheck(t *testing.T) {

}

func TestWaitWorkflow(t *testing.T) {
	//TODO: test WaitWorkflow
}

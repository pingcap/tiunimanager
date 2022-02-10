/*******************************************************************************
 * @File: main_test
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/17
*******************************************************************************/

package identification

import (
	"github.com/pingcap-inc/tiem/models"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	models.MockDB()

	os.Exit(m.Run())
}

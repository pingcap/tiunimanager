package instanceapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-cluster/client"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
	"net/http"
	"strconv"
)

// Query 查询实例接口
// @Summary 查询实例接口
// @Description 查询实例
// @Tags instance
// @Accept application/json
// @Produce application/json
// @Param Token header string true "登录token"
// @Param query body InstanceQuery true "查询参数"
// @Success 200 {object} controller.ResultWithPage{data=[]InstanceInfo}
// @Router /instance/query [post]
func Query(c *gin.Context) {
	var req InstanceQuery

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	resp,err := client.ClusterClient.QueryCluster(c, &cluster.QueryClusterRequest{
		Page: &cluster.ClusterPage{
			Page: int32(req.Page),
			PageSize: int32(req.PageSize),
		},
	})

	if err != nil {
		// 处理异常
	}

	instanceInfos := make([]InstanceInfo, len(resp.Clusters), cap(resp.Clusters))
	for index, dto := range resp.Clusters {
		instanceInfos[index] = CopyInstanceInfoFromDTO(dto)
	}
	c.JSON(http.StatusOK, controller.SuccessWithPage(instanceInfos, controller.Page{Page: req.Page, PageSize: req.PageSize, Total: len(instanceInfos)}))
}

func CopyInstanceInfoFromDTO (dto *cluster.ClusterInfoDTO) (instanceInfo InstanceInfo) {
	instanceInfo.InstanceId = strconv.Itoa(int(dto.Id))
	instanceInfo.InstanceName = dto.Name
	instanceInfo.InstanceVersion = dto.Version
	instanceInfo.InstanceStatus = int(dto.Status)
	return
}

// Create 创建实例接口
// @Summary 创建实例接口
// @Description 创建实例
// @Tags instance
// @Accept application/json
// @Produce application/json
// @Param Token header string true "登录token"
// @Param createInfo body InstanceCreate true "创建参数"
// @Success 200 {object} controller.CommonResult{data=InstanceInfo}
// @Router /instance/create [post]
func Create(c *gin.Context) {
	var req InstanceCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	accountName := c.GetString("accountName")
	tenantId := c.GetInt("tenantId")

	fmt.Println("accountName", accountName)
	fmt.Println("tenantId", tenantId)

	resp,err := client.ClusterClient.CreateCluster(c, &cluster.CreateClusterRequest{
		Name:       req.InstanceName,
		Version:    req.InstanceVersion,
		DbPassword: req.DBPassword,
		PdCount:    int32(req.PDCount),
		TidbCount:  int32(req.TiDBCount),
		TikvCount: int32(req.TiKVCount),
		Operator: &cluster.ClusterOperatorDTO{
			TenantId: int32(tenantId),
			AccountName: accountName,
		},
	})

	if err != nil {
		// 处理异常
	}

	c.JSON(http.StatusOK, controller.Success(CopyInstanceInfoFromDTO(resp.Cluster)))
}
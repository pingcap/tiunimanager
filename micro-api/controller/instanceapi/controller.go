package instanceapi

// QueryParams 查询集群参数列表
// @Summary 查询集群参数列表
// @Description 查询集群参数列表
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param query body ParamQueryReq false "查询参数"
// @Success 200 {object} controller.ResultWithPage{data=[]ParamItem}
// @Router /param/query [get]
func QueryParams() {

}

// SubmitParams
// ParamUpdateReq
// ParamUpdateRsp
// SubmitParams 提交参数
// @Summary 提交参数
// @Description 提交参数
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param request body ParamUpdateReq true "要提交的参数信息"
// @Success 200 {object} controller.CommonResult{data=ParamUpdateRsp}
// @Router /param/submit [post]
func SubmitParams() {

}

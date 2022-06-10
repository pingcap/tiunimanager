# TiUniManager workflow design doc

- Author(s): [cchenkey](http://github.com/cchenkey)

## table of Contents

- [TiUniManagerworkflow design doc](#TiUniManager-workflow-design-doc)
  - [table of Contents](#table-of-Contents)
  - [workflow API](#workflow-API)
  - [design](#design)
    - [interface](#interface)
    - [workflow implement](#workflow-implement)
    - [errors](#errors)

## workflow API
micro-api/route/route.go，"/workflow" group define workflow API,
``` go
		flowworks := apiV1.Group("/workflow")
		{
			flowworks.GET("/", metrics.HandleMetrics(constants.MetricsWorkFlowQuery), flowtaskApi.Query)
			flowworks.GET("/:workFlowId", metrics.HandleMetrics(constants.MetricsWorkFlowDetail), flowtaskApi.Detail)
			flowworks.POST("/start", metrics.HandleMetrics(constants.MetricsWorkFlowStart), flowtaskApi.Start)
			flowworks.POST("/stop", metrics.HandleMetrics(constants.MetricsWorkFlowStop), flowtaskApi.Stop)
		}
```
The `HandleFunc` of each route can be tracked as the entry of each API，
- `flowtaskApi.Query` query workflow list；
- `flowtaskApi.Detail` query workflow detail info；
- `flowtaskApi.Start` start workflow；
- `flowtaskApi.Stop` stop workflow；

## design

workflowManager is the manager of workflow（workflow2），each file function：

- `service.go` definition interface of workflow；
- `workflow.go` implement of workflow interface；
- `meta.go` workflow meta data；
- `definition.go` workflow state machine definition；
- `common.go` constant definition of workflow；

### interface
workflow interface：
``` go
type WorkFlowService interface {
	// RegisterWorkFlow
	// @Description: register workflow define
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowName
	// @Parameter flowDefine
	RegisterWorkFlow(ctx context.Context, flowName string, flowDefine *WorkFlowDefine)

	// GetWorkFlowDefine
	// @Description: get workflow define by flowName
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowName
	// @Return *WorkFlowDefine
	// @Return error
	GetWorkFlowDefine(ctx context.Context, flowName string) (*WorkFlowDefine, error)

	// CreateWorkFlow
	// @Description: create new workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter bizId
	// @Parameter bizType
	// @Parameter flowName
	// @Return flowId
	// @Return error
	CreateWorkFlow(ctx context.Context, bizId string, bizType string, flowName string) (string, error)

	// ListWorkFlows
	// @Description: list workflows by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.QueryWorkFlowsResp
	// @Return total
	// @Return error
	ListWorkFlows(ctx context.Context, request message.QueryWorkFlowsReq) (message.QueryWorkFlowsResp, structs.Page, error)

	// DetailWorkFlow
	// @Description: create new workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.QueryWorkFlowDetailResp
	// @Return error
	DetailWorkFlow(ctx context.Context, request message.QueryWorkFlowDetailReq) (message.QueryWorkFlowDetailResp, error)

	// InitContext
	// @Description: init flow context for workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Parameter key
	// @Parameter value
	// @Return error
	InitContext(ctx context.Context, flowId string, key string, value interface{}) error

	// Start
	// @Description: async start workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Return error
	Start(ctx context.Context, flowId string) error

	// Stop
	// @Description: stop workflow in current node
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Return error
	Stop(ctx context.Context, flowId string) error

	// Cancel
	// @Description: cancel workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Parameter reason
	// @Return error
	Cancel(ctx context.Context, flowId string, reason string) error
}
```

### workflow implement

- in metadb of tiunimanager, workflow contains one workflow and some workflow nodes belongs to it
- workflow meta define info of workflow, definition and flow context
- user can call workflow interface will update workflow status
- after workflow manager init，it will start a watchLoop goroutine，it will handle unfinished workflow and push workflow until end, fail or pause

### errors

workflow API errors define in file common/errors/errorcode.go
``` go
	// workflow
	TIUNIMANAGER_WORKFLOW_CREATE_FAILED:         {"workflow create failed", 500},
	TIUNIMANAGER_WORKFLOW_QUERY_FAILED:          {"workflow query failed", 500},
	TIUNIMANAGER_WORKFLOW_DETAIL_FAILED:         {"workflow detail failed", 500},
	TIUNIMANAGER_WORKFLOW_START_FAILED:          {"workflow start failed", 500},
	TIUNIMANAGER_WORKFLOW_DEFINE_NOT_FOUND:      {"workflow define not found", 404},
	TIUNIMANAGER_WORKFLOW_NODE_POLLING_TIME_OUT: {"workflow node polling time out", 500},
```

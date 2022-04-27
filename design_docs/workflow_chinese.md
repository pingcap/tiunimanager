# TiEM workflow设计文档

- Author(s): [cchenkey](http://github.com/cchenkey)

## 目录

- [TiEM workflow设计文档](#tiem-workflow设计文档)
  - [目录](#目录)
  - [workflow模块API](#workflow模块API)
  - [设计](#设计)
    - [接口](#接口)
    - [workflow实现](#workflow实现)
    - [异常错误](#异常错误)

## workflow模块API
在micro-api/route/route.go中，"/workflow" group定义了导入导出模块提供的API,
``` go
		flowworks := apiV1.Group("/workflow")
		{
			flowworks.GET("/", metrics.HandleMetrics(constants.MetricsWorkFlowQuery), flowtaskApi.Query)
			flowworks.GET("/:workFlowId", metrics.HandleMetrics(constants.MetricsWorkFlowDetail), flowtaskApi.Detail)
			flowworks.POST("/start", metrics.HandleMetrics(constants.MetricsWorkFlowStart), flowtaskApi.Start)
			flowworks.POST("/stop", metrics.HandleMetrics(constants.MetricsWorkFlowStop), flowtaskApi.Stop)
		}
```
可以跟踪每条route后的HandleFunc进入每个API的入口，其中，
- `flowtaskApi.Query` 查询workflow列表；
- `flowtaskApi.Detail` 查询workflow详情；
- `flowtaskApi.Start` 启动workflow；
- `flowtaskApi.Stop` 停止workflow；

## 设计

workflowManager是workflow具体流程的入口（workflow2），文件功能如下：

- `service.go` workflow的接口定义；
- `workflow.go` workflow定义和接口实现；
- `meta.go` workflow元数据的持久化、加载；
- `definition.go` workflow状态机定义；
- `common.go` workflow相关常量定义；

### 接口
workflow的接口如下：
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

### workflow实现
workflow具体实现：

- workflow在数据库里结构由workflow关联多个workflowNode，组成一个完整的workflow
- meta里定义了workflow的基础信息，状态机运转信息，业务上下文信息
- workflow的接口主要是修改workflow的状态
- workflow manager初始化后，会启动一个watchLoop协程，不断处理没有结束的任务，直到状态机到达结束和失败，或者中断

### 异常错误

workflow的API报错定义在common/errors/errorcode.go
``` go
	// workflow
	TIEM_WORKFLOW_CREATE_FAILED:         {"workflow create failed", 500},
	TIEM_WORKFLOW_QUERY_FAILED:          {"workflow query failed", 500},
	TIEM_WORKFLOW_DETAIL_FAILED:         {"workflow detail failed", 500},
	TIEM_WORKFLOW_START_FAILED:          {"workflow start failed", 500},
	TIEM_WORKFLOW_DEFINE_NOT_FOUND:      {"workflow define not found", 404},
	TIEM_WORKFLOW_NODE_POLLING_TIME_OUT: {"workflow node polling time out", 500},
```

package service

import (
	"context"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
	"github.com/pingcap/ticp/micro-cluster/service/cluster/domain"
)

var TiCPClusterServiceName = "go.micro.ticp.cluster"

var SuccessResponseStatus = &cluster.ResponseStatusDTO {Code:0}
var BizErrorResponseStatus = &cluster.ResponseStatusDTO {Code:1}

type ClusterServiceHandler struct {}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *cluster.ClusterCreateReqDTO, resp *cluster.ClusterCreateRespDTO) error {
	clusterAggregation, err := domain.CreateCluster(req.GetOperator(), req.GetCluster(), req.GetDemands())

	if err != nil {
		// todo
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		resp.BaseInfo = clusterAggregation.ExtractBaseInfoDTO()
		resp.ClusterStatus = clusterAggregation.ExtractStatusDTO()
		return nil
	}

}

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *cluster.ClusterQueryReqDTO, resp *cluster.ClusterQueryRespDTO) error {
	return nil
}

func (c ClusterServiceHandler) DeleteCluster(ctx context.Context, req *cluster.ClusterDeleteReqDTO, resp *cluster.ClusterDeleteRespDTO) error {
	clusterAggregation, err := domain.DeleteCluster(req.GetOperator(), req.GetClusterId())
	if err != nil {
		// todo
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.ClusterId = clusterAggregation.Cluster.Id
		resp.ClusterStatus = clusterAggregation.ExtractStatusDTO()
		return nil
	}
}

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *cluster.ClusterDetailReqDTO, resp *cluster.ClusterDetailRespDTO) error {
	cluster, err := domain.GetClusterDetail(req.ClusterId, req.Operator)

	if err != nil {
		// todo
		return nil
	} else {
		resp.RespStatus = SuccessResponseStatus
		resp.DisplayInfo = cluster.ExtractDisplayDTO()
		resp.Components = cluster.ExtractComponentDTOs()
		resp.MaintenanceInfo = cluster.ExtractMaintenanceDTO()

		return nil
	}
}

func (c ClusterServiceHandler) ExportData(ctx context.Context, request *cluster.DataExportRequest, resp *cluster.DataExportResponse) error {
	//todo: implement
	return nils
}

func (c ClusterServiceHandler) ImportData(ctx context.Context, request *cluster.DataImportRequest, resp *cluster.DataImportResponse) error {
	//todo: implement
	return nil
}

func (c ClusterServiceHandler) DescribeDataTransport(ctx context.Context, request *cluster.DataTransportQueryRequest, resp *cluster.DataTransportQueryResponse) error {
	//todo: implement
	return nil
}
/**
 * @Author: guobob
 * @Description:
 * @File:  trafficreplay.go
 * @Version: 1.0.0
 * @Date: 2021/12/17 14:40
 */

package cluster

import "time"

type TrafficReplayTask struct {
 ID                  string    `json:"id" form:"id" example:"CLUSTER_ID_IN_TIEM__22"`
 ProductClusterID    string    `json:"product_cluster_id" form:"product_cluster_id" example:"CLUSTER_ID_IN_TIEM__22"`
 SimulationClusterID string    `json:"simulation_cluster_id" form:"simulation_cluster_id" example:"CLUSTER_ID_IN_TIEM__22"`
 RunTime             int64     `json:"run_time" form:"run_time" example:"415241823337054209"`
 DeployPath          string    `json:"deploy_path" form:"deploy_path" example:"/home/deploy"`
 TcpdumpTempPath     string    `json:"tcpdump_temp_path" form:"tcpdump_temp_path" example:"/home/temp"`
 PcapTempPath        string    `json:"pcap_temp_path" form:"pcap_emp_path" example:"/home/temp"`
 PcapFileStorePath   string    `json:"pcap_file_store_path" form:"pacp_file_store_path" example:"/home/pcap"`
 ResultFileStorePath string    `json:"result_file_store_path" form:"result_file_store_path" example:"/home/result"`
 CreateTime          time.Time `json:"createTime" form:"createTime"`
 UpdateTime          time.Time `json:"updateTime" form:"updateTime"`
}

type TrafficReplayResp struct {
 TaskID string `json:"task_id"`
}


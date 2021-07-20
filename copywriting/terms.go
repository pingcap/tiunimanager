package copywriting

var DefaultLanguage = "cn"

func DisplayByDefault(key string) string {
	return CopyWritingContainer[DefaultLanguage][key]
}

func Display(key string, language string) string {
	return CopyWritingContainer[language][key]
}

// CopyWritingContainer map[language]map[copyWritingKey]copyWritingValue
var CopyWritingContainer = map[string]map[string]string{
	"cn":{
		CWClusterStatusUnlined:	"未启动",
		CWClusterStatusOnline:	"运行中",
		CWClusterStatusOffline:	"已停止",
		CWClusterStatusDeleted:	"已删除",
		CWFlowCreateCluster:	"创建中",
		CWFlowDeleteCluster:	"删除中",
	},
}

var CWClusterStatusUnlined = "CW_ClusterStatusUnlined"
var CWClusterStatusOnline = "CW_ClusterStatusOnline"
var CWClusterStatusOffline = "CW_ClusterStatusOffline"
var CWClusterStatusDeleted = "CW_ClusterStatusDeleted"

var CWFlowCreateCluster = "CW_FlowCreateCluster"
var CWFlowDeleteCluster = "CW_FlowDeleteCluster"
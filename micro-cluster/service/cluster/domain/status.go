package domain

type ClusterStatus int

const (
	ClusterStatusUnlined ClusterStatus = 0
	ClusterStatusOnline  ClusterStatus = 1
	ClusterStatusOffline ClusterStatus = 2
	ClusterStatusDeleted ClusterStatus = 3
)

var allClusterStatus = []ClusterStatus{
	ClusterStatusUnlined,
	ClusterStatusOnline,
	ClusterStatusOffline,
	ClusterStatusDeleted,
}

func ClusterStatusFromValue(v int) ClusterStatus {
	for _, s := range allClusterStatus {
		if int(s) == v {
			return s
		}
	}
	return -1
}

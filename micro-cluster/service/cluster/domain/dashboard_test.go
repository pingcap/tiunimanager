package domain

import (
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDashboardUrlFromCluser(t *testing.T) {
	clusterAggregation := &ClusterAggregation{
		CurrentTiUPConfigRecord: &TiUPConfigRecord{
			ConfigModel: &spec.Specification{
				PDServers: []*spec.PDSpec{
					{
						Host: "127.0.0.1",
						ClientPort: 2379,
					},
				},
			},
		},
	}
	url := getDashboardUrlFromCluser(clusterAggregation)
	assert.Equal(t, "http://127.0.0.1:2379/dashboard/", url)
}

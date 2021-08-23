package framework

type TiemService string

const (
	MetaDBService TiemService = "go.micro.tiem.db"
	ClusterService = "go.micro.tiem.cluster"
	ApiService = "go.micro.tiem.api"
)

func (p TiemService) ToString() string {
	return string(p)
}
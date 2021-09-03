package domain

func setupMockAdapter() {
	TaskRepo = MockTaskRepo{}
	ClusterRepo = MockClusterRepo{}
	InstanceRepo = MockInstanceRepo{}
}

type MockTaskRepo struct {}

var id uint = 0

func getId() uint {
	id = id + 1
	return id
}
func (m MockTaskRepo) AddFlowWork(flowWork *FlowWorkEntity) error {
	flowWork.Id = getId()
	return nil
}

func (m MockTaskRepo) AddFlowTask(task *TaskEntity, flowId uint) error {
	task.Id = getId()
	return nil
}

func (m MockTaskRepo) AddCronTask(cronTask *CronTaskEntity) error {
	panic("implement me")
}

func (m MockTaskRepo) Persist(flowWork *FlowWorkAggregation) error {
	return nil
}

func (m MockTaskRepo) LoadFlowWork(id uint) (*FlowWorkEntity, error) {
	panic("implement me")
}

func (m MockTaskRepo) Load(id uint) (flowWork *FlowWorkAggregation, err error) {
	panic("implement me")
}

func (m MockTaskRepo) QueryCronTask(bizId string, cronTaskType int) (cronTask *CronTaskEntity, err error) {
	panic("implement me")
}

func (m MockTaskRepo) PersistCronTask(cronTask *CronTaskEntity) (err error) {
	panic("implement me")
}

type MockClusterRepo struct {}

func (m MockClusterRepo) AddCluster(cluster *Cluster) error {
	panic("implement me")
}

func (m MockClusterRepo) Persist(aggregation *ClusterAggregation) error {
	return nil
}

func (m MockClusterRepo) Load(id string) (cluster *ClusterAggregation, err error) {
	panic("implement me")
}

func (m MockClusterRepo) Query(clusterId, clusterName, clusterType, clusterStatus, clusterTag string, page, pageSize int) ([]*ClusterAggregation, int, error) {
	panic("implement me")
}

type MockInstanceRepo struct {}

func (m MockInstanceRepo) QueryParameterJson(clusterId string) (string, error) {
	panic("implement me")
}


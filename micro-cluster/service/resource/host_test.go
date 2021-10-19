package resource

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"testing"

	"github.com/asim/go-micro/v3/client"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func genHostInfo(hostName string) *clusterpb.HostInfo {
	host := clusterpb.HostInfo{
		Ip:       "192.168.56.11",
		HostName: hostName,
		Os:       "Centos",
		Kernel:   "3.10",
		Region:   "TEST_REGION",
		Az:       "TEST_AZ",
		Rack:     "TEST_RACK",
		Status:   0,
		Nic:      "10GE",
		Purpose:  "Compute",
	}
	host.Disks = append(host.Disks, &clusterpb.Disk{
		Name:     "sda",
		Path:     "/",
		Status:   0,
		Capacity: 512,
	})
	host.Disks = append(host.Disks, &clusterpb.Disk{
		Name:     "sdb",
		Path:     "/mnt/sdb",
		Status:   0,
		Capacity: 1024,
	})
	return &host
}

func genHostRspFromDB(hostId string) *dbpb.DBHostInfoDTO {
	host := dbpb.DBHostInfoDTO{
		HostId:   hostId,
		HostName: "Test_DB2",
		Ip:       "192.168.56.11",
		Os:       "Centos",
		Kernel:   "3.10",
		Region:   "TEST_REGION",
		Az:       "TEST_AZ",
		Rack:     "TEST_RACK",
		Status:   0,
		Nic:      "10GE",
		Purpose:  "Compute",
	}
	host.Disks = append(host.Disks, &dbpb.DBDiskDTO{
		Name:     "sda",
		Path:     "/",
		Status:   0,
		Capacity: 512,
	})
	host.Disks = append(host.Disks, &dbpb.DBDiskDTO{
		Name:     "sdb",
		Path:     "/mnt/sdb",
		Status:   0,
		Capacity: 1024,
	})
	return &host
}

func Test_ImportHost_Succeed(t *testing.T) {
	fake_str := "import succeed"
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHost(func(ctx context.Context, in *dbpb.DBAddHostRequest, opts ...client.CallOption) (*dbpb.DBAddHostResponse, error) {
		if in.Host.HostName == "TEST_HOST1" {
			rsp := new(dbpb.DBAddHostResponse)
			rsp.Rs = new(dbpb.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			rsp.HostId = fake_hostId
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(clusterpb.ImportHostRequest)
	in.Host = genHostInfo("TEST_HOST1")
	out := new(clusterpb.ImportHostResponse)

	if err := resourceManager.ImportHost(context.TODO(), in, out); err != nil {
		t.Errorf("import host %s failed, err: %v\n", in.Host.HostName, err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.HostId != fake_hostId {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, hostId: %s\n", out.Rs.Code, out.Rs.Message, out.HostId)
	}
}

func Test_ImportHost_WithErr(t *testing.T) {
	fake_str := "host already exists"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHost(func(ctx context.Context, in *dbpb.DBAddHostRequest, opts ...client.CallOption) (*dbpb.DBAddHostResponse, error) {
		return nil, status.Errorf(codes.AlreadyExists, fake_str)
	})

	in := new(clusterpb.ImportHostRequest)
	in.Host = genHostInfo("TEST_HOST1")
	out := new(clusterpb.ImportHostResponse)

	if err := resourceManager.ImportHost(context.TODO(), in, out); err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() != codes.AlreadyExists || st.Message() != fake_str {
				t.Errorf("Error has wrong type: code: %d, msg: %s", st.Code(), st.Message())
			}
		} else {
			t.Errorf("Error not status, err: %v\n", err)
		}
	} else {
		t.Errorf("Should Have a Error")
	}
}

func Test_ImportHost_WithErrCode(t *testing.T) {
	fake_str := "Host Ip is not Invalid"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHost(func(ctx context.Context, in *dbpb.DBAddHostRequest, opts ...client.CallOption) (*dbpb.DBAddHostResponse, error) {
		rsp := new(dbpb.DBAddHostResponse)
		rsp.Rs = new(dbpb.DBHostResponseStatus)
		rsp.Rs.Code = int32(codes.InvalidArgument)
		rsp.Rs.Message = fake_str
		return rsp, nil
	})

	in := new(clusterpb.ImportHostRequest)
	in.Host = genHostInfo("TEST_HOST1")
	out := new(clusterpb.ImportHostResponse)

	if err := resourceManager.ImportHost(context.TODO(), in, out); err != nil {
		t.Errorf("Should not Have a Error")
	} else {
		if out.Rs.Code != int32(codes.InvalidArgument) || out.Rs.Message != fake_str {
			t.Errorf("Error has wrong type: code: %d, msg: %s", out.Rs.Code, out.Rs.Message)
		}
	}
}

func Test_ImportHostsInBatch_Succeed(t *testing.T) {
	fake_str := "import succeed"
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "yyyy-yyyy-xxxx-xxxx"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHostsInBatch(func(ctx context.Context, in *dbpb.DBAddHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBAddHostsInBatchResponse, error) {
		if in.Hosts[0].HostName == "TEST_HOST1" && in.Hosts[1].HostName == "TEST_HOST2" {
			rsp := new(dbpb.DBAddHostsInBatchResponse)
			rsp.Rs = new(dbpb.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			rsp.HostIds = make([]string, 2)
			rsp.HostIds[0] = fake_hostId1
			rsp.HostIds[1] = fake_hostId2
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(clusterpb.ImportHostsInBatchRequest)
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST1"))
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST2"))
	out := new(clusterpb.ImportHostsInBatchResponse)

	if err := resourceManager.ImportHostsInBatch(context.TODO(), in, out); err != nil {
		t.Errorf("import host in batch failed, err: %v\n", err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.HostIds[0] != fake_hostId1 || out.HostIds[1] != fake_hostId2 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, hostId0: %s, hostId1: %s\n", out.Rs.Code, out.Rs.Message, out.HostIds[0], out.HostIds[1])
	}
}

func Test_ImportHostsInBatch_WithErr(t *testing.T) {
	fake_str := "host already exists"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHostsInBatch(func(ctx context.Context, in *dbpb.DBAddHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBAddHostsInBatchResponse, error) {
		return nil, status.Errorf(codes.AlreadyExists, fake_str)
	})

	in := new(clusterpb.ImportHostsInBatchRequest)
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST1"))
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST2"))
	out := new(clusterpb.ImportHostsInBatchResponse)

	if err := resourceManager.ImportHostsInBatch(context.TODO(), in, out); err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() != codes.AlreadyExists || st.Message() != fake_str {
				t.Errorf("Error has wrong type: code: %d, msg: %s", st.Code(), st.Message())
			}
		} else {
			t.Errorf("Error not status, err: %v\n", err)
		}
	} else {
		t.Errorf("Should Have a Error")
	}
}

func Test_ImportHostsInBatch_WithErrCode(t *testing.T) {
	fake_str := "Host Ip is not Invalid"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHostsInBatch(func(ctx context.Context, in *dbpb.DBAddHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBAddHostsInBatchResponse, error) {
		rsp := new(dbpb.DBAddHostsInBatchResponse)
		rsp.Rs = new(dbpb.DBHostResponseStatus)
		rsp.Rs.Code = int32(codes.InvalidArgument)
		rsp.Rs.Message = fake_str
		return rsp, nil
	})

	in := new(clusterpb.ImportHostsInBatchRequest)
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST1"))
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST2"))
	out := new(clusterpb.ImportHostsInBatchResponse)

	if err := resourceManager.ImportHostsInBatch(context.TODO(), in, out); err != nil {
		t.Errorf("Should not Have a Error")
	} else {
		if out.Rs.Code != int32(codes.InvalidArgument) || out.Rs.Message != fake_str {
			t.Errorf("Error has wrong type: code: %d, msg: %s", out.Rs.Code, out.Rs.Message)
		}
	}
}

func Test_RemoveHost_Succeed(t *testing.T) {
	fake_str := "remove succeed"
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHost(func(ctx context.Context, in *dbpb.DBRemoveHostRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostResponse, error) {
		if in.HostId == fake_hostId {
			rsp := new(dbpb.DBRemoveHostResponse)
			rsp.Rs = new(dbpb.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(clusterpb.RemoveHostRequest)
	in.HostId = fake_hostId
	out := new(clusterpb.RemoveHostResponse)

	if err := resourceManager.RemoveHost(context.TODO(), in, out); err != nil {
		t.Errorf("remove host %s failed, err: %v\n", in.HostId, err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str {
		t.Errorf("Rsp not Expected, code: %d, msg: %s\n", out.Rs.Code, out.Rs.Message)
	}
}

func Test_RemoveHost_WithErr(t *testing.T) {
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	fake_str := "host not exists"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHost(func(ctx context.Context, in *dbpb.DBRemoveHostRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostResponse, error) {
		return nil, status.Errorf(codes.NotFound, fake_str)
	})

	in := new(clusterpb.RemoveHostRequest)
	in.HostId = fake_hostId
	out := new(clusterpb.RemoveHostResponse)

	if err := resourceManager.RemoveHost(context.TODO(), in, out); err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() != codes.NotFound || st.Message() != fake_str {
				t.Errorf("Error has wrong type: code: %d, msg: %s", st.Code(), st.Message())
			}
		} else {
			t.Errorf("Error not status, err: %v\n", err)
		}
	} else {
		t.Errorf("Should Have a Error")
	}
}

func Test_RemovetHost_WithErrCode(t *testing.T) {
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	fake_str := "Host Id is not Invalid"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHost(func(ctx context.Context, in *dbpb.DBRemoveHostRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostResponse, error) {
		rsp := new(dbpb.DBRemoveHostResponse)
		rsp.Rs = new(dbpb.DBHostResponseStatus)
		rsp.Rs.Code = int32(codes.InvalidArgument)
		rsp.Rs.Message = fake_str
		return rsp, nil
	})

	in := new(clusterpb.RemoveHostRequest)
	in.HostId = fake_hostId
	out := new(clusterpb.RemoveHostResponse)

	if err := resourceManager.RemoveHost(context.TODO(), in, out); err != nil {
		t.Errorf("Should not Have a Error")
	} else {
		if out.Rs.Code != int32(codes.InvalidArgument) || out.Rs.Message != fake_str {
			t.Errorf("Error has wrong type: code: %d, msg: %s", out.Rs.Code, out.Rs.Message)
		}
	}
}
func Test_RemoveHostsInBatch_Succeed(t *testing.T) {
	fake_str := "remove in batch succeed"
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "yyyy-yyyy-xxxx-xxxx"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHostsInBatch(func(ctx context.Context, in *dbpb.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostsInBatchResponse, error) {
		if in.HostIds[0] == fake_hostId1 && in.HostIds[1] == fake_hostId2 {
			rsp := new(dbpb.DBRemoveHostsInBatchResponse)
			rsp.Rs = new(dbpb.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(clusterpb.RemoveHostsInBatchRequest)
	in.HostIds = append(in.HostIds, fake_hostId1)
	in.HostIds = append(in.HostIds, fake_hostId2)
	out := new(clusterpb.RemoveHostsInBatchResponse)

	if err := resourceManager.RemoveHostsInBatch(context.TODO(), in, out); err != nil {
		t.Errorf("remove host in batch failed, err: %v\n", err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str {
		t.Errorf("Rsp not Expected, code: %d, msg: %s\n", out.Rs.Code, out.Rs.Message)
	}
}

func Test_RemoveHostsInBatch_WithErr(t *testing.T) {
	fake_str := "host already exists"
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "yyyy-yyyy-xxxx-xxxx"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHostsInBatch(func(ctx context.Context, in *dbpb.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostsInBatchResponse, error) {
		return nil, status.Errorf(codes.NotFound, fake_str)
	})

	in := new(clusterpb.RemoveHostsInBatchRequest)
	in.HostIds = append(in.HostIds, fake_hostId1)
	in.HostIds = append(in.HostIds, fake_hostId2)
	out := new(clusterpb.RemoveHostsInBatchResponse)

	if err := resourceManager.RemoveHostsInBatch(context.TODO(), in, out); err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() != codes.NotFound || st.Message() != fake_str {
				t.Errorf("Error has wrong type: code: %d, msg: %s", st.Code(), st.Message())
			}
		} else {
			t.Errorf("Error not status, err: %v\n", err)
		}
	} else {
		t.Errorf("Should Have a Error")
	}
}

func Test_RemoveHostsInBatch_WithErrCode(t *testing.T) {
	fake_str := "Host Ip is not Invalid"
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "yyyy-yyyy-xxxx-xxxx"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHostsInBatch(func(ctx context.Context, in *dbpb.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*dbpb.DBRemoveHostsInBatchResponse, error) {
		rsp := new(dbpb.DBRemoveHostsInBatchResponse)
		rsp.Rs = new(dbpb.DBHostResponseStatus)
		rsp.Rs.Code = int32(codes.InvalidArgument)
		rsp.Rs.Message = fake_str
		return rsp, nil
	})

	in := new(clusterpb.RemoveHostsInBatchRequest)
	in.HostIds = append(in.HostIds, fake_hostId1)
	in.HostIds = append(in.HostIds, fake_hostId2)
	out := new(clusterpb.RemoveHostsInBatchResponse)

	if err := resourceManager.RemoveHostsInBatch(context.TODO(), in, out); err != nil {
		t.Errorf("Should not Have a Error")
	} else {
		if out.Rs.Code != int32(codes.InvalidArgument) || out.Rs.Message != fake_str {
			t.Errorf("Error has wrong type: code: %d, msg: %s", out.Rs.Code, out.Rs.Message)
		}
	}
}

func Test_CheckDetails_Succeed(t *testing.T) {
	fake_str := "check details succeed"
	fake_hostId := "this-isxx-axxx-fake"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockCheckDetails(func(ctx context.Context, in *dbpb.DBCheckDetailsRequest, opts ...client.CallOption) (*dbpb.DBCheckDetailsResponse, error) {
		if in.HostId == fake_hostId {
			rsp := new(dbpb.DBCheckDetailsResponse)
			rsp.Rs = new(dbpb.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			rsp.Details = genHostRspFromDB(fake_hostId)
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(clusterpb.CheckDetailsRequest)
	in.HostId = fake_hostId
	out := new(clusterpb.CheckDetailsResponse)

	if err := resourceManager.CheckDetails(context.TODO(), in, out); err != nil {
		t.Errorf("check host details %s failed, err: %v\n", in.HostId, err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.Details.HostId != fake_hostId {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, hostId = %s\n", out.Rs.Code, out.Rs.Message, out.Details.HostId)
	}
}

func Test_ListHosts_Succeed(t *testing.T) {
	fake_str := "list hosts succeed"
	fake_hostId1 := "this-isxf-irst-fake"
	fake_hostId2 := "this-isse-cond-fake"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockListHost(func(ctx context.Context, in *dbpb.DBListHostsRequest, opts ...client.CallOption) (*dbpb.DBListHostsResponse, error) {
		if in.Page.PageSize == 2 {
			rsp := new(dbpb.DBListHostsResponse)
			rsp.Rs = new(dbpb.DBHostResponseStatus)
			rsp.Page = new(dbpb.DBHostPageDTO)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			rsp.HostList = append(rsp.HostList, genHostRspFromDB(fake_hostId1))
			rsp.HostList = append(rsp.HostList, genHostRspFromDB(fake_hostId2))
			rsp.Page.Total = 2
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(clusterpb.ListHostsRequest)
	in.PageReq = new(clusterpb.PageDTO)
	in.PageReq.PageSize = 2
	out := new(clusterpb.ListHostsResponse)

	if err := resourceManager.ListHost(context.TODO(), in, out); err != nil {
		t.Errorf("list hosts for pagesize %d failed, err: %v\n", in.PageReq.PageSize, err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.PageReq.Total != 2 || out.HostList[0].HostId != fake_hostId1 || out.HostList[1].HostId != fake_hostId2 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, total = %d, hostId1 = %s, hostId2 = %s\n", out.Rs.Code, out.Rs.Message, out.PageReq.Total, out.HostList[0].HostId, out.HostList[1].HostId)
	}
}

func Test_GetFailureDomain_Succeed(t *testing.T) {
	fake_str := "get failuredomain succeed"
	fake_name1 := "TEST_Zone1"
	fake_name2 := "TEST_Zone2"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockGetFailureDomain(func(ctx context.Context, in *dbpb.DBGetFailureDomainRequest, opts ...client.CallOption) (*dbpb.DBGetFailureDomainResponse, error) {
		if in.FailureDomainType == 2 {
			rsp := new(dbpb.DBGetFailureDomainResponse)
			rsp.Rs = new(dbpb.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			rsp.FdList = append(rsp.FdList, &dbpb.DBFailureDomainResource{
				FailureDomain: fake_name1,
				Spec:          "4u8c",
				Count:         2,
			})
			rsp.FdList = append(rsp.FdList, &dbpb.DBFailureDomainResource{
				FailureDomain: fake_name2,
				Spec:          "4u8c",
				Count:         3,
			})
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(clusterpb.GetFailureDomainRequest)
	in.FailureDomainType = 2
	out := new(clusterpb.GetFailureDomainResponse)

	if err := resourceManager.GetFailureDomain(context.TODO(), in, out); err != nil {
		t.Errorf("get failuredomains %d failed, err: %v\n", in.FailureDomainType, err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.FdList[0].FailureDomain != fake_name1 || out.FdList[1].Count != 3 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, out.FdList[0].FailureDomamin = %s, out.FdList[1].Count = %d\n", out.Rs.Code, out.Rs.Message, out.FdList[0].FailureDomain, out.FdList[1].Count)
	}
}

func Test_AllocHosts_Succeed(t *testing.T) {
	fake_str := "alloc hosts succeed"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAllocHosts(func(ctx context.Context, in *dbpb.DBAllocHostsRequest, opts ...client.CallOption) (*dbpb.DBAllocHostsResponse, error) {
		rsp := new(dbpb.DBAllocHostsResponse)
		rsp.Rs = new(dbpb.DBHostResponseStatus)
		if in.PdReq[0].FailureDomain == "Zone1" && in.TidbReq[0].Count == 1 && in.TikvReq[0].FailureDomain == "Zone2" {
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			rsp.PdHosts = append(rsp.PdHosts, &dbpb.DBAllocHostDTO{
				HostName: "TEST_HOST1",
				Ip:       "192.168.56.83",
			})
			rsp.PdHosts[0].Disk = &dbpb.DBDiskDTO{
				Name: "sdb",
				Path: "/mnt/pd",
			}
			rsp.TidbHosts = append(rsp.TidbHosts, &dbpb.DBAllocHostDTO{
				HostName: "TEST_HOST2",
				Ip:       "192.168.56.84",
			})
			rsp.TidbHosts[0].Disk = &dbpb.DBDiskDTO{
				Name: "sde",
				Path: "/mnt/tidb",
			}
			rsp.TikvHosts = append(rsp.TikvHosts, &dbpb.DBAllocHostDTO{
				HostName: "TEST_HOST3",
				Ip:       "192.168.56.85",
			})
			rsp.TikvHosts[0].Disk = &dbpb.DBDiskDTO{
				Name: "sdf",
				Path: "/mnt/tikv",
			}
			return rsp, nil
		} else {
			return nil, status.Errorf(codes.Internal, "BAD REQUEST in Alloc Hosts, pd zone %s, tidb count %d, tikv zone %s", in.PdReq[0].FailureDomain, in.TidbReq[0].Count, in.TikvReq[0].FailureDomain)
		}
	})

	in := new(clusterpb.AllocHostsRequest)
	in.PdReq = append(in.PdReq, &clusterpb.AllocationReq{
		FailureDomain: "Zone1",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	in.TidbReq = append(in.TidbReq, &clusterpb.AllocationReq{
		FailureDomain: "Zone1",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	in.TikvReq = append(in.TikvReq, &clusterpb.AllocationReq{
		FailureDomain: "Zone2",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	out := new(clusterpb.AllocHostResponse)

	if err := resourceManager.AllocHosts(context.TODO(), in, out); err != nil {
		t.Errorf("alloc hosts failed, err: %v\n", err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str ||
		out.PdHosts[0].HostName != "TEST_HOST1" || out.TidbHosts[0].HostName != "TEST_HOST2" || out.TikvHosts[0].HostName != "TEST_HOST3" ||
		out.PdHosts[0].Disk.Name != "sdb" || out.TidbHosts[0].Disk.Path != "/mnt/tidb" || out.TikvHosts[0].Disk.Name != "sdf" {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, out.PdHosts[0].HostName = %s, out.TidbHosts[0].HostName = %s, out.TikvHosts[0].HostName = %s\n",
			out.Rs.Code, out.Rs.Message, out.PdHosts[0].HostName, out.TidbHosts[0].HostName, out.TikvHosts[0].HostName)
	}
}

func Test_AllocResourcesInBatch_Succeed(t *testing.T) {

	fake_host_id1 := "TEST_host_id1"
	fake_host_ip1 := "199.199.199.1"
	fake_holder_id := "TEST_holder1"
	fake_request_id := "TEST_reqeust1"
	fake_disk_id1 := "TEST_disk_id1"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAllocResourcesInBatch(func(ctx context.Context, in *dbpb.DBBatchAllocRequest, opts ...client.CallOption) (*dbpb.DBBatchAllocResponse, error) {
		rsp := new(dbpb.DBBatchAllocResponse)
		rsp.Rs = new(dbpb.DBAllocResponseStatus)
		if in.BatchRequests[0].Applicant.HolderId == fake_holder_id && in.BatchRequests[1].Applicant.RequestId == fake_request_id &&
			in.BatchRequests[0].Requires[0].Count == 1 && in.BatchRequests[1].Requires[1].Require.PortReq[1].PortCnt == 2 {
			rsp.Rs.Code = int32(codes.OK)
		} else {
			return nil, status.Error(codes.Internal, "BAD alloc resource request")
		}
		var r dbpb.DBHostResource
		r.Reqseq = 0
		r.HostId = fake_host_id1
		r.HostIp = fake_host_ip1
		r.ComputeRes = new(dbpb.DBComputeRequirement)
		r.ComputeRes.CpuCores = in.BatchRequests[0].Requires[0].Require.ComputeReq.CpuCores
		r.ComputeRes.Memory = in.BatchRequests[0].Requires[0].Require.ComputeReq.CpuCores
		r.Location = new(dbpb.DBLocation)
		r.Location.Region = in.BatchRequests[0].Requires[0].Location.Region
		r.Location.Zone = in.BatchRequests[0].Requires[0].Location.Zone
		r.DiskRes = new(dbpb.DBDiskResource)
		r.DiskRes.DiskId = fake_disk_id1
		r.DiskRes.Type = in.BatchRequests[0].Requires[0].Require.DiskReq.DiskType
		r.DiskRes.Capacity = in.BatchRequests[0].Requires[0].Require.DiskReq.Capacity
		for _, portRes := range in.BatchRequests[0].Requires[0].Require.PortReq {
			var portResource dbpb.DBPortResource
			portResource.Start = portRes.Start
			portResource.End = portRes.End
			portResource.Ports = append(portResource.Ports, portRes.Start+1)
			portResource.Ports = append(portResource.Ports, portRes.Start+2)
			r.PortRes = append(r.PortRes, &portResource)
		}

		var one_rsp dbpb.DBAllocResponse
		one_rsp.Rs = new(dbpb.DBAllocResponseStatus)
		one_rsp.Rs.Code = int32(codes.OK)
		one_rsp.Results = append(one_rsp.Results, &r)

		rsp.BatchResults = append(rsp.BatchResults, &one_rsp)

		var two_rsp dbpb.DBAllocResponse
		two_rsp.Rs = new(dbpb.DBAllocResponseStatus)
		two_rsp.Rs.Code = int32(codes.OK)
		two_rsp.Results = append(two_rsp.Results, &r)
		two_rsp.Results = append(two_rsp.Results, &r)
		rsp.BatchResults = append(rsp.BatchResults, &two_rsp)
		return rsp, nil
	})

	in := new(clusterpb.BatchAllocRequest)

	var require clusterpb.AllocRequirement
	require.Location = new(clusterpb.Location)
	require.Location.Region = "TesT_Region1"
	require.Location.Zone = "TEST_Zone1"
	require.HostFilter = new(clusterpb.Filter)
	require.HostFilter.Arch = "X86"
	require.HostFilter.DiskType = "sata"
	require.HostFilter.Purpose = "General"
	require.Require = new(clusterpb.Requirement)
	require.Require.ComputeReq = new(clusterpb.ComputeRequirement)
	require.Require.ComputeReq.CpuCores = 4
	require.Require.ComputeReq.Memory = 8
	require.Require.DiskReq = new(clusterpb.DiskRequirement)
	require.Require.DiskReq.NeedDisk = true
	require.Require.DiskReq.DiskType = "sata"
	require.Require.DiskReq.Capacity = 256
	require.Require.PortReq = append(require.Require.PortReq, &clusterpb.PortRequirement{
		Start:   10000,
		End:     10010,
		PortCnt: 2,
	})
	require.Require.PortReq = append(require.Require.PortReq, &clusterpb.PortRequirement{
		Start:   10010,
		End:     10020,
		PortCnt: 2,
	})
	require.Count = 1

	var req1 clusterpb.AllocRequest
	req1.Applicant = new(clusterpb.Applicant)
	req1.Applicant.HolderId = fake_holder_id
	req1.Applicant.RequestId = fake_request_id
	req1.Requires = append(req1.Requires, &require)

	var req2 clusterpb.AllocRequest
	req2.Applicant = new(clusterpb.Applicant)
	req2.Applicant.HolderId = fake_holder_id
	req2.Applicant.RequestId = fake_request_id
	req2.Requires = append(req2.Requires, &require)
	req2.Requires = append(req2.Requires, &require)

	in.BatchRequests = append(in.BatchRequests, &req1)
	in.BatchRequests = append(in.BatchRequests, &req2)

	out := new(clusterpb.BatchAllocResponse)
	if err := resourceManager.AllocResourcesInBatch(context.TODO(), in, out); err != nil {
		t.Errorf("alloc resource failed, err: %v\n", err)
	}

	assert.True(t, out.Rs.Code == int32(codes.OK))
	assert.Equal(t, 2, len(out.BatchResults))
	assert.Equal(t, 1, len(out.BatchResults[0].Results))
	assert.Equal(t, 2, len(out.BatchResults[1].Results))
	assert.True(t, out.BatchResults[0].Results[0].DiskRes.DiskId == fake_disk_id1 && out.BatchResults[0].Results[0].DiskRes.Capacity == 256 && out.BatchResults[1].Results[0].DiskRes.Type == "sata")
	assert.True(t, out.BatchResults[1].Results[1].PortRes[0].Ports[0] == 10001 && out.BatchResults[1].Results[1].PortRes[0].Ports[1] == 10002)
	assert.True(t, out.BatchResults[1].Results[1].PortRes[1].Ports[0] == 10011 && out.BatchResults[1].Results[1].PortRes[1].Ports[1] == 10012)
}

func Test_RecycleResources_Succeed(t *testing.T) {
	fake_cluster_id := "TEST_Fake_CLUSTER_ID"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRecycleResources(func(ctx context.Context, in *dbpb.DBRecycleRequest, opts ...client.CallOption) (*dbpb.DBRecycleResponse, error) {
		rsp := new(dbpb.DBRecycleResponse)
		rsp.Rs = new(dbpb.DBAllocResponseStatus)
		if in.RecycleReqs[0].RecycleType == 2 && in.RecycleReqs[0].HolderId == fake_cluster_id {
			rsp.Rs.Code = int32(codes.OK)
		} else {
			return nil, status.Error(codes.Internal, "BAD recycle resource request")
		}
		return rsp, nil
	})

	var req clusterpb.RecycleRequest
	var require clusterpb.RecycleRequire
	require.HolderId = fake_cluster_id
	require.RecycleType = 2

	req.RecycleReqs = append(req.RecycleReqs, &require)
	out := new(clusterpb.RecycleResponse)
	if err := resourceManager.RecycleResources(context.TODO(), &req, out); err != nil {
		t.Errorf("recycle resource failed, err: %v\n", err)
	}

	assert.True(t, out.Rs.Code == int32(codes.OK))
}

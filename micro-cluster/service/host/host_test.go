package host

import (
	"context"
	"testing"

	"github.com/asim/go-micro/v3/client"
	"github.com/pingcap-inc/tiem/library/firstparty/config"
	hostPb "github.com/pingcap-inc/tiem/micro-cluster/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func initTestLog() {
	config.InitConfigForDev(config.MicroClusterMod)
	InitLoggerByKey(config.KEY_CLUSTER_LOG)
}

func genHostInfo(hostName string) *hostPb.HostInfo {
	host := hostPb.HostInfo{
		Ip:       "192.168.56.11",
		HostName: hostName,
		Os:       "Centos",
		Kernel:   "3.10",
		Dc:       "TEST_DC",
		Az:       "TEST_AZ",
		Rack:     "TEST_RACK",
		Status:   0,
		Nic:      "10GE",
		Purpose:  "Compute",
	}
	host.Disks = append(host.Disks, &hostPb.Disk{
		Name:     "sda",
		Path:     "/",
		Status:   0,
		Capacity: 512,
	})
	host.Disks = append(host.Disks, &hostPb.Disk{
		Name:     "sdb",
		Path:     "/mnt/sdb",
		Status:   0,
		Capacity: 1024,
	})
	return &host
}

func genHostRspFromDB(hostId string) *db.DBHostInfoDTO {
	host := db.DBHostInfoDTO{
		HostId:   hostId,
		HostName: "Test_DB2",
		Ip:       "192.168.56.11",
		Os:       "Centos",
		Kernel:   "3.10",
		Dc:       "TEST_DC",
		Az:       "TEST_AZ",
		Rack:     "TEST_RACK",
		Status:   0,
		Nic:      "10GE",
		Purpose:  "Compute",
	}
	host.Disks = append(host.Disks, &db.DBDiskDTO{
		Name:     "sda",
		Path:     "/",
		Status:   0,
		Capacity: 512,
	})
	host.Disks = append(host.Disks, &db.DBDiskDTO{
		Name:     "sdb",
		Path:     "/mnt/sdb",
		Status:   0,
		Capacity: 1024,
	})
	return &host
}

func Test_ImportHost_Succeed(t *testing.T) {
	initTestLog()
	fake_str := "import succeed"
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHost(func(ctx context.Context, in *db.DBAddHostRequest, opts ...client.CallOption) (*db.DBAddHostResponse, error) {
		if in.Host.HostName == "TEST_HOST1" {
			rsp := new(db.DBAddHostResponse)
			rsp.Rs = new(db.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			rsp.HostId = fake_hostId
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(hostPb.ImportHostRequest)
	in.Host = genHostInfo("TEST_HOST1")
	out := new(hostPb.ImportHostResponse)

	if err := ImportHost(context.TODO(), in, out); err != nil {
		t.Errorf("import host %s failed, err: %v\n", in.Host.HostName, err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.HostId != fake_hostId {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, hostId: %s\n", out.Rs.Code, out.Rs.Message, out.HostId)
	}
}

func Test_ImportHost_WithErr(t *testing.T) {
	initTestLog()
	fake_str := "host already exists"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHost(func(ctx context.Context, in *db.DBAddHostRequest, opts ...client.CallOption) (*db.DBAddHostResponse, error) {
		return nil, status.Errorf(codes.AlreadyExists, fake_str)
	})

	in := new(hostPb.ImportHostRequest)
	in.Host = genHostInfo("TEST_HOST1")
	out := new(hostPb.ImportHostResponse)

	if err := ImportHost(context.TODO(), in, out); err != nil {
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
	initTestLog()
	fake_str := "Host Ip is not Invalid"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHost(func(ctx context.Context, in *db.DBAddHostRequest, opts ...client.CallOption) (*db.DBAddHostResponse, error) {
		rsp := new(db.DBAddHostResponse)
		rsp.Rs = new(db.DBHostResponseStatus)
		rsp.Rs.Code = int32(codes.InvalidArgument)
		rsp.Rs.Message = fake_str
		return rsp, nil
	})

	in := new(hostPb.ImportHostRequest)
	in.Host = genHostInfo("TEST_HOST1")
	out := new(hostPb.ImportHostResponse)

	if err := ImportHost(context.TODO(), in, out); err != nil {
		t.Errorf("Should not Have a Error")
	} else {
		if out.Rs.Code != int32(codes.InvalidArgument) || out.Rs.Message != fake_str {
			t.Errorf("Error has wrong type: code: %d, msg: %s", out.Rs.Code, out.Rs.Message)
		}
	}
}

func Test_ImportHostsInBatch_Succeed(t *testing.T) {
	initTestLog()
	fake_str := "import succeed"
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "yyyy-yyyy-xxxx-xxxx"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHostsInBatch(func(ctx context.Context, in *db.DBAddHostsInBatchRequest, opts ...client.CallOption) (*db.DBAddHostsInBatchResponse, error) {
		if in.Hosts[0].HostName == "TEST_HOST1" && in.Hosts[1].HostName == "TEST_HOST2" {
			rsp := new(db.DBAddHostsInBatchResponse)
			rsp.Rs = new(db.DBHostResponseStatus)
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

	in := new(hostPb.ImportHostsInBatchRequest)
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST1"))
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST2"))
	out := new(hostPb.ImportHostsInBatchResponse)

	if err := ImportHostsInBatch(context.TODO(), in, out); err != nil {
		t.Errorf("import host in batch failed, err: %v\n", err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.HostIds[0] != fake_hostId1 || out.HostIds[1] != fake_hostId2 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, hostId0: %s, hostId1: %s\n", out.Rs.Code, out.Rs.Message, out.HostIds[0], out.HostIds[1])
	}
}

func Test_ImportHostsInBatch_WithErr(t *testing.T) {
	initTestLog()
	fake_str := "host already exists"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHostsInBatch(func(ctx context.Context, in *db.DBAddHostsInBatchRequest, opts ...client.CallOption) (*db.DBAddHostsInBatchResponse, error) {
		return nil, status.Errorf(codes.AlreadyExists, fake_str)
	})

	in := new(hostPb.ImportHostsInBatchRequest)
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST1"))
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST2"))
	out := new(hostPb.ImportHostsInBatchResponse)

	if err := ImportHostsInBatch(context.TODO(), in, out); err != nil {
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
	initTestLog()
	fake_str := "Host Ip is not Invalid"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockAddHostsInBatch(func(ctx context.Context, in *db.DBAddHostsInBatchRequest, opts ...client.CallOption) (*db.DBAddHostsInBatchResponse, error) {
		rsp := new(db.DBAddHostsInBatchResponse)
		rsp.Rs = new(db.DBHostResponseStatus)
		rsp.Rs.Code = int32(codes.InvalidArgument)
		rsp.Rs.Message = fake_str
		return rsp, nil
	})

	in := new(hostPb.ImportHostsInBatchRequest)
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST1"))
	in.Hosts = append(in.Hosts, genHostInfo("TEST_HOST2"))
	out := new(hostPb.ImportHostsInBatchResponse)

	if err := ImportHostsInBatch(context.TODO(), in, out); err != nil {
		t.Errorf("Should not Have a Error")
	} else {
		if out.Rs.Code != int32(codes.InvalidArgument) || out.Rs.Message != fake_str {
			t.Errorf("Error has wrong type: code: %d, msg: %s", out.Rs.Code, out.Rs.Message)
		}
	}
}

func Test_RemoveHost_Succeed(t *testing.T) {
	initTestLog()
	fake_str := "remove succeed"
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHost(func(ctx context.Context, in *db.DBRemoveHostRequest, opts ...client.CallOption) (*db.DBRemoveHostResponse, error) {
		if in.HostId == fake_hostId {
			rsp := new(db.DBRemoveHostResponse)
			rsp.Rs = new(db.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(hostPb.RemoveHostRequest)
	in.HostId = fake_hostId
	out := new(hostPb.RemoveHostResponse)

	if err := RemoveHost(context.TODO(), in, out); err != nil {
		t.Errorf("remove host %s failed, err: %v\n", in.HostId, err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str {
		t.Errorf("Rsp not Expected, code: %d, msg: %s\n", out.Rs.Code, out.Rs.Message)
	}
}

func Test_RemoveHost_WithErr(t *testing.T) {
	initTestLog()
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	fake_str := "host not exists"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHost(func(ctx context.Context, in *db.DBRemoveHostRequest, opts ...client.CallOption) (*db.DBRemoveHostResponse, error) {
		return nil, status.Errorf(codes.NotFound, fake_str)
	})

	in := new(hostPb.RemoveHostRequest)
	in.HostId = fake_hostId
	out := new(hostPb.RemoveHostResponse)

	if err := RemoveHost(context.TODO(), in, out); err != nil {
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
	initTestLog()
	fake_hostId := "xxxx-xxxx-yyyy-yyyy"
	fake_str := "Host Id is not Invalid"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHost(func(ctx context.Context, in *db.DBRemoveHostRequest, opts ...client.CallOption) (*db.DBRemoveHostResponse, error) {
		rsp := new(db.DBRemoveHostResponse)
		rsp.Rs = new(db.DBHostResponseStatus)
		rsp.Rs.Code = int32(codes.InvalidArgument)
		rsp.Rs.Message = fake_str
		return rsp, nil
	})

	in := new(hostPb.RemoveHostRequest)
	in.HostId = fake_hostId
	out := new(hostPb.RemoveHostResponse)

	if err := RemoveHost(context.TODO(), in, out); err != nil {
		t.Errorf("Should not Have a Error")
	} else {
		if out.Rs.Code != int32(codes.InvalidArgument) || out.Rs.Message != fake_str {
			t.Errorf("Error has wrong type: code: %d, msg: %s", out.Rs.Code, out.Rs.Message)
		}
	}
}
func Test_RemoveHostsInBatch_Succeed(t *testing.T) {
	initTestLog()
	fake_str := "remove in batch succeed"
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "yyyy-yyyy-xxxx-xxxx"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHostsInBatch(func(ctx context.Context, in *db.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*db.DBRemoveHostsInBatchResponse, error) {
		if in.HostIds[0] == fake_hostId1 && in.HostIds[1] == fake_hostId2 {
			rsp := new(db.DBRemoveHostsInBatchResponse)
			rsp.Rs = new(db.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(hostPb.RemoveHostsInBatchRequest)
	in.HostIds = append(in.HostIds, fake_hostId1)
	in.HostIds = append(in.HostIds, fake_hostId2)
	out := new(hostPb.RemoveHostsInBatchResponse)

	if err := RemoveHostsInBatch(context.TODO(), in, out); err != nil {
		t.Errorf("remove host in batch failed, err: %v\n", err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str {
		t.Errorf("Rsp not Expected, code: %d, msg: %s\n", out.Rs.Code, out.Rs.Message)
	}
}

func Test_RemoveHostsInBatch_WithErr(t *testing.T) {
	initTestLog()
	fake_str := "host already exists"
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "yyyy-yyyy-xxxx-xxxx"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHostsInBatch(func(ctx context.Context, in *db.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*db.DBRemoveHostsInBatchResponse, error) {
		return nil, status.Errorf(codes.NotFound, fake_str)
	})

	in := new(hostPb.RemoveHostsInBatchRequest)
	in.HostIds = append(in.HostIds, fake_hostId1)
	in.HostIds = append(in.HostIds, fake_hostId2)
	out := new(hostPb.RemoveHostsInBatchResponse)

	if err := RemoveHostsInBatch(context.TODO(), in, out); err != nil {
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
	initTestLog()
	fake_str := "Host Ip is not Invalid"
	fake_hostId1 := "xxxx-xxxx-yyyy-yyyy"
	fake_hostId2 := "yyyy-yyyy-xxxx-xxxx"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockRemoveHostsInBatch(func(ctx context.Context, in *db.DBRemoveHostsInBatchRequest, opts ...client.CallOption) (*db.DBRemoveHostsInBatchResponse, error) {
		rsp := new(db.DBRemoveHostsInBatchResponse)
		rsp.Rs = new(db.DBHostResponseStatus)
		rsp.Rs.Code = int32(codes.InvalidArgument)
		rsp.Rs.Message = fake_str
		return rsp, nil
	})

	in := new(hostPb.RemoveHostsInBatchRequest)
	in.HostIds = append(in.HostIds, fake_hostId1)
	in.HostIds = append(in.HostIds, fake_hostId2)
	out := new(hostPb.RemoveHostsInBatchResponse)

	if err := RemoveHostsInBatch(context.TODO(), in, out); err != nil {
		t.Errorf("Should not Have a Error")
	} else {
		if out.Rs.Code != int32(codes.InvalidArgument) || out.Rs.Message != fake_str {
			t.Errorf("Error has wrong type: code: %d, msg: %s", out.Rs.Code, out.Rs.Message)
		}
	}
}

func Test_CheckDetails_Succeed(t *testing.T) {
	initTestLog()
	fake_str := "check details succeed"
	fake_hostId := "this-isxx-axxx-fake"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockCheckDetails(func(ctx context.Context, in *db.DBCheckDetailsRequest, opts ...client.CallOption) (*db.DBCheckDetailsResponse, error) {
		if in.HostId == fake_hostId {
			rsp := new(db.DBCheckDetailsResponse)
			rsp.Rs = new(db.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			rsp.Details = genHostRspFromDB(fake_hostId)
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(hostPb.CheckDetailsRequest)
	in.HostId = fake_hostId
	out := new(hostPb.CheckDetailsResponse)

	if err := CheckDetails(context.TODO(), in, out); err != nil {
		t.Errorf("check host details %s failed, err: %v\n", in.HostId, err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.Details.HostId != fake_hostId {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, hostId = %s\n", out.Rs.Code, out.Rs.Message, out.Details.HostId)
	}
}

func Test_ListHosts_Succeed(t *testing.T) {
	initTestLog()
	fake_str := "list hosts succeed"
	fake_hostId1 := "this-isxf-irst-fake"
	fake_hostId2 := "this-isse-cond-fake"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockListHost(func(ctx context.Context, in *db.DBListHostsRequest, opts ...client.CallOption) (*db.DBListHostsResponse, error) {
		if in.Page.PageSize == 2 {
			rsp := new(db.DBListHostsResponse)
			rsp.Rs = new(db.DBHostResponseStatus)
			rsp.Page = new(db.DBHostPageDTO)
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

	in := new(hostPb.ListHostsRequest)
	in.PageReq = new(hostPb.PageDTO)
	in.PageReq.PageSize = 2
	out := new(hostPb.ListHostsResponse)

	if err := ListHost(context.TODO(), in, out); err != nil {
		t.Errorf("list hosts for pagesize %d failed, err: %v\n", in.PageReq.PageSize, err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.PageReq.Total != 2 || out.HostList[0].HostId != fake_hostId1 || out.HostList[1].HostId != fake_hostId2 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, total = %d, hostId1 = %s, hostId2 = %s\n", out.Rs.Code, out.Rs.Message, out.PageReq.Total, out.HostList[0].HostId, out.HostList[1].HostId)
	}
}

func Test_GetFailureDomain_Succeed(t *testing.T) {
	initTestLog()
	fake_str := "get failuredomain succeed"
	fake_name1 := "TEST_Zone1"
	fake_name2 := "TEST_Zone2"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockGetFailureDomain(func(ctx context.Context, in *db.DBGetFailureDomainRequest, opts ...client.CallOption) (*db.DBGetFailureDomainResponse, error) {
		if in.FailureDomainType == 2 {
			rsp := new(db.DBGetFailureDomainResponse)
			rsp.Rs = new(db.DBHostResponseStatus)
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
			rsp.FdList = append(rsp.FdList, &db.DBFailureDomainResource{
				FailureDomain: fake_name1,
				Spec:          "4u8c",
				Count:         2,
			})
			rsp.FdList = append(rsp.FdList, &db.DBFailureDomainResource{
				FailureDomain: fake_name2,
				Spec:          "4u8c",
				Count:         3,
			})
			return rsp, nil
		} else {
			return nil, status.Error(codes.Internal, "BAD REQUEST")
		}
	})

	in := new(hostPb.GetFailureDomainRequest)
	in.FailureDomainType = 2
	out := new(hostPb.GetFailureDomainResponse)

	if err := GetFailureDomain(context.TODO(), in, out); err != nil {
		t.Errorf("get failuredomains %d failed, err: %v\n", in.FailureDomainType, err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.FdList[0].FailureDomain != fake_name1 || out.FdList[1].Count != 3 {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, out.FdList[0].FailureDomamin = %s, out.FdList[1].Count = %d\n", out.Rs.Code, out.Rs.Message, out.FdList[0].FailureDomain, out.FdList[1].Count)
	}
}

func Test_AllocHosts_Succeed(t *testing.T) {
	initTestLog()
	fake_str := "alloc hosts succeed"
	fakeDBClient := InitMockDBClient()
	fakeDBClient.MockPreAllocHosts(func(ctx context.Context, in *db.DBPreAllocHostsRequest, opts ...client.CallOption) (*db.DBPreAllocHostsResponse, error) {
		rsp := new(db.DBPreAllocHostsResponse)
		rsp.Rs = new(db.DBHostResponseStatus)
		if in.Req.FailureDomain == "Zone1" && in.Req.Count == 2 {
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = "prealloc succeed"
			rsp.Results = append(rsp.Results, &db.DBPreAllocation{
				FailureDomain: in.Req.FailureDomain,
				HostName:      "TEST_HOST1",
				DiskName:      "sdb",
			})
			rsp.Results = append(rsp.Results, &db.DBPreAllocation{
				FailureDomain: in.Req.FailureDomain,
				HostName:      "TEST_HOST2",
				DiskName:      "sdb",
			})
			return rsp, nil
		} else if in.Req.FailureDomain == "Zone2" && in.Req.Count == 1 {
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = "prealloc succeed"
			rsp.Results = append(rsp.Results, &db.DBPreAllocation{
				FailureDomain: in.Req.FailureDomain,
				HostName:      "TEST_HOST3",
				DiskName:      "sdb",
			})
			return rsp, nil
		} else {
			return nil, status.Errorf(codes.Internal, "BAD REQUEST in PreAlloc, zone: %s, count %d", in.Req.FailureDomain, in.Req.Count)
		}
	})
	fakeDBClient.MockLockHosts(func(ctx context.Context, in *db.DBLockHostsRequest, opts ...client.CallOption) (*db.DBLockHostsResponse, error) {
		rsp := new(db.DBLockHostsResponse)
		rsp.Rs = new(db.DBHostResponseStatus)
		if len(in.Req) == 3 {
			rsp.Rs.Code = int32(codes.OK)
			rsp.Rs.Message = fake_str
		} else {
			return nil, status.Errorf(codes.Internal, "BAD REQUEST in LockHosts, total %d", len(in.Req))
		}
		return rsp, nil
	})

	in := new(hostPb.AllocHostsRequest)
	in.PdReq = append(in.PdReq, &hostPb.AllocationReq{
		FailureDomain: "Zone1",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	in.TidbReq = append(in.TidbReq, &hostPb.AllocationReq{
		FailureDomain: "Zone1",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	in.TikvReq = append(in.TikvReq, &hostPb.AllocationReq{
		FailureDomain: "Zone2",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	out := new(hostPb.AllocHostResponse)

	if err := AllocHosts(context.TODO(), in, out); err != nil {
		t.Errorf("alloc hosts failed, err: %v\n", err)
	}

	if out.Rs.Code != int32(codes.OK) || out.Rs.Message != fake_str || out.PdHosts[0].HostName != "TEST_HOST1" || out.TidbHosts[0].HostName != "TEST_HOST2" || out.TikvHosts[0].HostName != "TEST_HOST3" {
		t.Errorf("Rsp not Expected, code: %d, msg: %s, out.PdHosts[0].HostName = %s, out.TidbHosts[0].HostName = %s, out.TikvHosts[0].HostName = %s\n",
			out.Rs.Code, out.Rs.Message, out.PdHosts[0].HostName, out.TidbHosts[0].HostName, out.TikvHosts[0].HostName)
	}
}

package host

import (
	"context"
	"testing"

	"github.com/asim/go-micro/v3/client"
	"github.com/pingcap/tiem/library/firstparty/config"
	hostPb "github.com/pingcap/tiem/micro-cluster/proto"
	db "github.com/pingcap/tiem/micro-metadb/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func initTestLog() {
	config.InitForMonolith()
	InitLogger(config.KEY_CLUSTER_LOG)
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

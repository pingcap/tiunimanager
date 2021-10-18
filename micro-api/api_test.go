package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/asim/go-micro/v3/client"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/common/resource-type"

	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/resource/management"
	"github.com/pingcap-inc/tiem/micro-api/controller/resource/warehouse"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func performRequest(method, path, contentType string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Authorization", "Bearer fake-token")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("accept", "application/json")
	w := httptest.NewRecorder()

	// todo use httpClient to request
	g.ServeHTTP(w, req)
	return w
}

func Test_ListHosts_Succeed(t *testing.T) {
	fakeHostId1 := "fake-host-uuid-0001"
	fakeHostId2 := "fake-host-uuid-0002"
	fakeService := InitFakeClusterClient()
	fakeService.MockListHost(func(ctx context.Context, in *clusterpb.ListHostsRequest, opts ...client.CallOption) (*clusterpb.ListHostsResponse, error) {
		if in.Status != -1 {
			return nil, status.Errorf(codes.InvalidArgument, "file row count wrong")
		}
		rsp := new(clusterpb.ListHostsResponse)
		rsp.Rs = new(clusterpb.ResponseStatus)
		rsp.Rs.Code = int32(codes.OK)
		rsp.HostList = append(rsp.HostList, &clusterpb.HostInfo{
			HostId: fakeHostId1,
			Status: 2,
		})
		rsp.HostList = append(rsp.HostList, &clusterpb.HostInfo{
			HostId: fakeHostId2,
			Status: 2,
		})
		rsp.PageReq = new(clusterpb.PageDTO)
		rsp.PageReq.Page = 1
		rsp.PageReq.PageSize = 10
		rsp.PageReq.Total = 1
		return rsp, nil
	})

	w := performRequest("GET", "/api/v1/resources/hosts", "application/json", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, http.StatusOK, w.Code)

	type ResultWithPage struct {
		controller.ResultMark
		Data []management.HostInfo `json:"data"`
		Page controller.Page       `json:"page"`
	}
	var result ResultWithPage
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(t, err)

	value := len(result.Data)
	assert.Equal(t, 2, value)

	assert.Equal(t, result.Data[0].ID, fakeHostId1)
	assert.Equal(t, result.Data[1].ID, fakeHostId2)
	assert.True(t, result.Data[0].Status == 2)
	assert.True(t, result.Data[1].Status == 2)

}

func Test_ImportHost_Succeed(t *testing.T) {
	fakeHostId1 := "fake-host-uuid-0001"
	fakeHostIp := "l92.168.56.11"
	fakeService := InitFakeClusterClient()
	fakeService.MockImportHost(func(ctx context.Context, in *clusterpb.ImportHostRequest, opts ...client.CallOption) (*clusterpb.ImportHostResponse, error) {
		if in.Host.Ip != fakeHostIp || in.Host.Disks[0].Name != "nvme0p1" || in.Host.Disks[1].Path != "/mnt/disk2" {
			return nil, status.Errorf(codes.InvalidArgument, "import host info failed")
		}
		rsp := new(clusterpb.ImportHostResponse)
		rsp.Rs = new(clusterpb.ResponseStatus)
		rsp.Rs.Code = int32(codes.OK)
		rsp.HostId = fakeHostId1

		return rsp, nil
	})

	var hostInfo = []byte(`
	{
		"hostName": "TEST_HOST1",
		"ip": "l92.168.56.11",
		"dc": "TEST_DC",
		"az": "TEST_ZONE",
		"rack": "TEST_RACK",
		"disks": [
		  {
			"name": "nvme0p1",
			"path": "/mnt/disk1",
			"capacity": 256,
			"status": 1
		  },
		  {
			"name": "nvme0p2",
			"path": "/mnt/disk2",
			"capacity": 256,
			"status": 1
		  }
		]
	  }
	`)
	w := performRequest("POST", "/api/v1/resources/host", "application/json", bytes.NewBuffer(hostInfo))

	assert.Equal(t, http.StatusOK, w.Code)

	type CommonResult struct {
		controller.ResultMark
		Data management.ImportHostRsp `json:"data"`
	}
	var result CommonResult

	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(t, err)

	assert.Equal(t, result.Data.HostId, fakeHostId1)
}

func createBatchImportBody(filePath string) (string, io.Reader, error) {
	var err error
	buf := new(bytes.Buffer)
	bw := multipart.NewWriter(buf)

	f, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	_, filename := filepath.Split(filePath)
	fw, _ := bw.CreateFormFile("file", filename)
	io.Copy(fw, f)

	bw.Close()
	return bw.FormDataContentType(), buf, nil
}

func Test_ImportHostsInBatch_Succeed(t *testing.T) {
	fakeHostId1 := "fake-host-uuid-0001"
	fakeHostId2 := "fake-host-uuid-0002"
	fakeService := InitFakeClusterClient()
	fakeService.MockImportHostsInBatch(func(ctx context.Context, in *clusterpb.ImportHostsInBatchRequest, opts ...client.CallOption) (*clusterpb.ImportHostsInBatchResponse, error) {
		if len(in.Hosts) != 3 {
			return nil, status.Errorf(codes.InvalidArgument, "file row count wrong")
		}
		if in.Hosts[0].Ip != "192.168.56.11" || in.Hosts[1].Ip != "192.168.56.12" || in.Hosts[2].Ip != "192.168.56.13" {
			return nil, status.Errorf(codes.Internal, "Ip wrong")
		}
		if in.Hosts[0].Disks[0].Name != "vda" || in.Hosts[0].Disks[0].Type != string(resource.Sata) || in.Hosts[1].Disks[2].Path != "/mnt/path2" || in.Hosts[2].Disks[0].Type != string(resource.Sata) {
			return nil, status.Errorf(codes.Internal, "Disk wrong")
		}
		if in.Hosts[0].Reserved != false || in.Hosts[1].Reserved != false || in.Hosts[2].Reserved != false {
			return nil, status.Errorf(codes.Internal, "Reserved error")
		}
		if in.Hosts[0].Region != "Region1" || in.Hosts[1].Arch != "X86" {
			return nil, status.Errorf(codes.Internal, "Field error")
		}
		rsp := new(clusterpb.ImportHostsInBatchResponse)
		rsp.Rs = new(clusterpb.ResponseStatus)
		rsp.HostIds = append(rsp.HostIds, fakeHostId1)
		rsp.HostIds = append(rsp.HostIds, fakeHostId2)
		return rsp, nil
	})

	contentType, reader, err := createBatchImportBody("../etc/hostInfo_template.xlsx")
	if err != nil {
		t.Errorf("open template file failed\n")
	}

	w := performRequest("POST", "/api/v1/resources/hosts", contentType, reader)

	assert.Equal(t, http.StatusOK, w.Code)

	type CommonResult struct {
		controller.ResultMark
		Data management.ImportHostsRsp `json:"data"`
	}
	var result CommonResult

	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(t, err)

	value := len(result.Data.HostIds)
	assert.Equal(t, 2, value)

	assert.Equal(t, result.Data.HostIds[0], fakeHostId1)
	assert.Equal(t, result.Data.HostIds[1], fakeHostId2)
}

func Test_RemoveHostsInBatch_Succeed(t *testing.T) {
	fakeHostId1 := "fake-host-uuid-0001"
	fakeHostId2 := "fake-host-uuid-0002"
	fakeHostId3 := "fake-host-uuid-0003"
	fakeService := InitFakeClusterClient()
	fakeService.MockRemoveHostsInBatch(func(ctx context.Context, in *clusterpb.RemoveHostsInBatchRequest, opts ...client.CallOption) (*clusterpb.RemoveHostsInBatchResponse, error) {
		if in.HostIds[0] != fakeHostId1 || in.HostIds[1] != fakeHostId2 || in.HostIds[2] != fakeHostId3 {
			return nil, status.Errorf(codes.InvalidArgument, "input hostIds wrong")
		}
		rsp := new(clusterpb.RemoveHostsInBatchResponse)
		rsp.Rs = new(clusterpb.ResponseStatus)
		rsp.Rs.Code = int32(codes.OK)

		return rsp, nil
	})

	var hostIds = []byte(`
	[
		"fake-host-uuid-0001",
		"fake-host-uuid-0002",
		"fake-host-uuid-0003"
	]
	`)
	w := performRequest("DELETE", "/api/v1/resources/hosts", "application/json", bytes.NewBuffer(hostIds))

	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_RemoveHost_Succeed(t *testing.T) {
	fakeHostId1 := "fake-host-uuid-0001"
	fakeService := InitFakeClusterClient()
	fakeService.MockRemoveHost(func(ctx context.Context, in *clusterpb.RemoveHostRequest, opts ...client.CallOption) (*clusterpb.RemoveHostResponse, error) {
		if in.HostId != fakeHostId1 {
			return nil, status.Errorf(codes.InvalidArgument, "input hostIds wrong")
		}
		rsp := new(clusterpb.RemoveHostResponse)
		rsp.Rs = new(clusterpb.ResponseStatus)
		rsp.Rs.Code = int32(codes.OK)

		return rsp, nil
	})

	url := fmt.Sprintf("/api/v1/resources/hosts/%s", fakeHostId1)
	w := performRequest("DELETE", url, "application/json", nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_CheckDetails_Succeed(t *testing.T) {
	fakeHostId1 := "fake-host-uuid-0001"
	fakeHostName := "TEST_HOST1"
	fakeHostIp := "192.168.56.18"
	fakeDiskName := "sda"
	fakeDiskPath := "/"
	fakeService := InitFakeClusterClient()
	fakeService.MockCheckDetails(func(ctx context.Context, in *clusterpb.CheckDetailsRequest, opts ...client.CallOption) (*clusterpb.CheckDetailsResponse, error) {
		if in.HostId != fakeHostId1 {
			return nil, status.Errorf(codes.InvalidArgument, "input hostIds wrong")
		}
		rsp := new(clusterpb.CheckDetailsResponse)
		rsp.Rs = new(clusterpb.ResponseStatus)
		rsp.Rs.Code = int32(codes.OK)
		rsp.Details = &clusterpb.HostInfo{
			HostId:   fakeHostId1,
			HostName: fakeHostName,
			Ip:       fakeHostIp,
		}
		rsp.Details.Disks = append(rsp.Details.Disks, &clusterpb.Disk{
			Name:     fakeDiskName,
			Path:     fakeDiskPath,
			Capacity: 256,
			Status:   0,
		})
		return rsp, nil
	})

	url := fmt.Sprintf("/api/v1/resources/hosts/%s", fakeHostId1)
	w := performRequest("GET", url, "application/json", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	type CommonResult struct {
		controller.ResultMark
		Data management.HostDetailsRsp `json:"data"`
	}
	var result CommonResult

	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(t, err)
	assert.Equal(t, result.Data.Host.HostName, fakeHostName)
	assert.Equal(t, result.Data.Host.IP, fakeHostIp)
	assert.Equal(t, result.Data.Host.Disks[0].Name, fakeDiskName)
	assert.Equal(t, result.Data.Host.Disks[0].Path, fakeDiskPath)
}

func Test_DownloadTemplate_Succeed(t *testing.T) {
	common.TemplateFilePath = "../etc"
	w := performRequest("GET", "/api/v1/resources/hosts-template", "application/json", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, w.Header().Get("Content-Type"), "application/octet-stream")
	templateName := "hostInfo_template.xlsx"
	assert.Equal(t, w.Header().Get("Content-Disposition"), "attachment; filename="+templateName)
}

func Test_GetFailureDomain_Succeed(t *testing.T) {
	fakeZone1, fakeSpec1, fakeCount1, fakePurpose1 := "TEST_Zone1", "4u8g", 1, "Compute"
	fakeZone2, fakeSpec2, fakeCount2, fakePurpose2 := "TEST_Zone2", "8u16g", 2, "Storage"
	fakeZone3, fakeSpec3, fakeCount3, fakePurpose3 := "TEST_Zone3", "16u64g", 3, "Compute/Storage"
	fakeService := InitFakeClusterClient()
	fakeService.MockGetFailureDomain(func(ctx context.Context, in *clusterpb.GetFailureDomainRequest, opts ...client.CallOption) (*clusterpb.GetFailureDomainResponse, error) {
		if in.FailureDomainType != 2 {
			return nil, status.Errorf(codes.InvalidArgument, "input failuredomain type wrong")
		}
		rsp := new(clusterpb.GetFailureDomainResponse)
		rsp.Rs = new(clusterpb.ResponseStatus)
		rsp.Rs.Code = int32(codes.OK)
		rsp.FdList = append(rsp.FdList, &clusterpb.FailureDomainResource{
			FailureDomain: fakeZone1,
			Spec:          fakeSpec1,
			Count:         int32(fakeCount1),
			Purpose:       fakePurpose1,
		})
		rsp.FdList = append(rsp.FdList, &clusterpb.FailureDomainResource{
			FailureDomain: fakeZone2,
			Spec:          fakeSpec2,
			Count:         int32(fakeCount2),
			Purpose:       fakePurpose2,
		})
		rsp.FdList = append(rsp.FdList, &clusterpb.FailureDomainResource{
			FailureDomain: fakeZone3,
			Spec:          fakeSpec3,
			Count:         int32(fakeCount3),
			Purpose:       fakePurpose3,
		})
		return rsp, nil
	})

	w := performRequest("GET", "/api/v1/resources/failuredomains?failureDomainType=2", "application/json", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	type CommonResult struct {
		controller.ResultMark
		Data []warehouse.DomainResource `json:"data"`
	}
	var result CommonResult

	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(t, err)
	assert.Equal(t, result.Data[0].ZoneName, fakeZone1)
	assert.Equal(t, result.Data[1].SpecCode, fakeSpec2)
	assert.Equal(t, result.Data[2].Purpose, fakePurpose3)
}

func Test_AllocHosts_Succeed(t *testing.T) {
	fakeHostName1, fakeIp1, fakeDiskName1, fakeDiskPath1 := "TEST_HOST1", "192.168.56.11", "vdb", "/mnt/1"
	fakeHostName2, fakeIp2, fakeDiskName2, fakeDiskPath2 := "TEST_HOST2", "192.168.56.12", "sdb", "/mnt/disk2"
	fakeHostName3, fakeIp3, fakeDiskName3, fakeDiskPath3 := "TEST_HOST3", "192.168.56.13", "nvmep0", "/mnt/disk3"
	fakeHostName4, fakeIp4, fakeDiskName4, fakeDiskPath4 := "TEST_HOST4", "192.168.56.14", "sdc", "/mnt/disk4"
	fakeService := InitFakeClusterClient()
	fakeService.MockAllocHosts(func(ctx context.Context, in *clusterpb.AllocHostsRequest, opts ...client.CallOption) (*clusterpb.AllocHostResponse, error) {
		if in.PdReq[0].FailureDomain != "TEST_Zone1" || in.TidbReq[0].FailureDomain != "TEST_Zone2" || in.TikvReq[0].FailureDomain != "TEST_Zone3" || in.TikvReq[1].Memory != 64 {
			return nil, status.Errorf(codes.InvalidArgument, "input allocHosts type wrong, %s, %s, %s, %d",
				in.PdReq[0].FailureDomain, in.TidbReq[0].FailureDomain, in.TikvReq[0].FailureDomain, in.TikvReq[1].Memory)
		}
		rsp := new(clusterpb.AllocHostResponse)
		rsp.Rs = new(clusterpb.ResponseStatus)
		rsp.Rs.Code = int32(codes.OK)
		rsp.PdHosts = append(rsp.PdHosts, &clusterpb.AllocHost{
			HostName: fakeHostName1,
			Ip:       fakeIp1,
			UserName: "root",
			Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428", // "admin2"
		})
		rsp.PdHosts[0].Disk = &clusterpb.Disk{
			Name: fakeDiskName1,
			Path: fakeDiskPath1,
		}
		rsp.TidbHosts = append(rsp.TidbHosts, &clusterpb.AllocHost{
			HostName: fakeHostName2,
			Ip:       fakeIp2,
			UserName: "root",
			Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		})
		rsp.TidbHosts[0].Disk = &clusterpb.Disk{
			Name: fakeDiskName2,
			Path: fakeDiskPath2,
		}
		rsp.TikvHosts = append(rsp.TikvHosts, &clusterpb.AllocHost{
			HostName: fakeHostName3,
			Ip:       fakeIp3,
			UserName: "root",
			Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		})
		rsp.TikvHosts[0].Disk = &clusterpb.Disk{
			Name: fakeDiskName3,
			Path: fakeDiskPath3,
		}
		rsp.TikvHosts = append(rsp.TikvHosts, &clusterpb.AllocHost{
			HostName: fakeHostName4,
			Ip:       fakeIp4,
			UserName: "root",
			Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		})
		rsp.TikvHosts[1].Disk = &clusterpb.Disk{
			Name: fakeDiskName4,
			Path: fakeDiskPath4,
		}
		return rsp, nil
	})

	var allocReq = []byte(`
	{
		"pdReq": [
		  {
			"count": 1,
			"cpuCores": 4,
			"failureDomain": "TEST_Zone1",
			"memory": 8
		  }
		],
		"tidbReq": [
		  {
			"count": 1,
			"cpuCores": 8,
			"failureDomain": "TEST_Zone2",
			"memory": 16
		  }
		],
		"tikvReq": [
		  {
			"count": 1,
			"cpuCores": 8,
			"failureDomain": "TEST_Zone3",
			"memory": 16
		  },
		  {
			"count": 1,
			"cpuCores": 16,
			"failureDomain": "TEST_Zone3",
			"memory": 64
		  }
		]
	  }
	`)
	w := performRequest("POST", "/api/v1/resources/allochosts", "application/json", bytes.NewBuffer(allocReq))

	assert.Equal(t, http.StatusOK, w.Code)

	type CommonResult struct {
		controller.ResultMark
		Data management.AllocHostsRsp `json:"data"`
	}
	var result CommonResult

	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(t, err)
	assert.Equal(t, result.Data.PdHosts[0].HostName, fakeHostName1)
	assert.Equal(t, result.Data.PdHosts[0].Disk.Name, fakeDiskName1)
	assert.Equal(t, result.Data.TidbHosts[0].Ip, fakeIp2)
	assert.Equal(t, result.Data.TidbHosts[0].Disk.Path, fakeDiskPath2)
	assert.Equal(t, result.Data.TikvHosts[0].HostName, fakeHostName3)
	assert.Equal(t, result.Data.TikvHosts[1].HostName, fakeHostName4)
	assert.Equal(t, result.Data.TikvHosts[0].Disk.Name, fakeDiskName3)
	assert.Equal(t, result.Data.TikvHosts[1].Disk.Path, fakeDiskPath4)
	assert.Equal(t, result.Data.TikvHosts[1].Ip, fakeIp4)
	assert.Equal(t, result.Data.TidbHosts[0].UserName, "root")
	assert.Equal(t, result.Data.TikvHosts[0].Passwd, "admin2")
}

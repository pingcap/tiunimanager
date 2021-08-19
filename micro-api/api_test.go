package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/asim/go-micro/v3/client"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiem/micro-api/controller"
	"github.com/pingcap/tiem/micro-api/controller/hostapi"
	"github.com/pingcap/tiem/micro-api/route"
	managerPb "github.com/pingcap/tiem/micro-cluster/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func performRequest(g *gin.Engine, method, path, contentType string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Token", "fake-token")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("accept", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)
	return w
}

func TestListHosts(t *testing.T) {
	initConfig()
	fakeHostId1 := "fake-host-uuid-0001"
	fakeHostId2 := "fake-host-uuid-0002"
	fakeService := InitFakeManagerClient()
	fakeService.MockListHost(func(ctx context.Context, in *managerPb.ListHostsRequest, opts ...client.CallOption) (*managerPb.ListHostsResponse, error) {
		if in.Status != 0 {
			return nil, status.Errorf(codes.InvalidArgument, "file row count wrong")
		}
		rsp := new(managerPb.ListHostsResponse)
		rsp.Rs = new(managerPb.ResponseStatus)
		rsp.Rs.Code = int32(codes.OK)
		rsp.HostList = append(rsp.HostList, &managerPb.HostInfo{
			HostId: fakeHostId1,
			Status: 2,
		})
		rsp.HostList = append(rsp.HostList, &managerPb.HostInfo{
			HostId: fakeHostId2,
			Status: 2,
		})
		rsp.PageReq = new(managerPb.PageDTO)
		rsp.PageReq.Page = 1
		rsp.PageReq.PageSize = 10
		rsp.PageReq.Total = 1
		return rsp, nil
	})

	gin.SetMode(gin.ReleaseMode)
	g := gin.New()

	route.Route(g)
	//g.Group("/api/v1").Group("/").Use(nil)

	w := performRequest(g, "GET", "/api/v1/resources/hosts", "application/json", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	type ResultWithPage struct {
		controller.ResultMark
		Data []hostapi.HostInfo `json:"data"`
		Page controller.Page    `json:"page"`
	}
	var result ResultWithPage
	err := json.Unmarshal([]byte(w.Body.String()), &result)
	assert.Nil(t, err)

	value := len(result.Data)
	assert.Equal(t, 2, value)

	assert.Equal(t, result.Data[0].HostId, fakeHostId1)
	assert.Equal(t, result.Data[1].HostId, fakeHostId2)
	assert.True(t, result.Data[0].Status == 2)
	assert.True(t, result.Data[1].Status == 2)

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
	fakeService := InitFakeManagerClient()
	fakeService.MockImportHostsInBatch(func(ctx context.Context, in *managerPb.ImportHostsInBatchRequest, opts ...client.CallOption) (*managerPb.ImportHostsInBatchResponse, error) {
		if len(in.Hosts) != 2 {
			return nil, status.Errorf(codes.InvalidArgument, "file row count wrong")
		}
		if in.Hosts[0].Ip != "192.168.56.11" || in.Hosts[1].Ip != "192.168.56.12" {
			return nil, status.Errorf(codes.Internal, "Ip wrong")
		}
		if in.Hosts[0].Disks[0].Name != "vda" || in.Hosts[1].Disks[2].Path != "/mnt/path2" {
			return nil, status.Errorf(codes.Internal, "Disk wrong")
		}
		rsp := new(managerPb.ImportHostsInBatchResponse)
		rsp.Rs = new(managerPb.ResponseStatus)
		rsp.HostIds = append(rsp.HostIds, fakeHostId1)
		rsp.HostIds = append(rsp.HostIds, fakeHostId2)
		return rsp, nil
	})

	contentType, reader, err := createBatchImportBody("./hostInfo_template.xlsx")
	if err != nil {
		t.Errorf("open template file failed\n")
	}

	gin.SetMode(gin.ReleaseMode)
	g := gin.New()

	route.Route(g)
	//g.Group("/api/v1").Group("/").Use(nil)

	w := performRequest(g, "POST", "/api/v1/resources/hosts", contentType, reader)

	assert.Equal(t, http.StatusOK, w.Code)

	type CommonResult struct {
		controller.ResultMark
		Data hostapi.ImportHostsRsp `json:"data"`
	}
	var result CommonResult

	err = json.Unmarshal([]byte(w.Body.String()), &result)
	assert.Nil(t, err)

	value := len(result.Data.HostIds)
	assert.Equal(t, 2, value)

	assert.Equal(t, result.Data.HostIds[0], fakeHostId1)
	assert.Equal(t, result.Data.HostIds[1], fakeHostId2)
}

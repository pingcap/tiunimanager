package cdc

import (
	"context"
	"fmt"
	asserts "github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestChangeFeedServiceImpl_CreateChangeFeedTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil)
			r.ParseForm()
		} else if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			bytes := []byte("{\"checkpoint_tso\": 1, \"state\":\"normal\"}")
			w.Write(bytes)
		}
	}))
	defer server.Close()

	ipAndPort := strings.TrimPrefix(server.URL, "http://")
	host := strings.Split(ipAndPort, ":")[0]
	portStr := strings.Split(ipAndPort, ":")[1]
	port, _ := strconv.Atoi(portStr)

	cdcAddress := fmt.Sprintf("%s:%d", host, port)
	type args struct {
		ctx context.Context
		req ChangeFeedCreateReq
	}
	tests := []struct {
		name     string
		args     args
		wantResp ChangeFeedCmdAcceptResp
		wantErr  bool
	}{
		{
			name: "normal",
			args: args{
				ctx: context.TODO(),
				req: ChangeFeedCreateReq{CDCAddress: cdcAddress, IgnoreTxnStartTS: []uint64{}, FilterRules: []string{}, SinkConfig: []string{}},
			},
			wantResp: ChangeFeedCmdAcceptResp{
				Accepted: true,
				Succeed: true,
				ErrorCode: "",
				ErrorMsg: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResp, err := CDCService.CreateChangeFeedTask(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateChangeFeedTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("CreateChangeFeedTask() gotResp = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}

func TestChangeFeedServiceImpl_UpdateChangeFeedTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil)
			r.ParseForm()
		} else if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			bytes := []byte("{\"checkpoint_tso\": 1, \"state\":\"normal\"}")
			w.Write(bytes)
		}
	}))
	defer server.Close()

	ipAndPort := strings.TrimPrefix(server.URL, "http://")
	host := strings.Split(ipAndPort, ":")[0]
	portStr := strings.Split(ipAndPort, ":")[1]
	port, _ := strconv.Atoi(portStr)

	cdcAddress := fmt.Sprintf("%s:%d", host, port)
	t.Run("normal", func(t *testing.T) {
		gotResp, err := CDCService.UpdateChangeFeedTask(context.TODO(), ChangeFeedUpdateReq {CDCAddress: cdcAddress, IgnoreTxnStartTS: []uint64{}, FilterRules: []string{}, SinkConfig: []string{}})
		asserts.NoError(t, err)
		asserts.True(t, gotResp.Accepted)
		asserts.True(t, gotResp.Succeed)
	})
}

func TestChangeFeedServiceImpl_PauseChangeFeedTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil)
			r.ParseForm()
		} else if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			bytes := []byte("{\"checkpoint_tso\": 1, \"state\":\"stopped\"}")
			w.Write(bytes)
		}
	}))
	defer server.Close()

	ipAndPort := strings.TrimPrefix(server.URL, "http://")
	host := strings.Split(ipAndPort, ":")[0]
	portStr := strings.Split(ipAndPort, ":")[1]
	port, _ := strconv.Atoi(portStr)

	cdcAddress := fmt.Sprintf("%s:%d", host, port)
	t.Run("normal", func(t *testing.T) {
		gotResp, err := CDCService.PauseChangeFeedTask(context.TODO(), ChangeFeedPauseReq{
			CDCAddress: cdcAddress,
		})
		asserts.NoError(t, err)
		asserts.True(t, gotResp.Accepted)
		asserts.True(t, gotResp.Succeed)
	})
}

func TestChangeFeedServiceImpl_ResumeChangeFeedTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil)
			r.ParseForm()
		} else if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			bytes := []byte("{\"checkpoint_tso\": 1, \"state\":\"normal\"}")
			w.Write(bytes)
		}
	}))
	defer server.Close()

	ipAndPort := strings.TrimPrefix(server.URL, "http://")
	host := strings.Split(ipAndPort, ":")[0]
	portStr := strings.Split(ipAndPort, ":")[1]
	port, _ := strconv.Atoi(portStr)

	cdcAddress := fmt.Sprintf("%s:%d", host, port)
	t.Run("normal", func(t *testing.T) {
		gotResp, err := CDCService.ResumeChangeFeedTask(context.TODO(), ChangeFeedResumeReq{
			CDCAddress: cdcAddress,
		})
		asserts.NoError(t, err)
		asserts.True(t, gotResp.Accepted)
		asserts.True(t, gotResp.Succeed)
	})
}

func TestChangeFeedServiceImpl_DeleteChangeFeedTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil)
			r.ParseForm()
		}
	}))
	defer server.Close()

	ipAndPort := strings.TrimPrefix(server.URL, "http://")
	host := strings.Split(ipAndPort, ":")[0]
	portStr := strings.Split(ipAndPort, ":")[1]
	port, _ := strconv.Atoi(portStr)

	cdcAddress := fmt.Sprintf("%s:%d", host, port)
	t.Run("normal", func(t *testing.T) {
		gotResp, err := CDCService.DeleteChangeFeedTask(context.TODO(), ChangeFeedDeleteReq{
			CDCAddress: cdcAddress,
		})
		asserts.NoError(t, err)
		asserts.True(t, gotResp.Accepted)
	})
}

func TestChangeFeedServiceImpl_QueryChangeFeedTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil)
			r.ParseForm()
		} else if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			bytes := []byte("[{\"checkpoint_tso\": 1, \"state\":\"normal\"},{\"checkpoint_tso\": 1, \"state\":\"normal\"}]")
			w.Write(bytes)
		}
	}))
	defer server.Close()

	ipAndPort := strings.TrimPrefix(server.URL, "http://")
	host := strings.Split(ipAndPort, ":")[0]
	portStr := strings.Split(ipAndPort, ":")[1]
	port, _ := strconv.Atoi(portStr)

	cdcAddress := fmt.Sprintf("%s:%d", host, port)

	t.Run("normal", func(t *testing.T) {
		gotResp, err := CDCService.QueryChangeFeedTasks(context.TODO(), ChangeFeedQueryReq{
			CDCAddress: cdcAddress,
		})
		asserts.NoError(t, err)
		asserts.Equal(t, "normal", gotResp.Tasks[0].State)
		asserts.Equal(t, uint64(1), gotResp.Tasks[1].CheckPointTSO)
	})
}

func TestChangeFeedServiceImpl_DetailChangeFeedTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil)
			r.ParseForm()
		} else if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			bytes := []byte("{\"checkpoint_tso\": 1, \"state\":\"normal\"}")
			w.Write(bytes)
		}
	}))
	defer server.Close()

	ipAndPort := strings.TrimPrefix(server.URL, "http://")
	host := strings.Split(ipAndPort, ":")[0]
	portStr := strings.Split(ipAndPort, ":")[1]
	port, _ := strconv.Atoi(portStr)

	cdcAddress := fmt.Sprintf("%s:%d", host, port)

	t.Run("normal", func(t *testing.T) {
		gotResp, err := CDCService.DetailChangeFeedTask(context.TODO(), ChangeFeedDetailReq{
			CDCAddress: cdcAddress,
		})
		asserts.NoError(t, err)
		asserts.Equal(t, "normal", gotResp.State)
		asserts.Equal(t, uint64(1), gotResp.CheckPointTSO)
	})
}
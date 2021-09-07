package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/library/secondparty/libtiup"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiup/pkg/utils/rand"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type LoginRequest struct {
	Type     int    `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}

type ShareRequest struct {
	ExpireInSeconds int64 `json:"expire_in_sec"`
	RevokeWritePriv bool  `json:"revoke_write_priv"`
}

type ShareResponse struct {
	Code string `json:"code"`
}

type Dashboard struct {
	ClusterId string `json:"clusterId"`
	Url       string `json:"url"`
	Token 	  string `json:"token"`
}

var loginUrlSuffix string = "api/user/login"
var defaultExpire int64 = 60 * 60 * 3 //3 hour expire

func DescribeDashboard(ope *proto.OperatorDTO, clusterId string) (*Dashboard, error) {
	//todo: check operator and clusterId
	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil || clusterAggregation == nil || clusterAggregation.Cluster == nil {
		return nil, errors.New("load cluster aggregation")
	}

	/*
	url, err := getDashboardUrl(clusterAggregation)
	if err != nil {
		return nil, err
	}*/
	url := getDashboardUrlFromCluser(clusterAggregation)

	token, err := getLoginToken(url, "root", "") //todo: replace by real data
	if err != nil {
		return nil, err
	}

	dashboard := &Dashboard{
		ClusterId: clusterId,
		Url:       url,
		Token:     token,
	}

	return dashboard, nil
}

func getDashboardUrlFromCluser(clusterAggregation *ClusterAggregation) string {
	configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
	pdNum := len(configModel.PDServers)
	pdServer := configModel.PDServers[rand.Intn(pdNum)]
	pdClientPort := pdServer.ClientPort
	if pdClientPort == 0 {
		pdClientPort = DefaultPDClientPort
	}
	return fmt.Sprintf("http://%s:%d/dashboard/", pdServer.Host, pdClientPort)
}

func getDashboardUrl(clusterAggregation *ClusterAggregation) (string, error) {
	clusterName := clusterAggregation.Cluster.ClusterName
	getLogger().Infof("begin call tiupmgr: tiup cluster display %s --dashboard", clusterName)

	//tiup cluster display CLUSTER_NAME --dashboard
	resp := libtiup.MicroSrvTiupClusterDisplay(clusterName, 0, []string{"--dashboard"})
	if resp.ErrorStr != "" {
		getLogger().Errorf("call tiupmgr cluster display failed, %s", resp.ErrorStr)
		return "", errors.New(resp.ErrorStr)
	}
	getLogger().Infof("call tiupmgr success, resp: %v", resp)
	//DisplayRespString: "Dashboard URL: http://127.0.0.1:2379/dashboard/\n"
	result := strings.Split(strings.Replace(resp.DisplayRespString, "\n", "", -1), " ")
	return result[2], nil
}

func getLoginToken(dashboardUrl, userName, password string) (string, error) {
	url := fmt.Sprintf("%s%s", dashboardUrl, loginUrlSuffix)
	body := &LoginRequest{
		Username: userName,
		Password: password,
	}
	resp, err := post(url, body, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var loginResp LoginResponse
	err = json.Unmarshal(data, &loginResp)
	if err != nil {
		return "", err
	}
	getLogger().Infof("getLoginToken resp: %v", loginResp)

	return loginResp.Token, nil
}

func post(url string, body interface{}, headers map[string]string) (*http.Response, error) {
	//add post body
	var bodyJson []byte
	var req *http.Request
	if body != nil {
		var err error
		bodyJson, err = json.Marshal(body)
		if err != nil {
			return nil, errors.New("http post body to json failed")
		}
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyJson))
	if err != nil {
		return nil, errors.New("new request is fail: %v \n")
	}
	req.Header.Set("Content-type", "application/json")
	//add headers

	for key, val := range headers {
		req.Header.Add(key, val)
	}

	//http client
	client := &http.Client{}
	getLogger().Infof("%s URL : %s \n", http.MethodPost, req.URL.String())
	return client.Do(req)
}

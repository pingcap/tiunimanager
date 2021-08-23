package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiem/library/secondparty/libtiup"
	proto "github.com/pingcap/tiem/micro-cluster/proto"
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
	ShareCode string `json:"shareCode"`
}

var shareCodeUrlSuffix string = "api/user/share/code"
var loginUrlSuffix string = "api/user/login"
var defaultExpire int64 = 60 * 60 * 3 //3 hour expire

func DecribeDashboard(ope *proto.OperatorDTO, clusterId string) (*Dashboard, error) {
	//todo: check operator and clusterId
	url, err := getDashboardUrl(clusterId)
	if err != nil {
		return nil, err
	}

	token, err := getLoginToken(url, "root", "") //todo: replace by real data
	if err != nil {
		return nil, err
	}

	shareCode, err := generateShareCode(url, token)
	if err != nil {
		return nil, err
	}

	dashboard := &Dashboard{
		ClusterId: clusterId,
		Url:       url,
		ShareCode: shareCode,
	}

	return dashboard, nil
}

func getDashboardUrl(clusterId string) (string, error) {
	log.Infof("begin call tiupmgr: tiup cluster display %s --dashboard", clusterId)
	//tiup cluster display CLUSTER_NAME --dashboard
	resp := libtiup.MicroSrvTiupClusterDisplay(clusterId, 0, []string{"--dashboard"})
	if resp.Error != nil {
		log.Errorf("call tiupmgr cluster display failed, %s", resp.Error.Error())
		return "", resp.Error
	}
	log.Infof("call tiupmgr success, resp: %v", resp)
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
	log.Infof("getLoginToken resp: %v", loginResp)

	return loginResp.Token, nil
}

func generateShareCode(dashboardUrl, token string) (string, error) {
	url := fmt.Sprintf("%s%s", dashboardUrl, shareCodeUrlSuffix)
	body := &ShareRequest{
		ExpireInSeconds: defaultExpire,
		RevokeWritePriv: true,
	}
	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	resp, err := post(url, body, headers)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var shareResp ShareResponse
	err = json.Unmarshal(data, &shareResp)
	if err != nil {
		return "", err
	}
	log.Infof("generateShareCode resp: %v", shareResp)

	return shareResp.Code, nil
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
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//http client
	client := &http.Client{}
	log.Infof("%s URL : %s \n", http.MethodPost, req.URL.String())
	return client.Do(req)
}

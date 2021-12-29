/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: http_test.go
 * @Description: util http method
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/2 17:44
*******************************************************************************/

package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// UploadFile
// @Description: defines a file to be uploaded
type UploadFile struct {
	Name     string
	Filepath string
}

var httpClient = &http.Client{}

// Get
// @Description: implement HTTP GET
// @Parameter reqURL
// @Parameter reqParams
// @Parameter headers
// @return *http.Response
// @return error
func Get(reqURL string, reqParams map[string]string, headers map[string]string) (*http.Response, error) {
	urlParams := url.Values{}
	parsedURL, _ := url.Parse(reqURL)
	for key, val := range reqParams {
		urlParams.Set(key, val)
	}

	parsedURL.RawQuery = urlParams.Encode()
	urlPath := parsedURL.String()

	httpRequest, _ := http.NewRequest("GET", urlPath, nil)
	for k, v := range headers {
		httpRequest.Header.Add(k, v)
	}
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PostForm
// @Description: implements HTTP POST with form
// @Parameter reqURL
// @Parameter reqParams
// @Parameter headers
// @return *http.Response
// @return error
func PostForm(reqURL string, reqParams map[string]interface{}, headers map[string]string) (*http.Response, error) {
	return post(reqURL, reqParams, "application/x-www-form-urlencoded", nil, headers)
}

// PostJSON
// @Description: implements HTTP POST with JSON format
// @Parameter reqURL
// @Parameter reqParams
// @Parameter headers
// @return *http.Response
// @return error
func PostJSON(reqURL string, reqParams map[string]interface{}, headers map[string]string) (*http.Response, error) {
	return post(reqURL, reqParams, "application/json", nil, headers)
}


// PutJSON
// @Description: implements HTTP PUT with JSON format
// @Parameter reqURL
// @Parameter reqParams
// @Parameter headers
// @return *http.Response
// @return error
func PutJSON(reqURL string, reqParams map[string]interface{}, headers map[string]string) (*http.Response, error) {
	return postOrPut("PUT", reqURL, reqParams, "application/json", nil, headers)
}

// PostFile
// @Description: implements HTTP POST with file
// @Parameter reqURL
// @Parameter reqParams
// @Parameter files
// @Parameter headers
// @return *http.Response
// @return error
func PostFile(reqURL string, reqParams map[string]interface{}, files []UploadFile, headers map[string]string) (*http.Response, error) {
	return post(reqURL, reqParams, "multipart/form-data", files, headers)
}

// post
// @Description: common handle post request
// @Parameter reqURL
// @Parameter reqParams
// @Parameter contentType
// @Parameter files
// @Parameter headers
// @return *http.Response
// @return error
func post(reqURL string, reqParams map[string]interface{}, contentType string, files []UploadFile, headers map[string]string) (*http.Response, error) {
	return postOrPut("POST", reqURL, reqParams, contentType, files, headers)
}

func postOrPut(method string, reqURL string, reqParams map[string]interface{}, contentType string, files []UploadFile, headers map[string]string) (*http.Response, error) {
	if reqParams == nil {
		return nil, errors.New("the request parameter cannot be nil")
	}
	requestBody, realContentType, err := getReader(reqParams, contentType, files)
	if err != nil {
		return nil, err
	}
	httpRequest, _ := http.NewRequest(method, reqURL, requestBody)
	httpRequest.Header.Add("Content-Type", realContentType)
	for k, v := range headers {
		httpRequest.Header.Add(k, v)
	}
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// getReader
// @Description: get reader data
// @Parameter reqParams
// @Parameter contentType
// @Parameter files
// @return io.Reader
// @return string
// @return error
func getReader(reqParams map[string]interface{}, contentType string, files []UploadFile) (io.Reader, string, error) {
	if strings.Contains(contentType, "json") {
		bytesData, _ := json.Marshal(reqParams)
		return bytes.NewReader(bytesData), contentType, nil
	} else if files != nil {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for _, uploadFile := range files {
			file, err := os.Open(uploadFile.Filepath)
			if err != nil {
				return nil, "", err
			}
			defer file.Close()
			part, err := writer.CreateFormFile(uploadFile.Name, filepath.Base(uploadFile.Filepath))
			if err != nil {
				return nil, "", err
			}
			if _, err = io.Copy(part, file); err != nil {
				return nil, "", err
			}
		}
		for k, v := range reqParams {
			if err := writer.WriteField(k, fmt.Sprint(v)); err != nil {
				return nil, "", err
			}
		}
		if err := writer.Close(); err != nil {
			return nil, "", err
		}
		return body, writer.FormDataContentType(), nil
	} else {
		urlValues := url.Values{}
		for key, val := range reqParams {
			urlValues.Set(key, fmt.Sprint(val))
		}
		reqBody := urlValues.Encode()
		return strings.NewReader(reqBody), contentType, nil
	}
}

func Delete(reqURL string) (*http.Response, error) {
	httpRequest, _ := http.NewRequest("DELETE", reqURL, nil)
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
// Copyright 2021 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// UploadFile defines a file to be uploaded
type UploadFile struct {
	Name     string
	Filepath string
}

var httpClient = &http.Client{}

// Get implement HTTP GET
func Get(reqURL string, reqParams map[string]string, headers map[string]string) (int, string) {
	urlParams := url.Values{}
	parsedURL, _ := url.Parse(reqURL)
	for key, val := range reqParams {
		urlParams.Set(key, val)
	}

	parsedURL.RawQuery = urlParams.Encode()
	urlPath := parsedURL.String()

	httpRequest, _ := http.NewRequest("GET", urlPath, nil)
	if headers != nil {
		for k, v := range headers {
			httpRequest.Header.Add(k, v)
		}
	}
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	response, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(response)
}

func PUT(reqURL string, reqParams map[string]interface{}, headers map[string]string) (int, string) {
	requestBody, realContentType := getReader(reqParams, "application/json", nil)
	httpRequest, _ := http.NewRequest("PUT", reqURL, requestBody)
	httpRequest.Header.Add("Content-Type", realContentType)
	if headers != nil {
		for k, v := range headers {
			httpRequest.Header.Add(k, v)
		}
	}
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	response, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(response)
}

// PostForm implements HTTP POST with form
func PostForm(reqURL string, reqParams map[string]interface{}, headers map[string]string) (int, string) {
	return post(reqURL, reqParams, "application/x-www-form-urlencoded", nil, headers)
}

// PostJSON implements HTTP POST with JSON format
func PostJSON(reqURL string, reqParams map[string]interface{}, headers map[string]string) (int, string) {
	return post(reqURL, reqParams, "application/json", nil, headers)
}

// PostFile implements HTTP POST with file
func PostFile(reqURL string, reqParams map[string]interface{}, files []UploadFile, headers map[string]string) (int, string) {
	return post(reqURL, reqParams, "multipart/form-data", files, headers)
}

func post(reqURL string, reqParams map[string]interface{}, contentType string, files []UploadFile, headers map[string]string) (int, string) {
	requestBody, realContentType := getReader(reqParams, contentType, files)
	httpRequest, _ := http.NewRequest("POST", reqURL, requestBody)
	httpRequest.Header.Add("Content-Type", realContentType)
	if headers != nil {
		for k, v := range headers {
			httpRequest.Header.Add(k, v)
		}
	}
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	response, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(response)
}

func getReader(reqParams map[string]interface{}, contentType string, files []UploadFile) (io.Reader, string) {
	if strings.Index(contentType, "json") > -1 {
		bytesData, _ := json.Marshal(reqParams)
		return bytes.NewReader(bytesData), contentType
	} else if files != nil {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for _, uploadFile := range files {
			file, err := os.Open(uploadFile.Filepath)
			if err != nil {
				panic(err)
			}
			part, err := writer.CreateFormFile(uploadFile.Name, filepath.Base(uploadFile.Filepath))
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(part, file)
			file.Close()
		}
		for k, v := range reqParams {
			if err := writer.WriteField(k, v.(string)); err != nil {
				panic(err)
			}
		}
		if err := writer.Close(); err != nil {
			panic(err)
		}
		return body, writer.FormDataContentType()
	} else {
		urlValues := url.Values{}
		for key, val := range reqParams {
			urlValues.Set(key, val.(string))
		}
		reqBody := urlValues.Encode()
		return strings.NewReader(reqBody), contentType
	}
}

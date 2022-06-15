/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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

package framework

import (
	"bytes"
	"context"
	"crypto/tls"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/elastic/go-elasticsearch/v7"
)

type ElasticSearchClient struct {
	esClient *elasticsearch.Client
}

func InitElasticsearch(esAddress string) *ElasticSearchClient {
	addresses := parseAddress(esAddress)
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second * 3,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	return &ElasticSearchClient{esClient: esClient}
}

func (es *ElasticSearchClient) Search(index string, buf *bytes.Buffer, from, size int) (*esapi.Response, error) {
	if from <= 0 {
		from = 0
	}
	if size <= 0 {
		size = 10
	}
	// Perform the search request.
	resp, err := es.esClient.Search(
		es.esClient.Search.WithContext(context.Background()),
		es.esClient.Search.WithTimeout(time.Second*10),
		es.esClient.Search.WithIndex(index),
		es.esClient.Search.WithBody(buf),
		es.esClient.Search.WithTrackTotalHits(true),
		es.esClient.Search.WithFrom(from),
		es.esClient.Search.WithSize(size),
		es.esClient.Search.WithSort("@timestamp:desc"),
		es.esClient.Search.WithPretty(),
	)
	return resp, err
}

func parseAddress(esAddress string) []string {
	esAddressArr := strings.Split(esAddress, ",")
	addresses := make([]string, 0)
	for _, addr := range esAddressArr {
		if !strings.HasPrefix(addr, "http://") {
			addr = "http://" + addr
		}
		addresses = append(addresses, addr)
	}
	return addresses
}

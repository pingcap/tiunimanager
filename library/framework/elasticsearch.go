package framework

import (
	"bytes"
	"context"
	"crypto/tls"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/pingcap-inc/tiem/library/common"

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
				MinVersion: tls.VersionTLS11,
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
		if !strings.HasPrefix(addr, common.HttpProtocol) {
			addr = common.HttpProtocol + addr
		}
		addresses = append(addresses, addr)
	}
	return addresses
}

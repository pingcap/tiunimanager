package log

import (
	"time"
)

type ElasticSearchVo struct {
	Took     int      `json:"took"`
	TimedOut bool     `json:"timed_out"`
	Hits     HitsInfo `json:"hits"`
}

type HitsInfo struct {
	Total    HitsTotal    `json:"total"`
	MaxScore float64      `json:"max_score"`
	Hits     []HitsDetail `json:"hits"`
}

type HitsTotal struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

type HitsDetail struct {
	Index  string                 `json:"_index"`
	Id     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

type SearchTiDBLogSourceDetail struct {
	Tidb      map[string]interface{} `json:"tidb"`
	Log       LogDetail              `json:"log"`
	Ip        string                 `json:"ip"`
	ClusterId string                 `json:"clusterId"`
	Message   string                 `json:"message"`
	Fileset   FilesetDetail          `json:"fileset"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"@timestamp"`
}

type LogDetail struct {
	Offset int          `json:"offset"`
	Level  string       `json:"level"`
	Logger string       `json:"logger"`
	Origin OriginDetail `json:"origin"`
}

type OriginDetail struct {
	File FileDetail `json:"file"`
}

type FileDetail struct {
	Line string `json:"line"`
}

type FilesetDetail struct {
	Name string `json:"name"`
}

package logapi

type SearchTiDBLogReq struct {
	Module    string `form:"module"`
	Level     string `form:"level"`
	Ip        string `form:"ip"`
	Message   string `form:"message"`
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
	From      int    `form:"from"`
	Size      int    `form:"size"`
}

package logapi

type SearchTiDBLogReq struct {
	Module    string `form:"module" example:"tidb"`
	Level     string `form:"level" example:"warn"`
	Ip        string `form:"ip" example:"127.0.0.1"`
	Message   string `form:"message" example:"tidb log"`
	StartTime string `form:"startTime" example:"2021-09-01 12:00:00"`
	EndTime   string `form:"endTime" example:"2021-12-01 12:00:00"`
	From      int    `form:"from" example:"0"`
	Size      int    `form:"size" example:"10"`
}

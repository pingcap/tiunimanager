module github.com/pingcap-inc/tiem/tiup

go 1.16

replace github.com/appleboy/easyssh-proxy => github.com/AstroProfundis/easyssh-proxy v1.3.10-0.20210615044136-d52fc631316d

require (
	github.com/fatih/color v1.12.0
	github.com/joomcode/errorx v1.0.3
	github.com/pingcap/check v0.0.0-20200212061837-5e12011dc712
	github.com/pingcap/tiup v1.6.0-dev.0.20210819033350-8f4dc5dd94c8
	github.com/spf13/cobra v1.1.3
	go.uber.org/zap v1.17.0
)

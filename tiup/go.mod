module github.com/pingcap-inc/tiem/tiup

go 1.16

replace github.com/appleboy/easyssh-proxy => github.com/AstroProfundis/easyssh-proxy v1.3.10-0.20210615044136-d52fc631316d

require (
	github.com/creasty/defaults v1.5.2
	github.com/fatih/color v1.12.0
	github.com/joomcode/errorx v1.0.3
	github.com/pingcap/check v0.0.0-20200212061837-5e12011dc712
	github.com/pingcap/errors v0.11.4
	github.com/pingcap/tiup v1.6.0-dev.0.20210907031459-0284d6284320
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.0
	gopkg.in/yaml.v2 v2.4.0
)

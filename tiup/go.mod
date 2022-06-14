module github.com/pingcap/tiunimanager/tiup

go 1.16

replace github.com/appleboy/easyssh-proxy => github.com/AstroProfundis/easyssh-proxy v1.3.10-0.20210615044136-d52fc631316d

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/creasty/defaults v1.5.2
	github.com/fatih/color v1.13.0
	github.com/google/uuid v1.3.0
	github.com/joomcode/errorx v1.0.3
	github.com/pingcap/check v0.0.0-20200212061837-5e12011dc712
	github.com/pingcap/errors v0.11.4
	github.com/pingcap/tiup v1.6.0
	github.com/prometheus/common v0.30.0
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	go.etcd.io/etcd/client/pkg/v3 v3.5.0
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/mod v0.5.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

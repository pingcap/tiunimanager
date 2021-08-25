package main

import (
	"github.com/pingcap-inc/tiem/library/firstparty/framework"
)

func main() {
	f := framework.NewDefaultFramework(framework.MetaDBService,
		initSqliteDB,
		initTables,
		initTenantDataForDev,
		initResourceDataForDev,
	)

	f.StartService()
}

package framework

import (
	"github.com/asim/go-micro/v3"
	"github.com/micro/cli/v2"
)

func argsParse(argFlags []cli.Flag) {
	if len(argFlags) == 0 {
		return
	}
	srv := micro.NewService(
		micro.Flags(argFlags...),
	)
	srv.Init()
	srv = nil
}

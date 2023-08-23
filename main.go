package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/xmirrorsecurity/opensca-cli/opensca"
	"github.com/xmirrorsecurity/opensca-cli/opensca/format"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/util/config"
)

var version string

func v() {
	var v bool
	flag.BoolVar(&v, "version", false, "-version 打印版本信息")
	flag.Parse()
	if v {
		fmt.Println(version)
		os.Exit(0)
	}
}

func main() {
	v()
	config.ParseArgs()
	path := config.Conf().Path
	if len(path) > 0 {
		report := format.Report{
			ToolVersion: version,
			TaskResult:  opensca.RunTask(context.Background(), &model.TaskArg{DataOrigin: path}),
		}
		format.Save(report, "")
	} else {
		flag.PrintDefaults()
	}
}

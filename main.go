package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/cmd/format"
	"github.com/xmirrorsecurity/opensca-cli/opensca"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
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
	if len(path) == 0 {
		flag.PrintDefaults()
		return
	}

	start := time.Now()
	report := format.Report{
		ToolVersion: version,
		AppName:     path,
		StartTime:   start.Format("2006-01-02 15:04:05"),
	}

	if f, err := os.Stat(path); err == nil {
		report.Size = f.Size()
	}

	deps, err := opensca.RunTask(context.Background(), &opensca.TaskArg{DataOrigin: path})
	if err != nil {
		report.ErrorString = err.Error()
	}

	root := &model.DepGraph{}
	if len(deps) > 1 {
		for _, dep := range deps {
			root.AppendChild(dep)
		}
	} else if len(deps) == 1 {
		root = deps[0]
	}

	report.DepDetailGraph = detail.NewDepDetailGraph(root)
	err = detail.SearchDetail(report.DepDetailGraph)
	if err != nil {
		if report.ErrorString != "" {
			report.ErrorString += "\n"
		}
		report.ErrorString += err.Error()
	}

	report.CostTime = time.Since(start).Seconds()
	report.EndTime = time.Now().Format("2006-01-02 15:04:05")

	fmt.Println(format.Statis(report))
	format.Save(report, "")

}

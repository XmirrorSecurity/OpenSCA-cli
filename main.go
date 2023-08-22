package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/util/args"
	"github.com/xmirrorsecurity/opensca-cli/util/config"
	"github.com/xmirrorsecurity/opensca-cli/util/logs"
	"github.com/xmirrorsecurity/opensca-cli/util/model"
	"github.com/xmirrorsecurity/opensca-cli/util/report"

	"github.com/xmirrorsecurity/opensca-cli/analyzer/engine"
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
	if len(config.Conf().Path) > 0 {
		output(engine.NewEngine().ParseFile(config.Conf().Path))
	} else {
		flag.PrintDefaults()
	}
}

// output 输出结果
func output(depRoot *model.DepTree, taskInfo report.TaskInfo) {
	taskInfo.ToolVersion = version
	report.Format(depRoot)
	// 记录依赖
	logs.Debug("\n" + depRoot.String())
	// 输出结果
	var reportFunc func(*model.DepTree, report.TaskInfo)
	out := args.Config.Out
	ext := path.Ext(out)
	switch ext {
	case ".html":
		reportFunc = report.Html
	case ".json":
		if strings.HasSuffix(out, ".spdx.json") {
			reportFunc = report.SpdxJson
		} else if strings.HasSuffix(out, ".cdx.json") {
			reportFunc = report.CycloneDXJson
		} else if strings.HasSuffix(out, ".swid.json") {
			out += ".zip"
			reportFunc = report.SwidJson
		} else {
			reportFunc = report.Json
		}
	case ".spdx":
		reportFunc = report.Spdx
	case ".xml":
		if strings.HasSuffix(out, ".spdx.xml") {
			reportFunc = report.SpdxXml
		} else if strings.HasSuffix(out, ".cdx.xml") {
			reportFunc = report.CycloneDXXml
		} else if strings.HasSuffix(out, ".swid.xml") {
			out += ".zip"
			reportFunc = report.SwidXml
		} else {
			reportFunc = report.Xml
		}
	case ".csv":
		reportFunc = report.Csv
	case ".sqlite", ".db":
		reportFunc = report.Sqlite
	default:
		reportFunc = report.Json
	}

	fmt.Println(report.Statis(depRoot, taskInfo))
	reportFunc(depRoot, taskInfo)

}

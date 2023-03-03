/*
 * @Descripation: 引擎入口
 * @Date: 2021-11-03 11:12:19
 */
package main

import (
	"analyzer/engine"
	"flag"
	"fmt"
	"io"
	"path"
	"strings"
	"util/args"
	"util/logs"
	"util/model"
	"util/report"
)

var version string

func main() {
	args.Parse()
	if len(args.Config.Path) > 0 {
		output(engine.NewEngine().ParseFile(args.Config.Path))
	} else {
		flag.PrintDefaults()
	}
}

// output 输出结果
func output(depRoot *model.DepTree, taskInfo report.TaskInfo) {
	taskInfo.ToolVersion = version
	// 记录依赖
	logs.Debug("\n" + depRoot.String())
	// 输出结果
	var reportFunc func(*model.DepTree, report.TaskInfo) []byte
	var reportByWriterFunc func(io.Writer, *model.DepTree, report.TaskInfo)
	out := args.Config.Out
	switch path.Ext(out) {
	case ".html":
		reportFunc = report.Html
	case ".json":
		if strings.HasSuffix(out, ".spdx.json") {
			reportFunc = report.SpdxJson
		} else if strings.HasSuffix(out, ".cdx.json") {
			reportByWriterFunc = report.CycloneDXJson
		} else {
			reportFunc = report.Json
		}
	case ".spdx":
		reportFunc = report.Spdx
	case ".xml":
		if strings.HasSuffix(out, ".spdx.xml") {
			reportFunc = report.SpdxXml
		} else if strings.HasSuffix(out, ".cdx.xml") {
			reportByWriterFunc = report.CycloneDXXml
		} else {
			logs.Warn(fmt.Sprintf("not support report format: %s", args.Config.Out))
		}
	default:
		reportFunc = report.Json
	}
	fmt.Println(report.Statis(depRoot, taskInfo))
	if args.Config.Out != "" {
		if reportFunc != nil {
			report.Save(reportFunc(depRoot, taskInfo), args.Config.Out)
		} else if reportByWriterFunc != nil {
			report.SaveByWriter(func(w io.Writer) {
				reportByWriterFunc(w, depRoot, taskInfo)
			}, args.Config.Out)
		}
	} else {
		fmt.Println(string(reportFunc(depRoot, taskInfo)))
	}
}

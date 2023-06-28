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
	"io/fs"
	"os"
	"path"
	"path/filepath"
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
	report.Format(depRoot)
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
		} else if strings.HasSuffix(out, ".swid.json") {
			out += ".zip"
			reportByWriterFunc = report.SwidJson
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
		} else if strings.HasSuffix(out, ".swid.xml") {
			out += ".zip"
			reportByWriterFunc = report.SwidXml
		} else {
			reportFunc = report.Xml
		}
	default:
		reportFunc = report.Json
	}
	fmt.Println(report.Statis(depRoot, taskInfo))
	if out != "" {
		// 尝试创建导出文件目录
		if err := os.MkdirAll(filepath.Dir(out), fs.ModePerm); err != nil {
			logs.Warn(err)
			fmt.Println(err)
			return
		}
		if reportFunc != nil {
			report.Save(reportFunc(depRoot, taskInfo), out)
		} else if reportByWriterFunc != nil {
			report.SaveByWriter(func(w io.Writer) {
				reportByWriterFunc(w, depRoot, taskInfo)
			}, out)
		}
	} else {
		fmt.Println(string(reportFunc(depRoot, taskInfo)))
	}
}

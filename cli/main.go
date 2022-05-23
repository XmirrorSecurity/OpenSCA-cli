/*
 * @Descripation: 引擎入口
 * @Date: 2021-11-03 11:12:19
 */
package main

import (
	"analyzer/engine"
	"flag"
	"fmt"
	"path"
	"util/args"
	"util/logs"
	"util/model"
	"util/report"
)

func main() {
	args.Parse()
	if len(args.Filepath) > 0 {
		output(engine.NewEngine().ParseFile(args.Filepath))
	} else {
		flag.PrintDefaults()
	}
}

// output 输出结果
func output(depRoot *model.DepTree, taskInfo report.TaskInfo) {
	// 记录依赖
	logs.Debug("\n" + depRoot.String())
	// 输出结果
	var reportFunc func(*model.DepTree, report.TaskInfo) []byte
	switch path.Ext(args.Out) {
	case ".html":
		reportFunc = report.Html
	case ".json":
		reportFunc = report.Json
	default:
		reportFunc = report.Json
	}
	if args.Out != "" {
		report.Save(reportFunc(depRoot, taskInfo), args.Out)
	} else {
		fmt.Println(string(reportFunc(depRoot, taskInfo)))
	}
}

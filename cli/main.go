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
func output(depRoot *model.DepTree, err error) {
	// 整理错误信息
	errInfo := ""
	if err != nil {
		errInfo = err.Error()
	}
	// 记录依赖
	logs.Debug("\n" + depRoot.String())
	// 输出结果
	if args.Out != "" {
		switch path.Ext(args.Out) {
		case ".html":
			report.SaveHtml(depRoot, errInfo, args.Out)
		case ".json":
			report.SaveJson(depRoot, errInfo, args.Out)
		default:
			report.SaveJson(depRoot, errInfo, args.Out)
		}
	} else {
		fmt.Println(string(report.Json(depRoot, errInfo)))
	}
}
